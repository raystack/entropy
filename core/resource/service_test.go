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
		URN:    "foo:bar:baz",
		Kind:   "foo",
		Name:   "baz",
		Parent: "bar",
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
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(errSample).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				return resource.NewService(nil, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "CreateResourceFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errSample).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "AlreadyExists",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.ErrConflict).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want:    nil,
			wantErr: errors.ErrConflict,
		},
		{
			name: "SyncFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusError, errSample).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want: &resource.Resource{
				URN:       "parent-child-mock",
				Kind:      "mock",
				Name:      "child",
				Parent:    "parent",
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
				Status:    resource.StatusError,
			},
			wantErr: nil,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusCompleted, nil).Once().Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(errSample).Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want:    nil,
			wantErr: errSample,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusCompleted, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want: &resource.Resource{
				URN:       "parent-child-mock",
				Kind:      "mock",
				Name:      "child",
				Parent:    "parent",
				Status:    resource.StatusCompleted,
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
		URN:       "parent-child-mock",
		Kind:      "mock",
		Name:      "child",
		Parent:    "parent",
		Status:    resource.StatusCompleted,
		CreatedAt: frozenTime,
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *resource.Service
		urn     string
		updates resource.Updates
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "parent-child-mock").
					Return(nil, errors.ErrNotFound).
					Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(errors.ErrInvalid).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "parent-child-mock").
					Return(&testResource, nil).
					Once()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: errors.ErrInvalid,
		},
		{
			name: "UpdateFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "parent-child-mock").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(testErr)

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: testErr,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusCompleted, nil).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().GetByURN(mock.Anything, "parent-child-mock").Return(&testResource, nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Twice()

				return resource.NewService(resourceRepo, modReg, deadClock)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want: &resource.Resource{
				URN:       "parent-child-mock",
				Kind:      "mock",
				Name:      "child",
				Parent:    "parent",
				Status:    "STATUS_COMPLETED",
				Configs:   map[string]interface{}{"foo": "bar"},
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)

			got, err := svc.UpdateResource(context.Background(), tt.urn, tt.updates)
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
		setup   func(t *testing.T) *resource.Service
		urn     string
		updates resource.Updates
		wantErr error
	}{
		{
			name: "InternalError",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Delete(mock.Anything, mock.Anything).Return(testErr).Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			wantErr: errors.ErrInternal,
		},
		{
			name: "NotFound",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.ErrNotFound).Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			wantErr: nil,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
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
				resourceRepo.EXPECT().GetByURN(mock.Anything, "urn::foo").Return(nil, errors.ErrNotFound).Once()

				return resource.NewService(resourceRepo, nil, deadClock)
			},
			urn:     "urn::foo",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
		{
			name: "ModuleResolutionFailure",
			setup: func(t *testing.T) *resource.Service {
				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().Get("mock").Return(nil, errors.ErrNotFound).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn::foo").
					Return(&resource.Resource{
						URN:  "urn::foo",
						Kind: "mock",
					}, nil).
					Once()

				return resource.NewService(resourceRepo, moduleReg, deadClock)
			},
			urn:     "urn::foo",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "ActFailure",
			setup: func(t *testing.T) *resource.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().
					Act(mock.Anything, "scale", mock.Anything).
					Return(nil, errors.New("failed")).
					Once()

				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().Get("mock").Return(mockModule, nil).Once()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn::foo").
					Return(&resource.Resource{
						URN:  "urn::foo",
						Kind: "mock",
					}, nil).
					Once()

				return resource.NewService(resourceRepo, moduleReg, deadClock)
			},
			urn:     "urn::foo",
			action:  sampleAction,
			want:    nil,
			wantErr: errors.ErrInternal,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().
					Act(mock.Anything, "scale", mock.Anything).
					Return(nil, nil).
					Once()
				mockModule.EXPECT().
					Apply(mock.Anything).
					Return(resource.StatusCompleted, nil).
					Once()

				moduleReg := &mocks.ModuleRegistry{}
				moduleReg.EXPECT().Get("mock").Return(mockModule, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN(mock.Anything, "urn::foo").
					Return(&resource.Resource{
						URN:       "urn::foo",
						Kind:      "mock",
						CreatedAt: frozenTime,
					}, nil).
					Once()
				resourceRepo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil).
					Once()

				return resource.NewService(resourceRepo, moduleReg, deadClock)
			},
			urn:    "urn::foo",
			action: sampleAction,
			want: &resource.Resource{
				URN:       "urn::foo",
				Kind:      "mock",
				Status:    resource.StatusCompleted,
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
