# How to write an Entropy Module

This guide will take you through the important points that must be take care while writing an Entropy Module.

## Module Descriptor
Every module shall export a `Module Descriptor`, which represents the supported actions, resource-kind the module can operate on, etc.

This is how a module Descriptor shall look like:

```
type Descriptor struct {
	Kind         string            `json:"kind"`
	Actions      []ActionDesc      `json:"actions"`
	Dependencies map[string]string `json:"dependencies"`
	Module       Module            `json:"-"`
}
```

For instance, this is how kubernetes descriptor looks like:

```
var Module = module.Descriptor{
	Kind: "kubernetes",
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			ParamSchema: configSchema,
		},
		{
			Name:        module.UpdateAction,
			ParamSchema: configSchema,
		},
	},
	Module: &kubeModule{},
}

type kubeModule struct{}
```

## Module Interface

The `Module` description must follow the Module Interface

```
type Module interface {
	Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error)

	Sync(ctx context.Context, spec Spec) (*resource.State, error)
}
```

Every Module must have a `Plan` and a `Sync` method with the signatures shown above.

## Loggable Modules

If you want to support streaming of logs from your module, just add a `Log` function to your module. Entropy will check if the module is loggable against the Loggable interface.

```
type Loggable interface {
	Module

	Log(ctx context.Context, spec Spec, filter map[string]string) (<-chan LogChunk, error)
}
```

Note: You may follow through the codebase to have a look at the Spec, ActionDesc, LogChunk etc interfaces.

## Important points to note

This is how the Resource.State looks like:

```
type State struct {
	Status     string          `json:"status"`
	Output     json.RawMessage `json:"output"`
	ModuleData json.RawMessage `json:"module_data,omitempty"`
}
```

***Some highlights:***

- ModuleData field to store all the internal state information of a module. This shall be used by the module to make actions.
- Output field is for the outer world. This can be used by the frontend to give feedbacks to the user.

Reader shall also go through the resource-life-cycle to have a look how a module affects a resource in the Plan & Sync phases.