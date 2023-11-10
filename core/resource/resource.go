package resource

//go:generate mockery --name=Store -r --case underscore --with-expecter --structname ResourceStore --filename=resource_store.go --output=../mocks

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/goto/entropy/pkg/errors"
)

const urnSeparator = ":"

var namingPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]+$`)
var namingPatternStartingWithDigits = regexp.MustCompile(`^\d*[A-Za-z0-9_-]+$`)

type Store interface {
	GetByURN(ctx context.Context, urn string) (*Resource, error)
	List(ctx context.Context, filter Filter, withSpecConfigs bool) ([]Resource, error)

	Create(ctx context.Context, r Resource, hooks ...MutationHook) error
	Update(ctx context.Context, r Resource, saveRevision bool, reason string, hooks ...MutationHook) error
	Delete(ctx context.Context, urn string, hooks ...MutationHook) error

	Revisions(ctx context.Context, selector RevisionsSelector) ([]Revision, error)

	SyncOne(ctx context.Context, syncFn SyncFn) error
}

type SyncFn func(ctx context.Context, res Resource) (*Resource, error)

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
	UpdatedBy string            `json:"updated_by"`
	CreatedBy string            `json:"created_by"`
	Spec      Spec              `json:"spec"`
	State     State             `json:"state"`
}

type Spec struct {
	Configs      json.RawMessage   `json:"configs"`
	Dependencies map[string]string `json:"dependencies"`
}

type Filter struct {
	Kind    string            `json:"kind"`
	Project string            `json:"project"`
	Labels  map[string]string `json:"labels"`
}

type UpdateRequest struct {
	Spec   Spec              `json:"spec"`
	Labels map[string]string `json:"labels"`
	UserID string
}

type RevisionsSelector struct {
	URN string `json:"urn"`
}

type Revision struct {
	ID        int64             `json:"id"`
	URN       string            `json:"urn"`
	Reason    string            `json:"reason"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	CreatedBy string            `json:"created_by"`

	Spec Spec `json:"spec"`
}

func (res *Resource) Validate(isCreate bool) error {
	res.Kind = strings.TrimSpace(res.Kind)
	res.Name = strings.TrimSpace(res.Name)
	res.Project = strings.TrimSpace(res.Project)

	if !namingPattern.MatchString(res.Kind) {
		return errors.ErrInvalid.WithMsgf("kind must match pattern '%s'", namingPattern)
	}
	if !namingPatternStartingWithDigits.MatchString(res.Name) {
		return errors.ErrInvalid.WithMsgf("name must match pattern '%s'", namingPatternStartingWithDigits)
	}
	if !namingPattern.MatchString(res.Project) {
		return errors.ErrInvalid.WithMsgf("project must match pattern '%s'", namingPattern)
	}

	if res.State.Status == "" {
		res.State.Status = StatusUnspecified
	}

	if isCreate {
		res.URN = GenerateURN(res.Kind, res.Project, res.Name)
	}
	return nil
}

func (f Filter) Apply(arr []Resource) []Resource {
	var res []Resource
	for _, r := range arr {
		if f.isMatch(r) {
			res = append(res, r)
		}
	}
	return res
}

func (f Filter) isMatch(r Resource) bool {
	kindMatch := f.Kind == "" || f.Kind == r.Kind
	projectMatch := f.Project == "" || f.Project == r.Project
	if !kindMatch || !projectMatch {
		return false
	}

	for k, v := range f.Labels {
		if r.Labels[k] != v {
			return false
		}
	}

	return true
}

// GenerateURN generates an Entropy URN address for the given combination.
// Note: Changing this will invalidate all existing resource identifiers.
func GenerateURN(kind, project, name string) string {
	parts := []string{"orn", "entropy", kind, project, name}
	return strings.Join(parts, urnSeparator)
}
