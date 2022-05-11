package resource

import (
	"regexp"
	"strings"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

const urnSeparator = ":"

var namingPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]+$`)

type Resource struct {
	URN       string            `json:"urn" bson:"urn"`
	Kind      string            `json:"kind" bson:"kind"`
	Name      string            `json:"name" bson:"name"`
	Project   string            `json:"project" bson:"project"`
	Labels    map[string]string `json:"labels" bson:"labels"`
	CreatedAt time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" bson:"updated_at"`
	Spec      Spec              `json:"spec" bson:"spec"`
	State     State             `json:"state" bson:"state"`
}

type Spec struct {
	Configs      map[string]interface{} `json:"configs" bson:"configs"`
	Dependencies map[string]string      `json:"dependencies" bson:"dependencies"`
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
