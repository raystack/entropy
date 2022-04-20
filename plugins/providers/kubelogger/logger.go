package kubelogger

import (
	"context"
	"flag"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/odpf/entropy/core/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type LogChannel struct {
	LogChan  chan resource.LogChunk
	StopChan chan struct{}
}

func GetStreamingLogs(ctx context.Context, namespace string, podName string) (chan resource.LogChunk, error) {
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

	pod, err := clientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var logChannels []LogChannel

	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		podLogOpts := v1.PodLogOptions{}
		podLogOpts.Follow = true
		podLogOpts.TailLines = &[]int64{int64(100)}[0]
		podLogOpts.Container = container.Name
		podLogs, err := clientSet.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts).Stream(ctx)
		if err != nil {
			return nil, err
		}

		lc, sc := generate(podLogs, 100*time.Millisecond)
		l := new(LogChannel)
		l.LogChan = lc
		l.StopChan = sc

		logChannels = append(logChannels, *l)
	}

	logs, _ := multiplex(logChannels)
	//wg.Wait()

	return logs, nil
}

func generate(podLogs io.ReadCloser, interval time.Duration) (chan resource.LogChunk, chan struct{}) {
	logs := make(chan resource.LogChunk)
	sc := make(chan struct{})

	go func() {
		defer func() {
			close(sc)
		}()

		for {
			select {
			case <-sc:
				return
			default:
				time.Sleep(interval)

				buf := make([]byte, 10000)
				numBytes, err := podLogs.Read(buf)
				if numBytes == 0 {
					break
				}
				if err == io.EOF {
					break
				}

				if err != nil {
					break
				}
				logs <- resource.LogChunk{
					Data:   []byte(string(buf[:numBytes])),
					Labels: map[string]string{"resource": "SOME LABEL"},
				}
			}
		}
	}()

	return logs, sc
}

func stopGenerating(mc chan resource.LogChunk, sc chan struct{}) {
	sc <- struct{}{}
	<-sc

	close(mc)
}

func multiplex(mcs []LogChannel) (chan resource.LogChunk, *sync.WaitGroup) {
	mmc := make(chan resource.LogChunk, 1000)
	wg := &sync.WaitGroup{}

	for _, mc := range mcs {
		//wg.Add(1)

		go func(mc chan resource.LogChunk, wg *sync.WaitGroup) {
			//defer wg.Done()

			for m := range mc {
				mmc <- m
			}
		}(mc.LogChan, wg)
	}

	//defer stopAll(mcs)
	return mmc, wg
}

func stopAll(mcs []LogChannel) {
	for _, mc := range mcs {
		stopGenerating(mc.LogChan, mc.StopChan)
	}
}
