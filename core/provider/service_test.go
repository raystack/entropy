package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/provider/mocks"
)

func TestService_CreateProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupRepo func(t *testing.T) provider.Repository
		provider  provider.Provider
		want      *provider.Provider
		wantErr   bool
	}{
		{
			name: "CreateError_Repository",
			setupRepo: func(t *testing.T) provider.Repository {
				repo := &mocks.ProviderRepository{}
				repo.EXPECT().
					Create(mock.Anything).
					Run(func(p provider.Provider) {
						assert.Equal(t, p.URN, "foo")
					}).
					Return(errors.New("failed")).
					Once()

				return repo
			},
			provider: provider.Provider{URN: "foo"},
			wantErr:  true,
		},
		{
			name: "GetError_Repository",
			setupRepo: func(t *testing.T) provider.Repository {
				repo := &mocks.ProviderRepository{}
				repo.EXPECT().
					Create(mock.Anything).
					Run(func(p provider.Provider) {
						assert.Equal(t, p.URN, "foo")
					}).
					Return(nil).
					Once()

				repo.EXPECT().
					GetByURN("foo").
					Run(func(urn string) {
						assert.Equal(t, urn, "foo")
					}).
					Return(nil, errors.New("failed")).
					Once()

				return repo
			},
			provider: provider.Provider{URN: "foo"},
			wantErr:  true,
		},
		{
			name: "Successful",
			setupRepo: func(t *testing.T) provider.Repository {
				var storedProvider provider.Provider

				repo := &mocks.ProviderRepository{}
				repo.EXPECT().
					Create(mock.Anything).
					Run(func(p provider.Provider) {
						assert.Equal(t, p.URN, "foo")
						storedProvider = p
					}).
					Return(nil).
					Once()

				repo.EXPECT().
					GetByURN("foo").
					Run(func(urn string) {
						assert.Equal(t, urn, "foo")
					}).
					Return(&storedProvider, nil).
					Once()

				return repo
			},
			provider: provider.Provider{URN: "foo"},
			want:     &provider.Provider{URN: "foo"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := provider.NewService(tt.setupRepo(t))

			got, err := s.CreateProvider(context.Background(), tt.provider)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
