package provider_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/provider/mocks"
	"github.com/odpf/entropy/pkg/errors"
)

var frozenTime = time.Unix(1650536955, 0)

func TestService_CreateProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupRepo func(t *testing.T) provider.Repository
		provider  provider.Provider
		want      *provider.Provider
		wantErr   error
	}{
		{
			name: "RepositoryConflict",
			setupRepo: func(t *testing.T) provider.Repository {
				repo := &mocks.ProviderRepository{}
				repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, p provider.Provider) {
						assert.Equal(t, "parent-child", p.URN)
					}).
					Return(errors.ErrConflict).
					Once()

				return repo
			},
			provider: provider.Provider{Parent: "parent", Name: "child", Kind: "bar"},
			wantErr:  errors.ErrConflict,
		},
		{
			name: "Successful",
			setupRepo: func(t *testing.T) provider.Repository {
				var storedProvider provider.Provider

				repo := &mocks.ProviderRepository{}
				repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, p provider.Provider) {
						assert.Equal(t, "parent-child", p.URN)
						storedProvider = p
					}).
					Return(nil).
					Once()

				repo.EXPECT().
					GetByURN(mock.Anything, "foo").
					Run(func(ctx context.Context, urn string) {
						assert.Equal(t, urn, "foo")
					}).
					Return(&storedProvider, nil).
					Once()

				return repo
			},
			provider: provider.Provider{Parent: "parent", Name: "child", Kind: "bar"},
			want: &provider.Provider{
				URN:       "parent-child",
				Kind:      "bar",
				Name:      "child",
				Parent:    "parent",
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := provider.NewService(tt.setupRepo(t), func() time.Time {
				return frozenTime
			})

			got, err := s.CreateProvider(context.Background(), tt.provider)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
