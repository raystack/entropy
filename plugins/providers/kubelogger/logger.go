package kubelogger

import (
	"context"
	"flag"
	"io"
	"log"
	"path/filepath"
	"sync"

	"github.com/odpf/entropy/core/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetStreamingLogs(ctx context.Context, namespace string, selector string) (<-chan resource.LogChunk, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

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
				if err := streamContainerLogs(ctx, namespace, podName, logCh, clientSet, c); err != nil {
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

func streamContainerLogs(ctx context.Context, ns, podName string, logCh chan<- resource.LogChunk, clientSet *kubernetes.Clientset, container v1.Container) error {
	podLogOpts := v1.PodLogOptions{}
	podLogOpts.Follow = true
	podLogOpts.TailLines = &[]int64{100}[0]
	podLogOpts.Container = container.Name

	podLogs, err := clientSet.CoreV1().Pods(ns).GetLogs(podName, &podLogOpts).Stream(ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	for {
		numBytes, err := podLogs.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		logChunk := resource.LogChunk{
			Data: []byte(string(buf[:numBytes])),
			Labels: map[string]string{
				"pod":       podName,
				"container": container.Name,
			},
		}

		select {
		case <-ctx.Done():
			return nil

		case logCh <- logChunk:
		}
	}
}
