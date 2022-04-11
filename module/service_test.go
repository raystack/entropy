package module_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/module"
	"github.com/odpf/entropy/resource"
)

func TestService_Sync(t *testing.T) {
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
	applyFailedErr := errors.New("apply failed")

	mockModule := &mocks.Module{}
	mockModule.EXPECT().ID().Return("mock")
	mockModuleRepo := &mocks.ModuleRepository{}

	tests := []struct {
		name    string
		setup   func(t *testing.T)
		fields  fields
		args    args
		want    *resource.Resource
		wantErr error
	}{
		{
			name: "test sync completed",
			setup: func(t *testing.T) {
				mockModule.EXPECT().Apply(r).Return(resource.StatusCompleted, nil).Once()
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusCompleted,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: nil,
		},
		{
			name: "test sync module not found error",
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
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusError,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: module.ErrModuleNotFound,
		},
		{
			name: "test sync module error while applying",
			setup: func(t *testing.T) {
				mockModule.EXPECT().Apply(r).Return(resource.StatusError, applyFailedErr).Once()

				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			want: &resource.Resource{
				URN:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    resource.StatusError,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: applyFailedErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := module.NewService(tt.fields.moduleRepository)
			tt.setup(t)

			got, err := s.Sync(tt.args.ctx, tt.args.r)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sync() got = %v, want %v", got, tt.want)
			}
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
			if err := s.Validate(tt.args.ctx, tt.args.r); !errors.Is(err, tt.wantErr) {
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
			got, err := s.Act(tt.args.ctx, tt.args.r, tt.args.action, tt.args.params)
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

			got, err := s.Log(tt.args.ctx, tt.args.r, tt.args.filter)
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
