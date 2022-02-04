package handlersv1

import (
	"context"
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/mocks"
	"github.com/odpf/entropy/store"
	"github.com/stretchr/testify/mock"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"testing"
	"time"
)

func TestAPIServer_CreateResource(t *testing.T) {
	t.Run("test create new resource", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := &entropyv1beta1.CreateResourceResponse{
			Resource: &entropyv1beta1.Resource{
				Urn:       "p-testdata-gl-testname-log",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "log",
				Configs:   configsStructValue,
				Labels:    nil,
				Status:    entropyv1beta1.Resource_STATUS_PENDING,
				CreatedAt: timestamppb.New(createdAt),
				UpdatedAt: timestamppb.New(updatedAt),
			},
		}
		wantErr := error(nil)

		ctx := context.Background()
		request := &entropyv1beta1.CreateResourceRequest{
			Resource: &entropyv1beta1.Resource{
				Name:    "testname",
				Parent:  "p-testdata-gl",
				Kind:    "log",
				Configs: configsStructValue,
				Labels:  nil,
			},
		}

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().CreateResource(mock.Anything, mock.Anything).Return(&domain.Resource{
			Urn:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    domain.ResourceStatusPending,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().TriggerSync(mock.Anything, "p-testdata-gl-testname-log").Return(nil)

		server := NewApiServer(resourceService, moduleService)
		got, err := server.CreateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test create duplicate resource", func(t *testing.T) {
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.CreateResourceResponse)(nil)
		wantErr := status.Error(codes.AlreadyExists, "resource already exists")

		ctx := context.Background()
		request := &entropyv1beta1.CreateResourceRequest{
			Resource: &entropyv1beta1.Resource{
				Name:    "testname",
				Parent:  "p-testdata-gl",
				Kind:    "log",
				Configs: configsStructValue,
				Labels:  nil,
			},
		}

		resourceService := &mocks.ResourceService{}

		resourceService.EXPECT().
			CreateResource(mock.Anything, mock.Anything).
			Return(nil, store.ResourceAlreadyExistsError).
			Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().TriggerSync(mock.Anything, "p-testdata-gl-testname-log").Return(nil)

		server := NewApiServer(resourceService, moduleService)
		got, err := server.CreateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})
}

func TestAPIServer_UpdateResource(t *testing.T) {
	t.Run("test update existing resource", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := &entropyv1beta1.UpdateResourceResponse{
			Resource: &entropyv1beta1.Resource{
				Urn:       "p-testdata-gl-testname-log",
				Name:      "testname",
				Parent:    "p-testdata-gl",
				Kind:      "log",
				Configs:   configsStructValue,
				Labels:    nil,
				Status:    entropyv1beta1.Resource_STATUS_PENDING,
				CreatedAt: timestamppb.New(createdAt),
				UpdatedAt: timestamppb.New(updatedAt),
			},
		}
		wantErr := error(nil)

		ctx := context.Background()
		request := &entropyv1beta1.UpdateResourceRequest{
			Urn:     "p-testdata-gl-testname-log",
			Configs: configsStructValue,
		}

		resourceService := &mocks.ResourceService{}

		resourceService.EXPECT().
			UpdateResource(mock.Anything, "p-testdata-gl-testname-log", map[string]interface{}{
				"replicas": "10",
			}).
			Return(&domain.Resource{
				Urn:    "p-testdata-gl-testname-log",
				Name:   "testname",
				Parent: "p-testdata-gl",
				Kind:   "log",
				Configs: map[string]interface{}{
					"replicas": "10",
				},
				Labels:    nil,
				Status:    domain.ResourceStatusPending,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().TriggerSync(mock.Anything, "p-testdata-gl-testname-log").Return(nil)

		server := NewApiServer(resourceService, moduleService)
		got, err := server.UpdateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test update non-existing resource", func(t *testing.T) {
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.UpdateResourceResponse)(nil)
		wantErr := status.Error(codes.NotFound, "could not find resource with given urn")

		ctx := context.Background()
		request := &entropyv1beta1.UpdateResourceRequest{
			Urn:     "p-testdata-gl-testname-log",
			Configs: configsStructValue,
		}

		resourceService := &mocks.ResourceService{}

		resourceService.EXPECT().
			UpdateResource(mock.Anything, "p-testdata-gl-testname-log", map[string]interface{}{
				"replicas": "10",
			}).
			Return(nil, store.ResourceNotFoundError).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().TriggerSync(mock.Anything, "p-testdata-gl-testname-log").Return(nil)

		server := NewApiServer(resourceService, moduleService)
		got, err := server.UpdateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})
}
