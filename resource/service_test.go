package resource_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/resource"
)

func TestService_CreateResource(t *testing.T) {
	t.Run("test create new resource", func(t *testing.T) {
		argResource := &resource.Resource{
			URN:     "p-testdata-gl-testname-log",
			Name:    "testname",
			Parent:  "p-testdata-gl",
			Kind:    "log",
			Configs: map[string]interface{}{},
			Labels:  map[string]string{},
		}
		currentTime := time.Now()
		want := &resource.Resource{
			URN:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}
		wantErr := error(nil)

		mockRepo := &mocks.ResourceRepository{}
		mockRepo.EXPECT().Create(mock.Anything).Run(func(r *resource.Resource) {
			assert.Equal(t, resource.StatusPending, r.Status)
		}).Return(nil).Once()

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&resource.Resource{
			URN:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}, nil).Once()

		s := resource.NewService(mockRepo)
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
		argResource := &resource.Resource{
			URN:     "p-testdata-gl-testname-log",
			Name:    "testname",
			Parent:  "p-testdata-gl",
			Kind:    "log",
			Configs: map[string]interface{}{},
			Labels:  map[string]string{},
		}
		want := (*resource.Resource)(nil)
		wantErr := resource.ErrResourceAlreadyExists
		mockRepo.EXPECT().Create(mock.Anything).Run(func(r *resource.Resource) {
			assert.Equal(t, resource.StatusPending, r.Status)
		}).Return(resource.ErrResourceAlreadyExists).Once()

		s := resource.NewService(mockRepo)
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
		currentTime := time.Now()
		updatedTime := time.Now()
		want := &resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}
		wantErr := error(nil)

		mockRepo.EXPECT().Update(mock.Anything).Run(func(r *resource.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", r.URN)
			assert.Equal(t, resource.StatusPending, r.Status)
			assert.Equal(t, currentTime, r.CreatedAt)
		}).Return(nil)

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}, nil).Once()

		s := resource.NewService(mockRepo)
		got, err := s.UpdateResource(context.Background(), &resource.Resource{
			URN:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		})
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

		want := (*resource.Resource)(nil)
		wantErr := resource.ErrResourceNotFound

		mockRepo.EXPECT().
			Update(mock.Anything).
			Return(resource.ErrResourceNotFound).
			Once()

		s := resource.NewService(mockRepo)
		got, err := s.UpdateResource(context.Background(), &resource.Resource{
			URN:       "p-testdata-gl-testname-log",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "log",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})
}

func TestService_GetResource(t *testing.T) {
	t.Run("test get existing resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}
		currentTime := time.Now()
		updatedTime := time.Now()
		want := &resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}
		wantErr := error(nil)

		mockRepo.EXPECT().GetByURN("p-testdata-gl-testname-log").Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusPending,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}, nil).Once()

		s := resource.NewService(mockRepo)
		got, err := s.GetResource(context.Background(), "p-testdata-gl-testname-log")
		if !errors.Is(err, wantErr) {
			t.Errorf("GetResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test get non-existent resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}

		want := (*resource.Resource)(nil)
		wantErr := resource.ErrResourceNotFound

		mockRepo.EXPECT().
			GetByURN(mock.Anything).
			Return(nil, resource.ErrResourceNotFound).
			Once()

		s := resource.NewService(mockRepo)
		got, err := s.GetResource(context.Background(), "p-testdata-gl-testname-log")
		if !errors.Is(err, wantErr) {
			t.Errorf("GetResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetResource() got = %v, want %v", got, want)
		}
	})
}

func TestService_ListResources(t *testing.T) {
	t.Run("test list resource", func(t *testing.T) {
		mockRepo := &mocks.ResourceRepository{}
		currentTime := time.Now()
		updatedTime := time.Now()
		want := []*resource.Resource{{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusCompleted,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}}
		wantErr := error(nil)

		mockRepo.EXPECT().List(map[string]string{"parent": "p-testdata-gl", "kind": "log"}).Return([]*resource.Resource{{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    map[string]string{},
			Status:    resource.StatusCompleted,
			CreatedAt: currentTime,
			UpdatedAt: updatedTime,
		}}, nil).Once()

		s := resource.NewService(mockRepo)
		got, err := s.ListResources(context.Background(), "p-testdata-gl", "log")
		if !errors.Is(err, wantErr) {
			t.Errorf("ListResources() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ListResources() got = %v, want %v", got, want)
		}
	})
}
