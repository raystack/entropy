package kube

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mcuadros/go-defaults"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/pkg/errors"
)

const (
	bufferSize                     = 4096
	sleepTime                      = 500
	defaultTTLSecondsAfterFinished = 60
)

type Client struct {
	restConfig      rest.Config
	streamingConfig rest.Config
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}

func DefaultClientConfig() Config {
	var defaultProviderConfig Config
	defaults.SetDefaults(&defaultProviderConfig)
	return defaultProviderConfig
}

func NewClient(config Config) *Client {
	return &Client{
		restConfig:      *config.RESTConfig(),
		streamingConfig: *config.StreamingConfig(),
	}
}

func (c Client) StreamLogs(ctx context.Context, namespace string, filter map[string]string) (<-chan LogChunk, error) {
	var selectors []string
	var podName, containerName, labelSelector, filedSelector string
	var sinceSeconds, tailLines int64
	var opts metav1.ListOptions

	for k, v := range filter {
		switch k {
		case "pod":
			podName = v
		case "container":
			containerName = v
		case "sinceSeconds":
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, errors.ErrInvalid.WithMsgf("invalid sinceSeconds filter value: %v", err)
			}
			sinceSeconds = i
		case "tailLine":
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, errors.ErrInvalid.WithMsgf("invalid tailLine filter value: %v", err)
			}
			tailLines = i
		default:
			s := fmt.Sprintf("%s=%s", k, v)
			selectors = append(selectors, s)
		}
	}

	if podName == "" {
		labelSelector = strings.Join(selectors, ",")
		opts = metav1.ListOptions{LabelSelector: labelSelector}
	} else {
		filedSelector = fmt.Sprintf("metadata.name=%s", podName)
		opts = metav1.ListOptions{FieldSelector: filedSelector}
	}

	return c.streamFromPods(ctx, namespace, containerName, opts, tailLines, sinceSeconds, filter)
}

func (c Client) RunJob(ctx context.Context, namespace, name string, image string, cmd []string, retries int32) error {
	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return err
	}

	jobs := clientSet.BatchV1().Jobs(namespace)

	var TTLSecondsAfterFinished int32 = defaultTTLSecondsAfterFinished

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,

							Command: cmd,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &retries,
		},
	}

	_, err = jobs.Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return waitForJob(ctx, name, jobs)
}

func waitForJob(ctx context.Context, jobName string, jobs typedbatchv1.JobInterface) error {
	for {
		job, err := jobs.Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return errors.ErrNotFound.WithMsgf("consumer group reset job not found")
		}

		// job hasn't started yet
		if job.Status.Active == 0 && job.Status.Succeeded == 0 && job.Status.Failed == 0 {
			continue
		}

		// job is still running
		if job.Status.Active > 0 {
			continue
		}

		// Job ran successfully
		if job.Status.Succeeded > 0 {
			return nil
		}

		return errors.ErrInternal.WithCausef("%s has failed with error", job.Name).WithCausef(job.Status.String())
	}
}

func (c Client) streamFromPods(ctx context.Context, namespace, containerName string, opts metav1.ListOptions, tailLines, sinceSeconds int64, filter map[string]string) (<-chan LogChunk, error) {
	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return nil, err
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	streamingClientSet, err := kubernetes.NewForConfig(&c.streamingConfig)
	if err != nil {
		return nil, err
	}

	logCh := make(chan LogChunk)

	wg := &sync.WaitGroup{}
	for _, pod := range pods.Items {
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
			if containerName != "" && container.Name != containerName {
				continue
			}
			wg.Add(1)
			go func(podName string, c corev1.Container) {
				defer wg.Done()
				if err := streamContainerLogs(ctx, namespace, podName, logCh, streamingClientSet, c, tailLines, sinceSeconds, filter); err != nil {
					log.Printf("[WARN] failed to stream from container '%s':%s", c.Name, err)
				}
			}(pod.Name, container)
		}
	}

	go func() {
		wg.Wait()
		close(logCh)
	}()

	return logCh, nil
}

func streamContainerLogs(ctx context.Context, ns, podName string, logCh chan<- LogChunk, clientSet *kubernetes.Clientset, container corev1.Container, tailLines, sinceSeconds int64, filter map[string]string) error {
	podLogOpts := corev1.PodLogOptions{}
	podLogOpts.Follow = true
	podLogOpts.Container = container.Name

	if sinceSeconds > 0 {
		podLogOpts.SinceSeconds = &sinceSeconds
	}

	if tailLines > 0 {
		podLogOpts.TailLines = &tailLines
	}

	podLogs, err := clientSet.CoreV1().Pods(ns).GetLogs(podName, &podLogOpts).Stream(ctx)
	if err != nil {
		return err
	}

	filter["pod"] = podName
	filter["container"] = container.Name

	buf := make([]byte, bufferSize)
	for {
		numBytes, err := podLogs.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		} else if numBytes == 0 {
			time.Sleep(sleepTime * time.Millisecond)
			continue
		}

		logChunk := LogChunk{
			Data:   []byte(string(buf[:numBytes])),
			Labels: filter,
		}

		select {
		case <-ctx.Done():
			return nil

		case logCh <- logChunk:
		}
	}
}
