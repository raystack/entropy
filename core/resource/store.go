package resource

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=../mocks

import "context"

type Repository interface {
	GetByURN(ctx context.Context, urn string) (*Resource, error)
	List(ctx context.Context, filter map[string]string) ([]*Resource, error)
	Create(ctx context.Context, r Resource) error
	Update(ctx context.Context, r Resource) error
	Delete(ctx context.Context, urn string) error
}
