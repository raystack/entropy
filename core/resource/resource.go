package resource

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=./mocks

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

const urnSeparator = ":"

var namingPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]+$`)

const (
	StatusUnspecified Status = "STATUS_UNSPECIFIED" // unknown
	StatusPending     Status = "STATUS_PENDING"     // intermediate
	StatusError       Status = "STATUS_ERROR"       // terminal
	StatusDeleted     Status = "STATUS_DELETED"     // terminal
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

func (res *Resource) Validate() error {
	res.Kind = strings.TrimSpace(res.Kind)
	res.Name = strings.TrimSpace(res.Name)
	res.Project = strings.TrimSpace(res.Project)

	if !namingPattern.MatchString(res.Kind) {
		return errors.ErrInvalid.WithMsgf("kind must match pattern '%s'", namingPattern)
	}
	if !namingPattern.MatchString(res.Name) {
		return errors.ErrInvalid.WithMsgf("name must match pattern '%s'", namingPattern)
	}
	if !namingPattern.MatchString(res.Project) {
		return errors.ErrInvalid.WithMsgf("project must match pattern '%s'", namingPattern)
	}

	if res.State.Status == "" {
		res.State.Status = StatusUnspecified
	}

	res.URN = generateURN(*res)
	return nil
}

func generateURN(res Resource) string {
	var parts = []string{"urn", "odpf", "entropy", res.Kind, res.Project, res.Name}
	return strings.Join(parts, urnSeparator)
}
