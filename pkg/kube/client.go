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
	"github.com/mitchellh/mapstructure"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	typedbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"

	"github.com/goto/entropy/pkg/errors"
)

const (
	bufferSize                     = 4096
	sleepTime                      = 500
	defaultTTLSecondsAfterFinished = 60
	trueString                     = "true"
)

var (
	ErrJobExecutionFailed = errors.ErrInternal.WithMsgf("job execution failed")
	ErrJobCreationFailed  = errors.ErrInternal.WithMsgf("job creation failed")
	ErrJobNotFound        = errors.ErrNotFound.WithMsgf("job not found")
)

type Client struct {
	restConfig      rest.Config
	streamingConfig rest.Config
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}

type Pod struct {
	Name       string   `json:"name"`
	Containers []string `json:"containers"`
}

type LogOptions struct {
	App          string `mapstructure:"app"`
	Pod          string `mapstructure:"pod"`
	Container    string `mapstructure:"container"`
	Follow       string `mapstructure:"follow"`
	Previous     string `mapstructure:"previous"`
	SinceSeconds string `mapstructure:"since_seconds"`
	Timestamps   string `mapstructure:"timestamps"`
	TailLines    string `mapstructure:"tail_lines"`
}

func (l LogOptions) getPodListOptions() (metav1.ListOptions, error) {
	labelSelector := labels.NewSelector()
	fieldSelector := fields.Everything()
	r, err := labels.NewRequirement("app", selection.Equals, []string{l.App})
	if err != nil {
		return metav1.ListOptions{}, err
	}
	labelSelector = labelSelector.Add(*r)

	if l.Pod != "" {
		fieldSelector = fields.AndSelectors(fieldSelector, fields.OneTermEqualSelector("metadata.name", l.Pod))
	}

	return metav1.ListOptions{
		LabelSelector: labelSelector.String(),
		FieldSelector: fieldSelector.String(),
	}, nil
}

func (l LogOptions) getPodLogOptions() (*corev1.PodLogOptions, error) {
	podLogOpts := &corev1.PodLogOptions{
		Container: l.Container,
	}

	if l.Follow == trueString {
		podLogOpts.Follow = true
	}

	if l.Previous == trueString {
		podLogOpts.Previous = true
	}

	if l.Timestamps == trueString {
		podLogOpts.Timestamps = true
	}

	if l.SinceSeconds != "" {
		ss, err := strconv.ParseInt(l.SinceSeconds, 10, 64)
		if err != nil {
			return nil, err
		}
		podLogOpts.SinceSeconds = &ss
	}

	if l.TailLines != "" {
		tl, err := strconv.ParseInt(l.TailLines, 10, 64)
		if err != nil {
			return nil, err
		}
		podLogOpts.TailLines = &tl
	}

	return podLogOpts, nil
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
	var logOptions LogOptions

	err := mapstructure.Decode(filter, &logOptions)
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf(err.Error())
	}

	return c.streamFromPods(ctx, namespace, logOptions)
}

func (c Client) RunJob(ctx context.Context, namespace, name string, image string, cmd []string, retries int32) error {
	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return ErrJobCreationFailed.WithCausef(err.Error())
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
		return ErrJobCreationFailed.WithCausef(err.Error())
	}

	return waitForJob(ctx, name, jobs)
}

func waitForJob(ctx context.Context, jobName string, jobs typedbatchv1.JobInterface) error {
	for {
		job, err := jobs.Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return ErrJobNotFound.WithCausef(err.Error())
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

		return ErrJobExecutionFailed.WithCausef(job.Status.String())
	}
}

func (c Client) streamFromPods(ctx context.Context, namespace string, logOptions LogOptions) (<-chan LogChunk, error) {
	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return nil, err
	}

	listOpts, err := logOptions.getPodListOptions()
	if err != nil {
		return nil, err
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, listOpts)
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
			if logOptions.Container != "" && container.Name != logOptions.Container {
				continue
			}
			plo, err := logOptions.getPodLogOptions()
			if err != nil {
				return nil, err
			}
			plo.Container = container.Name
			wg.Add(1)
			go func(podName string, plo corev1.PodLogOptions) {
				defer wg.Done()
				if err := streamContainerLogs(ctx, namespace, podName, logCh, streamingClientSet, plo); err != nil {
					log.Printf("[WARN] failed to stream from container '%s':%s", plo.Container, err)
				}
			}(pod.Name, *plo)
		}
	}

	go func() {
		wg.Wait()
		close(logCh)
	}()

	return logCh, nil
}

func (c Client) GetPodDetails(ctx context.Context, namespace string, labelSelectors map[string]string) ([]Pod, error) {
	var podDetails []Pod
	var selectors []string
	var labelSelector string
	var opts metav1.ListOptions

	for k, v := range labelSelectors {
		s := fmt.Sprintf("%s=%s", k, v)
		selectors = append(selectors, s)
	}
	labelSelector = strings.Join(selectors, ",")
	opts = metav1.ListOptions{LabelSelector: labelSelector}

	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return nil, err
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		// not listing pods that are not in running state or are about to terminate
		if pod.Status.Phase != corev1.PodRunning || pod.DeletionTimestamp != nil {
			continue
		}

		podDetail := Pod{
			Name: pod.Name,
		}

		for _, container := range pod.Spec.Containers {
			podDetail.Containers = append(podDetail.Containers, container.Name)
		}
		podDetails = append(podDetails, podDetail)
	}

	return podDetails, nil
}

func streamContainerLogs(ctx context.Context, ns, podName string, logCh chan<- LogChunk, clientSet *kubernetes.Clientset,
	podLogOpts corev1.PodLogOptions,
) error {
	podLogs, err := clientSet.CoreV1().Pods(ns).GetLogs(podName, &podLogOpts).Stream(ctx)
	if err != nil {
		return err
	}

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
			Labels: map[string]string{"pod": podName, "container": podLogOpts.Container},
		}

		select {
		case <-ctx.Done():
			return nil

		case logCh <- logChunk:
		}
	}
}
