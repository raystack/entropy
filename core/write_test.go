package core_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestService_CreateResource(t *testing.T) {
	t.Parallel()

	errSample := errors.New("some failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		res     resource.Resource
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errSample).Once()

				return core.New(nil, mod, deadClock, nil)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "project",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "CreateResourceFailure",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "project",
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errSample).Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "project",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "AlreadyExists",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "project",
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.ErrConflict).Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "project",
			},
			want:    nil,
			wantErr: errors.ErrConflict,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "project",
						State:   resource.State{Status: resource.StatusCompleted},
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "project",
			},
			want: &resource.Resource{
				URN:       "urn:odpf:entropy:mock:project:child",
				Kind:      "mock",
				Name:      "child",
				Project:   "project",
				State:     resource.State{Status: resource.StatusCompleted},
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.CreateResource(context.Background(), tt.res)
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

func TestService_UpdateResource(t *testing.T) {
	t.Parallel()
	testErr := errors.New("failed")
	testResource := resource.Resource{
		URN:       "urn:odpf:entropy:mock:project:child",
		Kind:      "mock",
		Name:      "child",
		Project:   "project",
		State:     resource.State{Status: resource.StatusCompleted},
		CreatedAt: frozenTime,
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		urn     string
		newSpec resource.Spec
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *core.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:project:child").
					Return(nil, errors.ErrNotFound).
					Once()

				return core.New(resourceRepo, nil, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:project:child",
			newSpec: resource.Spec{Configs: []byte(`{"foo": "bar"}`)},
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.ErrInvalid).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:project:child").
					Return(&testResource, nil).
					Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:project:child",
			newSpec: resource.Spec{Configs: []byte(`{"foo": "bar"}`)},
			want:    nil,
			wantErr: errors.ErrInvalid,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&testResource, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:project:child").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(testErr)

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:project:child",
			newSpec: resource.Spec{Configs: []byte(`{"foo": "bar"}`)},
			want:    nil,
			wantErr: testErr,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						URN:     "urn:odpf:entropy:mock:project:child",
						Kind:    "mock",
						Name:    "child",
						Project: "project",
						Spec: resource.Spec{
							Configs: []byte(`{"foo": "bar"}`),
						},
						State:     resource.State{Status: resource.StatusPending},
						CreatedAt: frozenTime,
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:project:child").
					Return(&testResource, nil).Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).Twice()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:project:child",
			newSpec: resource.Spec{Configs: []byte(`{"foo": "bar"}`)},
			want: &resource.Resource{
				URN:       "urn:odpf:entropy:mock:project:child",
				Kind:      "mock",
				Name:      "child",
				Project:   "project",
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
				State:     resource.State{Status: resource.StatusPending},
				Spec: resource.Spec{
					Configs: []byte(`{"foo": "bar"}`),
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.UpdateResource(context.Background(), tt.urn, tt.newSpec)
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

func TestService_DeleteResource(t *testing.T) {
	t.Parallel()

	testErr := errors.New("failed")

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		urn     string
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *core.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(nil, testErr).
					Once()

				return core.New(resourceRepo, nil, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			wantErr: testErr,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:project:child",
						Kind:      "mock",
						Name:      "child",
						Project:   "project",
						State:     resource.State{Status: resource.StatusPending},
						CreatedAt: frozenTime,
						UpdatedAt: frozenTime,
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:project:child",
						Kind:      "mock",
						Name:      "child",
						Project:   "project",
						State:     resource.State{Status: resource.StatusCompleted},
						CreatedAt: frozenTime,
						UpdatedAt: frozenTime,
					}, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(testErr).
					Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:project:child",
						Kind:      "mock",
						Name:      "child",
						Project:   "project",
						State:     resource.State{Status: resource.StatusPending},
						CreatedAt: frozenTime,
						UpdatedAt: frozenTime,
					}, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:project:child",
						Kind:      "mock",
						Name:      "child",
						Project:   "project",
						State:     resource.State{Status: resource.StatusCompleted},
						CreatedAt: frozenTime,
						UpdatedAt: frozenTime,
					}, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).
					Once()
				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			err := svc.DeleteResource(context.Background(), tt.urn)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ApplyAction(t *testing.T) {
	t.Parallel()

	sampleAction := module.ActionRequest{
		Name:   "scale",
		Params: []byte(`{"replicas": 8}`),
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *core.Service
		urn     string
		action  module.ActionRequest
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "NotFound",
			setup: func(t *testing.T) *core.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(nil, errors.ErrNotFound).
					Once()

				return core.New(resourceRepo, nil, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleResolutionFailure",
			setup: func(t *testing.T) *core.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:     "urn:odpf:entropy:mock:foo:bar",
						Kind:    "mock",
						Project: "foo",
						Name:    "bar",
					}, nil).
					Once()

				return core.New(resourceRepo, nil, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInvalid,
		},
		{
			name: "PlanFailure",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, sampleAction).
					Return(nil, errors.New("failed")).
					Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:     "urn:odpf:entropy:mock:foo:bar",
						Kind:    "mock",
						Project: "foo",
						Name:    "bar",
						State:   resource.State{Status: resource.StatusCompleted},
					}, nil).
					Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *core.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, sampleAction).
					Return(&resource.Resource{
						URN:     "urn:odpf:entropy:mock:foo:bar",
						Kind:    "mock",
						Project: "foo",
						Name:    "bar",
						State:   resource.State{Status: resource.StatusPending},
					}, nil).
					Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:foo:bar",
						Kind:      "mock",
						Project:   "foo",
						Name:      "bar",
						CreatedAt: frozenTime,
						State:     resource.State{Status: resource.StatusCompleted},
					}, nil).
					Once()
				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				return core.New(resourceRepo, mod, deadClock, nil)
			},
			urn:    "urn:odpf:entropy:mock:foo:bar",
			action: sampleAction,
			want: &resource.Resource{
				URN:       "urn:odpf:entropy:mock:foo:bar",
				Kind:      "mock",
				Project:   "foo",
				Name:      "bar",
				State:     resource.State{Status: resource.StatusPending},
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.ApplyAction(context.Background(), tt.urn, tt.action)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), cmp.Diff(tt.want, err))
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, cmp.Diff(tt.want, got))
		})
	}
}
