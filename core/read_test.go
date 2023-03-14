package core_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goto/entropy/core"
	"github.com/goto/entropy/core/mocks"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
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
				t.Helper()
				repo := &mocks.ResourceStore{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(nil, errors.ErrNotFound).
					Once()
				return core.New(repo, nil, &mocks.AsyncWorker{}, nil, nil)
			},
			urn:     "foo:bar:baz",
			wantErr: errors.ErrNotFound,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				t.Helper()
				repo := &mocks.ResourceStore{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(&sampleResource, nil).
					Once()
				mod := &mocks.ModuleService{}
				mod.EXPECT().
					GetOutput(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()

				return core.New(repo, mod, &mocks.AsyncWorker{}, deadClock, nil)
			},
			urn:     "foo:bar:baz",
			want:    &sampleResource,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

	errStoreFailure := errors.New("some store error")

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		filter  resource.Filter
		want    []resource.Resource
		wantErr error
	}{
		{
			name: "EmptyResult",
			setup: func(t *testing.T) *core.Service {
				t.Helper()
				repo := &mocks.ResourceStore{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
				return core.New(repo, nil, &mocks.AsyncWorker{}, deadClock, nil)
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "StoreError",
			setup: func(t *testing.T) *core.Service {
				t.Helper()
				repo := &mocks.ResourceStore{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, errStoreFailure).
					Once()
				return core.New(repo, nil, &mocks.AsyncWorker{}, deadClock, nil)
			},
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				t.Helper()
				repo := &mocks.ResourceStore{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return([]resource.Resource{sampleResource}, nil).
					Once()
				return core.New(repo, nil, &mocks.AsyncWorker{}, deadClock, nil)
			},
			want:    []resource.Resource{sampleResource},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := tt.setup(t)

			got, err := svc.ListResources(context.Background(), tt.filter)
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
