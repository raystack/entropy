package resource

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=./mocks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

const resourceURNPrefix = "urn:odpf:entropy"

const (
	StatusUnspecified Status = "STATUS_UNSPECIFIED" // unknown
	StatusPending     Status = "STATUS_PENDING"     // intermediate
	StatusError       Status = "STATUS_ERROR"       // terminal
	StatusRunning     Status = "STATUS_DELETED"     // terminal
	StatusCompleted   Status = "STATUS_COMPLETED"   // terminal
)

type Repository interface {
	Migrate(ctx context.Context) error

	GetByURN(ctx context.Context, urn string) (*Resource, error)
	List(ctx context.Context, filter map[string]string) ([]*Resource, error)
	Create(ctx context.Context, r Resource) error
	Update(ctx context.Context, r Resource) error
	Delete(ctx context.Context, urn string) error
}

type Resource struct {
	URN       string            `bson:"urn"`
	Kind      string            `bson:"kind"`
	Name      string            `bson:"name"`
	Project   string            `bson:"project"`
	Labels    map[string]string `bson:"labels"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`

	Spec  Spec  `bson:"spec"`
	State State `bson:"state"`
}

type Output map[string]interface{}

type State struct {
	Status     Status          `bson:"status"`
	Output     Output          `bson:"output"`
	ModuleData json.RawMessage `bson:"module_data"`
}

type Spec struct {
	Configs      map[string]interface{} `bson:"configs"`
	Dependencies map[string]string      `bson:"dependencies"`
}

type Status string

type Action struct {
	Name   string
	Params map[string]interface{}
}

type Updates struct {
	Configs map[string]interface{}
}

type ProviderSelector struct {
	URN    string `bson:"urn"`
	Target string `bson:"target"`
}

func (res *Resource) Validate() error {
	res.Kind = strings.TrimSpace(res.Kind)
	res.Name = strings.TrimSpace(res.Name)
	res.Project = strings.TrimSpace(res.Project)

	if res.Kind == "" {
		return errors.ErrInvalid.WithMsgf("kind must be set")
	}
	if res.Name == "" {
		return errors.ErrInvalid.WithMsgf("name must be set")
	}
	if res.Project == "" {
		return errors.ErrInvalid.WithMsgf("project must be set")
	}

	if res.State.Status == "" {
		res.State.Status = StatusUnspecified
	}

	res.URN = generateURN(*res)
	return nil
}

func generateURN(res Resource) string {
	return strings.Join([]string{
		resourceURNPrefix,
		sanitizeString(res.Kind),
		sanitizeString(res.Project),
		sanitizeString(res.Name),
	}, ":")
}

func sanitizeString(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
