package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/core/resource/mocks"
)

var sampleResource = resource.Resource{
	URN:    "foo:bar:baz",
	Kind:   "foo",
	Name:   "baz",
	Parent: "bar",
}

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
					GetByURN(mock.Anything).
					Return(nil, resource.ErrResourceNotFound).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{})
			},
			urn:     "foo:bar:baz",
			wantErr: resource.ErrResourceNotFound,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					GetByURN(mock.Anything).
					Return(&sampleResource, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{})
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
					List(mock.Anything).
					Return(nil, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{})
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "RepositoryError",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything).
					Return(nil, errRepoFailure).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{})
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Success",
			setup: func(t *testing.T) *resource.Service {
				repo := &mocks.ResourceRepository{}
				repo.EXPECT().
					List(mock.Anything).
					Return([]*resource.Resource{&sampleResource}, nil).
					Once()
				return resource.NewService(repo, &mocks.ModuleRegistry{})
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

				return resource.NewService(nil, modReg)
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
				resourceRepo.EXPECT().Create(mock.Anything).Return(errSample).Once()

				return resource.NewService(resourceRepo, modReg)
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
			name: "SyncFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusError, errSample).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything).Return(nil).Once()

				return resource.NewService(resourceRepo, modReg)
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
			name: "UpdateFailure",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(nil).Once()
				mod.EXPECT().Apply(mock.Anything).Return(resource.StatusCompleted, nil).Once().Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().Create(mock.Anything).Return(nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything).Return(errSample).Once()

				return resource.NewService(resourceRepo, modReg)
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
				resourceRepo.EXPECT().Create(mock.Anything).Return(nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything).Return(nil)

				return resource.NewService(resourceRepo, modReg)
			},
			res: resource.Resource{
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
			},
			want: &resource.Resource{
				URN:    "parent-child-mock",
				Kind:   "mock",
				Name:   "child",
				Parent: "parent",
				Status: resource.StatusCompleted,
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
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_UpdateResource(t *testing.T) {
	t.Parallel()
	testErr := errors.New("failed")
	testResource := resource.Resource{
		URN:    "parent-child-mock",
		Kind:   "mock",
		Name:   "child",
		Parent: "parent",
		Status: resource.StatusCompleted,
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
					GetByURN("parent-child-mock").
					Return(nil, resource.ErrResourceNotFound).
					Once()

				return resource.NewService(resourceRepo, nil)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: resource.ErrResourceNotFound,
		},
		{
			name: "ModuleValidationError",
			setup: func(t *testing.T) *resource.Service {
				mod := &mocks.Module{}
				mod.EXPECT().ID().Return("mock").Once()
				mod.EXPECT().Validate(mock.Anything).Return(resource.ErrModuleConfigParseFailed).Once()

				modReg := &mocks.ModuleRegistry{}
				modReg.EXPECT().Get("mock").Return(mod, nil).Twice()

				resourceRepo := &mocks.ResourceRepository{}
				resourceRepo.EXPECT().
					GetByURN("parent-child-mock").
					Return(&testResource, nil).
					Once()

				return resource.NewService(resourceRepo, modReg)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want:    nil,
			wantErr: resource.ErrModuleConfigParseFailed,
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
					GetByURN("parent-child-mock").
					Return(&testResource, nil).
					Once()

				resourceRepo.EXPECT().Update(mock.Anything).Return(testErr)

				return resource.NewService(resourceRepo, modReg)
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
				resourceRepo.EXPECT().GetByURN("parent-child-mock").Return(&testResource, nil).Once()
				resourceRepo.EXPECT().Update(mock.Anything).Return(nil).Twice()

				return resource.NewService(resourceRepo, modReg)
			},
			urn:     "parent-child-mock",
			updates: resource.Updates{Configs: map[string]interface{}{"foo": "bar"}},
			want: &resource.Resource{
				URN:     "parent-child-mock",
				Kind:    "mock",
				Name:    "child",
				Parent:  "parent",
				Status:  "STATUS_COMPLETED",
				Configs: map[string]interface{}{"foo": "bar"},
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
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
