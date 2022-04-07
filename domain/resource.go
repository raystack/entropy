package domain

import (
	"strings"
	"time"
)

type ResourceStatus string

const (
	ResourceStatusUnspecified ResourceStatus = "STATUS_UNSPECIFIED"
	ResourceStatusPending     ResourceStatus = "STATUS_PENDING"
	ResourceStatusError       ResourceStatus = "STATUS_ERROR"
	ResourceStatusRunning     ResourceStatus = "STATUS_RUNNING"
	ResourceStatusStopped     ResourceStatus = "STATUS_STOPPED"
	ResourceStatusCompleted   ResourceStatus = "STATUS_COMPLETED"
)

type ProviderSelector struct {
	Urn    string `bson:"urn"`
	Target string `bson:"target"`
}

type Resource struct {
	Urn       string                 `bson:"urn"`
	Name      string                 `bson:"name"`
	Parent    string                 `bson:"parent"`
	Kind      string                 `bson:"kind"`
	Configs   map[string]interface{} `bson:"configs"`
	Labels    map[string]string      `bson:"labels"`
	Providers []ProviderSelector     `bson:"providers"`
	Status    ResourceStatus         `bson:"status"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

func GenerateResourceUrn(res *Resource) string {
	return strings.Join([]string{
		sanitizeString(res.Parent),
		sanitizeString(res.Name),
		sanitizeString(res.Kind),
	}, "-")
}

func sanitizeString(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
