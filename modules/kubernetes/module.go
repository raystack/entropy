package kubernetes

import (
	_ "embed"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/pkg/errors"
)

var Module = module.Descriptor{
	Kind: "kubernetes",
	Actions: []module.ActionDesc{
		{
			Name: module.CreateAction,
		},
		{
			Name: module.UpdateAction,
		},
	},
	DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
		kd := &kubeDriver{}
		err := json.Unmarshal(conf, &kd)
		if err != nil {
			return nil, errors.ErrInvalid.WithMsgf("failed to unmarshal module config: %v", err)
		}
		return kd, nil
	},
}
