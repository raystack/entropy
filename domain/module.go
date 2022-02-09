package domain

type Module interface {
	ID() string
	Apply(r *Resource) (ResourceStatus, error)
}
