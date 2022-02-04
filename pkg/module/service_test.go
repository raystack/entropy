package module

import (
	"context"
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestService_TriggerSync(t *testing.T) {
	type fields struct {
		resourceRepository store.ResourceRepository
		moduleRepository   store.ModuleRepository
	}
	type args struct {
		ctx context.Context
		urn string
	}

	type test struct {
		name    string
		fields  fields
		args    args
		wantErr error
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

	mockResourceRepo := &mocks.ResourceRepository{}
	mockResourceRepo.EXPECT().GetByURN("p-testdata-gl-testname-mock").Return(r, nil)
	mockResourceRepo.EXPECT().Update(mock.Anything).Run(func(r *domain.Resource) {
		assert.Equal(t, domain.ResourceStatusCompleted, r.Status)
	}).Return(nil).Once()

	mockModule := &mocks.Module{}
	mockModule.EXPECT().ID().Return("mock")
	mockModule.EXPECT().Apply(r).Return(domain.ResourceStatusCompleted, nil).Once()

	mockModuleRepo := &mocks.ModuleRepository{}
	mockModuleRepo.EXPECT().Get("mock").Return(mockModule, nil).Once()

	tt := test{
		name: "test trigger sync",
		fields: fields{
			resourceRepository: mockResourceRepo,
			moduleRepository:   mockModuleRepo,
		},
		args: args{
			ctx: context.Background(),
			urn: "p-testdata-gl-testname-mock",
		},
		wantErr: nil,
	}
	t.Run(tt.name, func(t *testing.T) {
		s := &Service{
			resourceRepository: tt.fields.resourceRepository,
			moduleRepository:   tt.fields.moduleRepository,
		}
		if err := s.TriggerSync(tt.args.ctx, tt.args.urn); !errors.Is(err, tt.wantErr) {
			t.Errorf("TriggerSync() error = %v, wantErr %v", err, tt.wantErr)
		}
	})

	mockResourceRepo.EXPECT().Update(mock.Anything).Run(func(r *domain.Resource) {
		assert.Equal(t, domain.ResourceStatusError, r.Status)
	}).Return(nil).Once()

	mockModuleRepo.EXPECT().Get("mock").Return(nil, store.ModuleNotFoundError).Once()

	tt = test{
		name: "test trigger sync module not found error",
		fields: fields{
			resourceRepository: mockResourceRepo,
			moduleRepository:   mockModuleRepo,
		},
		args: args{
			ctx: context.Background(),
			urn: "p-testdata-gl-testname-mock",
		},
		wantErr: store.ModuleNotFoundError,
	}
	t.Run(tt.name, func(t *testing.T) {
		s := &Service{
			resourceRepository: tt.fields.resourceRepository,
			moduleRepository:   tt.fields.moduleRepository,
		}
		if err := s.TriggerSync(tt.args.ctx, tt.args.urn); !errors.Is(err, tt.wantErr) {
			t.Errorf("TriggerSync() error = %v, wantErr %v", err, tt.wantErr)
		}
	})
}
