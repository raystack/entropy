package domain

import (
	"strings"
	"time"
)

type Resource struct {
	Urn       string                 `bson:"urn"`
	Name      string                 `bson:"name"`
	Parent    string                 `bson:"parent"`
	Kind      string                 `bson:"kind"`
	Configs   map[string]interface{} `bson:"configs"`
	Labels    map[string]string      `bson:"labels"`
	Status    string                 `bson:"status"`
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
	return strings.Replace(s, " ", "_", -1)
}
