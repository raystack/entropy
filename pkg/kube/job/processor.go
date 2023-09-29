package job

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/batch/v1"

	"github.com/goto/entropy/pkg/errors"
)

const (
	Invalid StatusType = iota
	Success
	Failed
	Running
	Ready
	Finished
)

var deletionPolicy = metav1.DeletePropagationForeground

type (
	StatusType int
	Processor  struct {
		Job              *Job
		Client           v1.JobInterface
		watch            watch.Interface
		JobDeleteOptions metav1.DeleteOptions
	}
)

type Status struct {
	Status StatusType
	Err    error
}

func (jp *Processor) GetWatch() watch.Interface {
	return jp.watch
}

func NewProcessor(job *Job, client v1.JobInterface) *Processor {
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &deletionPolicy}
	return &Processor{Job: job, Client: client, JobDeleteOptions: deleteOptions}
}

func (jp *Processor) SubmitJob() error {
	_, err := jp.Client.Create(context.Background(), jp.Job.Template(), metav1.CreateOptions{})
	return err
}

func (jp *Processor) UpdateJob(suspend bool) error {
	job, err := jp.Client.Get(context.Background(), jp.Job.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	job.Spec.Suspend = &suspend
	_, err = jp.Client.Update(context.Background(), job, metav1.UpdateOptions{})
	return err
}

func (jp *Processor) CreateWatch() error {
	w, err := jp.Client.Watch(context.Background(), jp.Job.WatchOptions())
	jp.watch = w
	return err
}

func (jp *Processor) GetStatus() Status {
	job, err := jp.Client.Get(context.Background(), jp.Job.Name, metav1.GetOptions{})
	status := Status{}
	if err != nil {
		status.Status = Invalid
		status.Err = err
		return status
	}
	if *job.Status.Ready >= 1 {
		status.Status = Ready
		return status
	}
	if job.Status.Active >= 1 {
		status.Status = Running
		return status
	}
	if job.Status.Succeeded >= 1 {
		status.Status = Success
		return status
	} else {
		zap.L().Error(fmt.Sprintf("JOB FAILED %v\n", job))
		status.Status = Failed
		return status
	}
}

func (jp *Processor) DeleteJob() error {
	return jp.Client.Delete(context.Background(), jp.Job.Name, jp.JobDeleteOptions)
}

func (jp *Processor) WatchCompletion(exitChan chan Status) {
	if jp.GetWatch() == nil {
		exitChan <- Status{
			Status: Invalid,
			Err:    errors.New("watcher Object is not initialized"),
		}
		return
	}

	for {
		_, more := <-jp.GetWatch().ResultChan()
		if !more {
			break
		}
	}
	exitChan <- Status{Status: Finished}
}
