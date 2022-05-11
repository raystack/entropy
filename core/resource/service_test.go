package resource_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/core/resource/mocks"
	"github.com/odpf/entropy/pkg/errors"
)

var (
	sampleResource = resource.Resource{
		URN:     "foo:bar:baz",
		Kind:    "foo",
		Name:    "baz",
		Project: "bar",
	}

	frozenTime = time.Unix(1650536955, 0)
	deadClock  = func() time.Time { return frozenTime }
)

func TestService_GetResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		urn     string
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "NotFound",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(nil, errors.ErrNotFound).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{}, deadClock)
			},
			urn:     "foo:bar:baz",
			wantErr: errors.ErrNotFound,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					GetByURN(mock.Anything, mock.Anything).
					Return(&sampleResource, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{}, deadClock)
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
		setup   func(t *testing.T) *resource.Service
		parent  string
		kind    string
		want    []resource.Resource
		wantErr error
	}{
		{
			name: "EmptyResult",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{}, deadClock)
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "RepositoryError",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return(nil, errRepoFailure).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{}, deadClock)
			},
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything, mock.Anything).
					Return([]*resource.Resource{&sampleResource}, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{}, deadClock)
			},
			want:    []resource.Resource{sampleResource},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.ListResources(context.Background(), tt.parent, tt.kind)
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

func TestService_CreateResource(t *testing.T) {
	t.Parallel()

	errSample := errors.New("some failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		res     resource.Resource
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errSample).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				return resource.NewService(nil, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "parent",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "CreateResourceFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "parent",
					}, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errSample).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "parent",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "AlreadyExists",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "parent",
					}, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.ErrConflict).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "parent",
			},
			want:    nil,
			wantErr: errors.ErrConflict,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&resource.Resource{
						Kind:    "mock",
						Name:    "child",
						Project: "parent",
						State:   resource.State{Status: resource.StatusCompleted},
					}, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:    "mock",
				Name:    "child",
				Project: "parent",
			},
			want: &resource.Resource{
				URN:       "urn:odpf:entropy:mock:parent:child",
				Kind:      "mock",
				Name:      "child",
				Project:   "parent",
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
		URN:       "urn:odpf:entropy:mock:parent:child",
		Kind:      "mock",
		Name:      "child",
		Project:   "parent",
		State:     resource.State{Status: resource.StatusCompleted},
		CreatedAt: frozenTime,
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		urn     string
		newSpec resource.Spec
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:parent:child").
					Return(nil, errors.ErrNotFound).
					Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:parent:child",
			newSpec: resource.Spec{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.ErrInvalid).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:parent:child").
					Return(&testResource, nil).
					Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:parent:child",
			newSpec: resource.Spec{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: errors.ErrInvalid,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&testResource, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:parent:child").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(testErr)

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:parent:child",
			newSpec: resource.Spec{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: testErr,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().
					Plan(mock.Anything, mock.Anything, mock.Anything).
					Return(&testResource, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:parent:child").
					Return(&testResource, nil).Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).Twice()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:parent:child",
			newSpec: resource.Spec{Configs: map[string]interface{}{"foo": "bar"}},
			want: &resource.Resource{
				URN:       "urn:odpf:entropy:mock:parent:child",
				Kind:      "mock",
				Name:      "child",
				Project:   "parent",
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
				State:     resource.State{Status: resource.StatusPending},
				Spec: resource.Spec{
					Configs: map[string]interface{}{"foo": "bar"},
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
	testResource := resource.Resource{
		URN:       "urn:odpf:entropy:mock:parent:child",
		Kind:      "mock",
		Name:      "child",
		Project:   "parent",
		State:     resource.State{Status: resource.StatusCompleted},
		CreatedAt: frozenTime,
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		urn     string
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(nil, testErr).
					Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			wantErr: testErr,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(testErr).
					Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).
					Once()
				return resource.NewService(resourceRepo, nil, deadClock)
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

	sampleAction := resource.Action{
		Name:   "scale",
		Params: map[string]interface{}{"replicas": 8},
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		urn     string
		action  resource.Action
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "NotFound",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(nil, errors.ErrNotFound).
					Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleResolutionFailure",
			setup: func(t *testing.T) *resource.Service {
				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().
					Get("mock").
					Return(nil, errors.ErrNotFound).
					Once()

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

				return resource.NewService(resourceRepo, moduleReg, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInvalid,
		},
		{
			name: "PlanFailure",
			setup: func(t *testing.T) *resource.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().
					Plan(mock.Anything, mock.Anything, sampleAction).
					Return(nil, errors.New("failed")).
					Once()

				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().
					Get("mock").
					Return(mockModule, nil).
					Once()

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

				return resource.NewService(resourceRepo, moduleReg, deadClock)
			},
			urn:     "urn:odpf:entropy:mock:foo:bar",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().
					Plan(mock.Anything, mock.Anything, sampleAction).
					Return(&resource.Resource{
						URN:     "urn:odpf:entropy:mock:foo:bar",
						Kind:    "mock",
						Project: "foo",
						Name:    "bar",
						State:   resource.State{Status: resource.StatusPending},
					}, nil).
					Once()

				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().
					Get("mock").
					Return(mockModule, nil).
					Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn:odpf:entropy:mock:foo:bar").
					Return(&resource.Resource{
						URN:       "urn:odpf:entropy:mock:foo:bar",
						Kind:      "mock",
						Project:   "foo",
						Name:      "bar",
						CreatedAt: frozenTime,
					}, nil).
					Once()
				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				return resource.NewService(resourceRepo, moduleReg, deadClock)
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
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
