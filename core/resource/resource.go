package resource

//go:generate mockery --name=Store -r --case underscore --with-expecter --structname ResourceStore --filename=resource_store.go --output=../mocks

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

type Store interface {
	GetByURN(ctx context.Context, urn string) (*Resource, error)
	List(ctx context.Context, filter map[string]string) ([]*Resource, error)

	Create(ctx context.Context, r Resource, hooks ...MutationHook) error
	Update(ctx context.Context, r Resource, hooks ...MutationHook) error
	Delete(ctx context.Context, urn string, hooks ...MutationHook) error

	DoPending(ctx context.Context, fn PendingHandler) error
}

// MutationHook values are passed to mutation operations of resource storage
// to handle any transactional requirements.
type MutationHook func(ctx context.Context) error

type PendingHandler func(ctx context.Context, res Resource) (*Resource, bool, error)

type Resource struct {
	URN       string            `json:"urn"`
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Project   string            `json:"project"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Spec      Spec              `json:"spec"`
	State     State             `json:"state"`
}

type Spec struct {
	Configs      json.RawMessage   `json:"configs"`
	Dependencies map[string]string `json:"dependencies"`
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
	parts := []string{"urn", "odpf", "entropy", res.Kind, res.Project, res.Name}
	return strings.Join(parts, urnSeparator)
}
