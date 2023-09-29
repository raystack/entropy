package job

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/job/config"
	"github.com/goto/entropy/modules/job/driver"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
	"github.com/goto/entropy/pkg/kube/job"
	"github.com/goto/entropy/pkg/validator"
)

var defaultDriverConf = config.DriverConf{
	Namespace: config.Default,
	RequestsAndLimits: map[string]config.RequestsAndLimits{
		config.Default: {
			Limits: config.UsageSpec{
				CPU:    "1",
				Memory: "2000Mi",
			},
			Requests: config.UsageSpec{
				CPU:    "1",
				Memory: "2000Mi",
			},
		},
	},
}

var Module = module.Descriptor{
	Kind: "job",
	Dependencies: map[string]string{
		driver.KeyKubeDependency: kubernetes.Module.Kind,
	},
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "Creates a new Kube job.",
		},
		{
			Name:        driver.SuspendAction,
			Description: "Suspend the kube Job.",
		},
		{
			Name:        driver.StartAction,
			Description: "Start the kube Job.",
		},
		{
			Name:        module.DeleteAction,
			Description: "Delete the kube Job.",
		},
	},
	DriverFactory: func(confJSON json.RawMessage) (module.Driver, error) {
		conf := defaultDriverConf
		if err := json.Unmarshal(confJSON, &conf); err != nil {
			return nil, err
		} else if err := validator.TaggedStruct(conf); err != nil {
			return nil, err
		}
		return &driver.Driver{
			Conf: conf,
			CreateJob: func(ctx context.Context, conf kube.Config, j *job.Job) error {
				kubeCl, err := kube.NewClient(ctx, conf)
				if err != nil {
					return errors.ErrInternal.WithMsgf("failed to create new kube client on job driver").WithCausef(err.Error())
				}
				processor, err := kubeCl.GetJobProcessor(j)
				if err != nil {
					return err
				}
				return processor.SubmitJob()
			},
			SuspendJob: func(ctx context.Context, conf kube.Config, j *job.Job) error {
				kubeCl, err := kube.NewClient(ctx, conf)
				if err != nil {
					return errors.ErrInternal.WithMsgf("failed to suspend the job").WithCausef(err.Error())
				}
				processor, err := kubeCl.GetJobProcessor(j)
				if err != nil {
					return err
				}
				return processor.UpdateJob(true)
			},
			DeleteJob: func(ctx context.Context, conf kube.Config, j *job.Job) error {
				kubeCl, err := kube.NewClient(ctx, conf)
				if err != nil {
					return errors.ErrInternal.WithMsgf("failed to delete the job").WithCausef(err.Error())
				}
				processor, err := kubeCl.GetJobProcessor(j)
				if err != nil {
					return err
				}
				return processor.DeleteJob()
			},
			StartJob: func(ctx context.Context, conf kube.Config, j *job.Job) error {
				kubeCl, err := kube.NewClient(ctx, conf)
				if err != nil {
					return errors.ErrInternal.WithMsgf("failed to start the job").WithCausef(err.Error())
				}
				processor, err := kubeCl.GetJobProcessor(j)
				if err != nil {
					return err
				}
				return processor.UpdateJob(false)
			},
		}, nil
	},
}
