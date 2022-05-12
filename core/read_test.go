package core_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestService_GetResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		urn     string
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "NotFound",
			setup: func(t *testing.T) *core.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(nil, errors.ErrNotFound).
					Once()
				return core.New(repo, nil, nil, nil)
			},
			urn:     "foo:bar:baz",
			wantErr: errors.ErrNotFound,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(&sampleResource, nil).
					Once()
				return core.New(repo, nil, deadClock, nil)
			},
			urn:     "foo:bar:baz",
			want:    &sampleResource,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.GetResource(context.Background(), tt.urn)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ListResources(t *testing.T) {
	t.Parallel()

	errRepoFailure := errors.New("some repository error")

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		project string
		kind    string
		want    []resource.Resource
		wantErr error
	}{
		{
			name: "EmptyResult",
			setup: func(t *testing.T) *core.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
				return core.New(repo, nil, deadClock, nil)
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "RepositoryError",
			setup: func(t *testing.T) *core.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, errRepoFailure).
					Once()
				return core.New(repo, nil, deadClock, nil)
			},
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return([]*resource.Resource{&sampleResource}, nil).
					Once()
				return core.New(repo, nil, deadClock, nil)
			},
			want:    []resource.Resource{sampleResource},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.ListResources(context.Background(), tt.project, tt.kind)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}