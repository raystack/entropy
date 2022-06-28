# Modules

Module are responsible for achieving desired external system states based on a resource in Entropy.

This is how the module interface looks like:

```
type Module interface {
	Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error)

	Sync(ctx context.Context, spec Spec) (*resource.State, error)
}
```

Every Module has a `Plan` and a `Sync` method which plays it's part in the resource lifecycle.

Entropy currently support firehose and kubernetes modules, with more lined up.