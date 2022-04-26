package kubelogger

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func GetStreamingLogs(ctx context.Context, namespace string, filter map[string]string, cfg rest.Config) (<-chan resource.LogChunk, error) {
	var selectors []string
	var podName, containerName string

	clientSet, err := kubernetes.NewForConfig(&cfg)
	if err != nil {
		panic(err)
	}

	for k, v := range filter {
		switch k {
		case "pod":
			podName = v
		case "container":
			containerName = v
		default:
			s := fmt.Sprintf("%s=%s", k, v)
			selectors = append(selectors, s)
		}
	}

	if podName == "" {
		selector := strings.Join(selectors, ",")
		return streamFromAllPods(ctx, clientSet, namespace, selector, filter)
	} else {
		return streamFromOnePod(ctx, clientSet, namespace, podName, containerName, filter)
	}
}

func streamFromAllPods(ctx context.Context, clientSet *kubernetes.Clientset, namespace, selector string, filter map[string]string) (<-chan resource.LogChunk, error) {
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	logCh := make(chan resource.LogChunk)

	wg := &sync.WaitGroup{}
	for _, pod := range pods.Items {
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
			wg.Add(1)
			go func(podName string, c v1.Container) {
				defer wg.Done()
				if err := streamContainerLogs(ctx, namespace, podName, logCh, clientSet, c, filter); err != nil {
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

func streamFromOnePod(ctx context.Context, clientSet *kubernetes.Clientset, namespace, podName, containerName string, filter map[string]string) (<-chan resource.LogChunk, error) {
	pod, err := clientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	logCh := make(chan resource.LogChunk)

	wg := &sync.WaitGroup{}

	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		if containerName != "" && container.Name != containerName {
			continue
		}
		wg.Add(1)
		go func(podName string, c v1.Container) {
			defer wg.Done()
			if err := streamContainerLogs(ctx, namespace, podName, logCh, clientSet, c, filter); err != nil {
				log.Printf("[WARN] failed to stream from container '%s':%s", c.Name, err)
			}
		}(pod.Name, container)
	}

	go func() {
		wg.Wait()
		close(logCh)
	}()

	return logCh, nil
}

func streamContainerLogs(ctx context.Context, ns, podName string, logCh chan<- resource.LogChunk, clientSet *kubernetes.Clientset, container v1.Container, filter map[string]string) error {
	podLogOpts := v1.PodLogOptions{}
	podLogOpts.Follow = true
	podLogOpts.TailLines = &[]int64{100}[0]
	podLogOpts.Container = container.Name

	podLogs, err := clientSet.CoreV1().Pods(ns).GetLogs(podName, &podLogOpts).Stream(ctx)
	if err != nil {
		return err
	}

	filter["pod"] = podName
	filter["container"] = container.Name

	buf := make([]byte, 4096)
	for {
		numBytes, err := podLogs.Read(buf)
		if err != nil {
			if err == io.EOF || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		} else if numBytes == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		logChunk := resource.LogChunk{
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
