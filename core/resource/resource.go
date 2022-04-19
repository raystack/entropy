package resource

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=./mocks

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusUnspecified Status = "STATUS_UNSPECIFIED"
	StatusPending     Status = "STATUS_PENDING"
	StatusError       Status = "STATUS_ERROR"
	StatusRunning     Status = "STATUS_RUNNING"
	StatusStopped     Status = "STATUS_STOPPED"
	StatusCompleted   Status = "STATUS_COMPLETED"
)

var (
	ErrResourceNotFound      = errors.New("no resource(s) found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
)

type Repository interface {
	Migrate() error

	GetByURN(urn string) (*Resource, error)
	List(filter map[string]string) ([]*Resource, error)
	Create(r Resource) error
	Update(r Resource) error
	Delete(urn string) error
}

type Resource struct {
	URN       string                 `bson:"urn"`
	Kind      string                 `bson:"kind"`
	Name      string                 `bson:"name"`
	Parent    string                 `bson:"parent"`
	Status    Status                 `bson:"status"`
	Labels    map[string]string      `bson:"labels"`
	Configs   map[string]interface{} `bson:"configs"`
	Providers []ProviderSelector     `bson:"providers"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

type Action struct {
	Name   string
	Params map[string]interface{}
}

type Status string

type Updates struct {
	Configs map[string]interface{}
}

type ProviderSelector struct {
	URN    string `bson:"urn"`
	Target string `bson:"target"`
}

func GenerateURN(res Resource) string {
	return strings.Join([]string{
		sanitizeString(res.Parent),
		sanitizeString(res.Name),
		sanitizeString(res.Kind),
	}, "-")
}

func sanitizeString(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
