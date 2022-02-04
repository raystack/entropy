package resource

import (
	"context"
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	"time"
)

func TestService_CreateResource(t *testing.T) {
	t.Run("test create new resource", func(t *testing.T) {
		argResource := &domain.Resource{
			Name:    "testname",
			Parent:  "p-testdata-gl",
			Kind:    "log",
			Configs: map[string]interface{}{},
			Labels:  map[string]string{},
		}
		currentTime := time.Now()
		want := &domain.Resource{
			Urn:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    domain.ResourceStatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}
		wantErr := error(nil)

		mockRepo := &mocks.ResourceRepository{}
		mockRepo.EXPECT().Create(mock.Anything).Run(func(r *domain.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", r.Urn)
			assert.Equal(t, domain.ResourceStatusPending, r.Status)
		}).Return(nil).Once()

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&domain.Resource{
			Urn:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    domain.ResourceStatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}, nil).Once()

		s := NewService(mockRepo)
		got, err := s.CreateResource(context.Background(), argResource)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test create duplicate resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}
		argResource := &domain.Resource{
			Name:    "testname",
			Parent:  "p-testdata-gl",
			Kind:    "log",
			Configs: map[string]interface{}{},
			Labels:  map[string]string{},
		}
		want := (*domain.Resource)(nil)
		wantErr := store.ResourceAlreadyExistsError
		mockRepo.EXPECT().Create(mock.Anything).Run(func(r *domain.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", r.Urn)
			assert.Equal(t, domain.ResourceStatusPending, r.Status)
		}).Return(store.ResourceAlreadyExistsError).Once()

		s := NewService(mockRepo)
		got, err := s.CreateResource(context.Background(), argResource)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})
}

func TestService_UpdateResource(t *testing.T) {
	t.Run("test update existing resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}
		argUrn := "p-testdata-gl-testname-log"
		argConfigs := map[string]interface{}{
			"replicas": "10",
		}
		currentTime := time.Now()
		updatedTime := time.Now()
		want := &domain.Resource{
			Urn:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    domain.ResourceStatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}
		wantErr := error(nil)

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&domain.Resource{
			Urn:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    domain.ResourceStatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}, nil).Once()

		mockRepo.EXPECT().Update(mock.Anything).Run(func(r *domain.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", r.Urn)
			assert.Equal(t, domain.ResourceStatusPending, r.Status)
			assert.Equal(t, currentTime, r.CreatedAt)
		}).Return(nil)

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&domain.Resource{
			Urn:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    domain.ResourceStatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}, nil).Once()

		s := NewService(mockRepo)
		got, err := s.UpdateResource(context.Background(), argUrn, argConfigs)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test update non-existent resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}
		argUrn := "p-testdata-gl-testname-log"
		argConfigs := map[string]interface{}{
			"replicas": "10",
		}
		want := (*domain.Resource)(nil)
		wantErr := store.ResourceNotFoundError

		mockRepo.EXPECT().
			GetByURN("p-testdata-gl-testname-log").
			Return(nil, store.ResourceNotFoundError).
			Once()

		s := NewService(mockRepo)
		got, err := s.UpdateResource(context.Background(), argUrn, argConfigs)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})
}
