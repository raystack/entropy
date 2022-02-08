package module

import (
	"context"
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/store"
	"reflect"
	"testing"
	"time"
)

func TestService_Sync(t *testing.T) {
	type fields struct {
		moduleRepository store.ModuleRepository
	}
	type args struct {
		ctx context.Context
		r   *domain.Resource
	}
	currentTime := time.Now()
	r := &domain.Resource{
		Urn:       "p-testdata-gl-testname-mock",
		Name:      "testname",
		Parent:    "p-testdata-gl",
		Kind:      "mock",
		Configs:   map[string]interface{}{},
		Labels:    map[string]string{},
		Status:    domain.ResourceStatusPending,
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
		want    *domain.Resource
		wantErr error
	}{
		{
			name: "test sync completed",
			setup: func(t *testing.T) {
				mockModule.EXPECT().Apply(r).Return(domain.ResourceStatusCompleted, nil).Once()
				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			want: &domain.Resource{
				Urn:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    domain.ResourceStatusCompleted,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: nil,
		},
		{
			name: "test sync module not found error",
			setup: func(t *testing.T) {
				mockModuleRepo.EXPECT().Get("mock").Return(nil, store.ModuleNotFoundError).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			want: &domain.Resource{
				Urn:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    domain.ResourceStatusError,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: store.ModuleNotFoundError,
		},
		{
			name: "test sync module error while applying",
			setup: func(t *testing.T) {
				mockModule.EXPECT().Apply(r).Return(domain.ResourceStatusError, applyFailedErr).Once()

				mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()
			},
			fields: fields{
				moduleRepository: mockModuleRepo,
			},
			args: args{
				ctx: nil,
				r:   r,
			},
			want: &domain.Resource{
				Urn:       "p-testdata-gl-testname-mock",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "mock",
				Configs:   map[string]interface{}{},
				Labels:    map[string]string{},
				Status:    domain.ResourceStatusError,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			wantErr: applyFailedErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				moduleRepository: tt.fields.moduleRepository,
			}
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
