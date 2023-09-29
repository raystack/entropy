package job

import (
	"strings"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/goto/entropy/pkg/kube/pod"
)

const WatchTimeout int64 = 60

type Job struct {
	Pod         *pod.Pod
	Name        string
	Namespace   string
	Labels      map[string]string
	Parallelism *int32
	BackOffList *int32
	TTLSeconds  *int32
}

func (j *Job) Template() *v1.Job {
	return &v1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.Name,
			Labels:    j.Labels,
			Namespace: j.Namespace,
		},
		Spec: v1.JobSpec{
			Template:                j.Pod.Template(),
			Parallelism:             j.Parallelism,
			BackoffLimit:            j.BackOffList,
			TTLSecondsAfterFinished: j.TTLSeconds,
		},
		Status: v1.JobStatus{},
	}
}

func (j *Job) WatchOptions() metav1.ListOptions {
	timout := WatchTimeout
	label := strings.Join([]string{"name", j.Name}, "=")
	return metav1.ListOptions{TimeoutSeconds: &timout, LabelSelector: label}
}
