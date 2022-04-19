package kubelogger

import (
	"context"
	"flag"
	"io"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetStreamingLogs(ctx context.Context, namespace string, podName string) (io.ReadCloser, error) {
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

	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		podLogOpts := v1.PodLogOptions{}
		podLogOpts.Follow = true
		podLogOpts.TailLines = &[]int64{int64(100)}[0]
		podLogOpts.Container = container.Name
		podLogs, err := clientSet.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts).Stream(ctx)
		if err != nil {
			return nil, err
		}
		return podLogs, nil
	}
	return nil, nil
}
