package kube

import (
	"context"
	"fmt"
	"github.com/mcuadros/go-defaults"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/pkg/errors"
)

const bufferSize = 4096
const sleepTime = 500

type Config struct {
	// Host - The hostname (in form of URI) of Kubernetes master.
	Host string
	// Insecure - Whether server should be accessed without verifying the TLS certificate.
	Insecure bool `default:"false"`
	// ClientCertificate - PEM-encoded client certificate for TLS authentication.
	ClientCertificate string
	// ClientKey - PEM-encoded client key for TLS authentication.
	ClientKey string
	// ClusterCACertificate - PEM-encoded root certificates bundle for TLS authentication.
	ClusterCACertificate string
	// Token - Token to authenticate a service account
	Token string
}

type Client struct {
	restConfig rest.Config
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}

func DefaultClientConfig() *Config {
	defaultProviderConfig := new(Config)
	defaults.SetDefaults(defaultProviderConfig)
	return defaultProviderConfig
}

func NewClient(config *Config) *Client {
	client := &Client{rest.Config{
		Host: config.Host,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(config.ClusterCACertificate),
			KeyData:  []byte(config.ClientKey),
			CertData: []byte(config.ClientCertificate),
		},
	}}

	if config.Token != "" {
		client.restConfig.BearerToken = config.Token
	}

	return client
}

func (c Client) StreamLogs(ctx context.Context, namespace string, filter map[string]string) (<-chan LogChunk, error) {
	var selectors []string
	var podName, containerName, labelSelector, filedSelector string
	var sinceSeconds, tailLines int64
	var opts metav1.ListOptions

	clientSet, err := kubernetes.NewForConfig(&c.restConfig)
	if err != nil {
		return nil, err
	}

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

	return streamFromPods(ctx, clientSet, namespace, containerName, opts, tailLines, sinceSeconds, filter)
}

func streamFromPods(ctx context.Context, clientSet *kubernetes.Clientset, namespace, containerName string, opts metav1.ListOptions, tailLines, sinceSeconds int64, filter map[string]string) (<-chan LogChunk, error) {
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, opts)
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
			go func(podName string, c v1.Container) {
				defer wg.Done()
				if err := streamContainerLogs(ctx, namespace, podName, logCh, clientSet, c, tailLines, sinceSeconds, filter); err != nil {
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

func streamContainerLogs(ctx context.Context, ns, podName string, logCh chan<- LogChunk, clientSet *kubernetes.Clientset, container v1.Container, tailLines, sinceSeconds int64, filter map[string]string) error {
	podLogOpts := v1.PodLogOptions{}
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
			if err == io.EOF || errors.Is(err, context.Canceled) {
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
