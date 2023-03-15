package kubernetes

import (
	_ "embed"
	"encoding/json"

	"github.com/goto/entropy/core/module"
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
		return &kubeDriver{}, nil
	},
}
