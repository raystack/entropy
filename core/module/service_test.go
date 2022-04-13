package module_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/module/mocks"
	"github.com/odpf/entropy/core/resource"
)

func TestService_Sync(t *testing.T) {
	t.Parallel()

	frozenTime := time.Now()
	sampleApplyErr := errors.New("apply failed")

	sampleResource := resource.Resource{
		URN:       "p-testdata-gl-testname-mock",
		Name:      "testname",
		Parent:    "p-testdata-gl",
		Kind:      "mock",
		Configs:   map[string]interface{}{},
		Labels:    map[string]string{},
		Status:    resource.StatusPending,
		CreatedAt: frozenTime,
		UpdatedAt: frozenTime,
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) *module.Service
		r       resource.Resource
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "Success",
			setup: func(t *testing.T) *module.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().ID().Return("mock")
				mockModule.EXPECT().Apply(sampleResource).Return(resource.StatusCompleted, nil).Once()

				mockModuleRepo := &mocks.ModuleRepository{}
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()

				return module.NewService(mockModuleRepo)
			},
			r: sampleResource,
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusCompleted,
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: nil,
		},
		{
			name: "ModuleNotFound",
			setup: func(t *testing.T) *module.Service {
				mockModuleRepo := &mocks.ModuleRepository{}
				mockModuleRepo.EXPECT().Get("mock").Return(nil, module.ErrModuleNotFound).Once()
				return module.NewService(mockModuleRepo)
			},
			r: sampleResource,
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusError,
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: module.ErrModuleNotFound,
		},
		{
			name: "ApplyFailure",
			setup: func(t *testing.T) *module.Service {
				mockModule := &mocks.Module{}
				mockModule.EXPECT().ID().Return("mock")
				mockModule.EXPECT().Apply(sampleResource).Return(resource.StatusError, sampleApplyErr).Once()

				mockModuleRepo := &mocks.ModuleRepository{}
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()

				return module.NewService(mockModuleRepo)
			},
			r: sampleResource,
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusError,
				CreatedAt: frozenTime,
				UpdatedAt: frozenTime,
			},
			wantErr: sampleApplyErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleSvc := tt.setup(t)

			got, err := moduleSvc.Sync(context.Background(), tt.r)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Validate(t *testing.T) {
	type fields struct {
		moduleRepository module.Repository
	}
	type args struct {
		ctx context.Context
		r   *resource.Resource
	}

	currentTime := time.Now()
	r := &resource.Resource{
		URN:       "p-testdata-gl-testname-mock",
		Name:      "testname",
		Parent:    "p-testdata-gl",
		Kind:      "mock",
		Configs:   map[string]interface{}{},
		Labels:    map[string]string{},
		Status:    resource.StatusPending,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
	validateFailedErr := errors.New("some validation failure error")

	mockModule := &mocks.Module{}
	mockModule.EXPECT().ID().Return("mock")
	mockModuleRepo := &mocks.ModuleRepository{}

	tests := []struct {
		name    string
		setup   func(t *testing.T)
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "test validate success",
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
				mockModule.EXPECT().Validate(mock.Anything).Return(nil).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			wantErr: nil,
		},
		{
			name: "test validate module not found error",
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(nil, module.ErrModuleNotFound).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			wantErr: module.ErrModuleNotFound,
		},
		{
			name: "test validation failed",
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
				mockModule.EXPECT().Validate(mock.Anything).Return(validateFailedErr)
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			wantErr: validateFailedErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := module.NewService(tt.fields.moduleRepository)
			tt.setup(t)
			if err := s.Validate(tt.args.ctx, *tt.args.r); !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Act(t *testing.T) {
	type fields struct {
		moduleRepository module.Repository
	}
	type args struct {
		ctx    context.Context
		r      *resource.Resource
		action string
		params map[string]interface{}
	}

	mockModule := &mocks.Module{}
	mockModule.EXPECT().ID().Return("mock")
	mockModuleRepo := &mocks.ModuleRepository{}

	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(t *testing.T)
		want    map[string]interface{}
		wantErr error
	}{
		{
			name: "test successfully applying action",
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r: &resource.Resource{
					URN:    "p-testdata-gl-testing-mock",
					Name:   "testing",
					Parent: "p-testdata-gl",
					Kind:   "mock",
					Configs: map[string]interface{}{
						"mock": true,
					},
					Labels:    nil,
					Status:    "COMPLETED",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				action: "test",
				params: map[string]interface{}{},
			},
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
				mockModule.EXPECT().Act(mock.Anything, "test", map[string]interface{}{}).Return(map[string]interface{}{
					"mock": true,
				}, nil)
			},
			want: map[string]interface{}{
				"mock": true,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := module.NewService(tt.fields.moduleRepository)
			tt.setup(t)
			got, err := s.Act(tt.args.ctx, *tt.args.r, tt.args.action, tt.args.params)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Act() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Act() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Log(t *testing.T) {
	type fields struct {
		moduleRepository module.Repository
	}
	type args struct {
		ctx    context.Context
		r      *resource.Resource
		filter map[string]string
	}

	mockModule := &mocks.Module{}
	mockModule.EXPECT().ID().Return("mock")
	mockModuleLogger := &mocks.LoggableModule{}
	mockModuleLogger.EXPECT().ID().Return("mock")
	mockModuleRepo := &mocks.ModuleRepository{}

	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(*testing.T)
		want    func(*testing.T, <-chan module.LogChunk) bool
		wantErr error
	}{
		{
			name: "test log streaming not supported",
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: context.Background(),
				r: &resource.Resource{
					URN:    "p-testdata-gl-testing-mock",
					Name:   "testing",
					Parent: "p-testdata-gl",
					Kind:   "mock",
					Configs: map[string]interface{}{
						"mock": true,
					},
					Labels:    nil,
					Status:    "COMPLETED",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				filter: map[string]string{},
			},
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
			},
			want: func(t *testing.T, chunks <-chan module.LogChunk) bool {
				return assert.Nil(t, chunks)
			},
			wantErr: module.ErrLogStreamingUnsupported,
		},
		{
			name: "test log streaming",
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: context.Background(),
				r: &resource.Resource{
					URN:    "p-testdata-gl-testing-mock",
					Name:   "testing",
					Parent: "p-testdata-gl",
					Kind:   "mock",
					Configs: map[string]interface{}{
						"mock": true,
					},
					Labels:    nil,
					Status:    "COMPLETED",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				filter: map[string]string{},
			},
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(mockModuleLogger, nil).Once()
				mockModuleLogger.EXPECT().Log(mock.Anything, mock.Anything, map[string]string{}).Return(make(chan module.LogChunk), nil)
			},
			want: func(t *testing.T, chunks <-chan module.LogChunk) bool {
				return assert.NotNil(t, chunks)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			s := module.NewService(tt.fields.moduleRepository)

			got, err := s.Log(tt.args.ctx, *tt.args.r, tt.args.filter)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Log() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want(t, got) {
				t.Errorf("Log() got = %v, want not nil", got)
			}
		})
	}
}
