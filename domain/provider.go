package domain

import (
	"strings"
	"time"
)

type Provider struct {
	Urn       string                 `bson:"urn"`
	Name      string                 `bson:"name"`
	Kind      string                 `bson:"kind"`
	Parent    string                 `bson:"parent"`
	Configs   map[string]interface{} `bson:"configs"`
	Labels    map[string]string      `bson:"labels"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

func GenerateProviderUrn(res *Provider) string {
	return strings.Join([]string{
		sanitizeString(res.Parent),
		sanitizeString(res.Name),
	}, "-")
}
