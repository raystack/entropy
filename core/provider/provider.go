package provider

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ProviderRepository --filename=provider_repository.go --output=../../internal/mocks

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrProviderNotFound      = errors.New("no provider(s) found")
	ErrProviderAlreadyExists = errors.New("provider already exists")
)

type Repository interface {
	Migrate() error

	GetByURN(urn string) (*Provider, error)
	List(filter map[string]string) ([]*Provider, error)
	Create(r *Provider) error
}

type Provider struct {
	URN       string                 `bson:"urn"`
	Name      string                 `bson:"name"`
	Kind      string                 `bson:"kind"`
	Parent    string                 `bson:"parent"`
	Labels    map[string]string      `bson:"labels"`
	Configs   map[string]interface{} `bson:"configs"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

func GenerateURN(pro Provider) string {
	return strings.Join([]string{
		sanitizeString(pro.Parent),
		sanitizeString(pro.Name),
	}, "-")
}

func sanitizeString(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
