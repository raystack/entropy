package handlersv1

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/internal/mocks"
	"github.com/odpf/entropy/module"
	"github.com/odpf/entropy/resource"
)

func TestAPIServer_CreateResource(t *testing.T) {
	t.Run("test create new resource", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
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
				Status:    entropyv1beta1.Resource_STATUS_COMPLETED,
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
		resourceService.EXPECT().CreateResource(mock.Anything, mock.Anything).Run(func(ctx context.Context, res *resource.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", res.URN)
		}).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusPending,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}, nil).Once()

		resourceService.EXPECT().UpdateResource(mock.Anything, mock.Anything).Run(func(ctx context.Context, res *resource.Resource) {
			assert.Equal(t, "p-testdata-gl-testname-log", res.URN)
			assert.Equal(t, resource.StatusCompleted, res.Status)
		}).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(nil)
		moduleService.EXPECT().Sync(mock.Anything, mock.Anything).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}, nil)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
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
			Return(nil, resource.ErrResourceAlreadyExists).
			Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(nil)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.CreateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test create resource of unknown kind", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.CreateResourceResponse)(nil)
		wantErr := status.Error(codes.InvalidArgument, "failed to find module to deploy this kind")

		ctx := context.Background()
		request := &entropyv1beta1.CreateResourceRequest{
			Resource: &entropyv1beta1.Resource{
				Name:    "testname",
				Parent:  "p-testdata-gl",
				Kind:    "unknown",
				Configs: configsStructValue,
				Labels:  nil,
			},
		}

		resourceService := &mocks.ResourceService{}

		resourceService.EXPECT().CreateResource(mock.Anything, mock.Anything).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-unknown",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "unkown",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusPending,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(module.ErrModuleNotFound)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.CreateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CreateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test create resource with validation failure", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.CreateResourceResponse)(nil)
		wantErr := status.Error(codes.InvalidArgument, "failed to parse configs")

		ctx := context.Background()
		request := &entropyv1beta1.CreateResourceRequest{
			Resource: &entropyv1beta1.Resource{
				Name:    "testname",
				Parent:  "p-testdata-gl",
				Kind:    "unknown",
				Configs: configsStructValue,
				Labels:  nil,
			},
		}

		resourceService := &mocks.ResourceService{}

		resourceService.EXPECT().CreateResource(mock.Anything, mock.Anything).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-unknown",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "unkown",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusPending,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(module.ErrModuleConfigParseFailed)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
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
				Status:    entropyv1beta1.Resource_STATUS_COMPLETED,
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
			GetResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(&resource.Resource{
				URN:    "p-testdata-gl-testname-log",
				Name:   "testname",
				Parent: "p-testdata-gl",
				Kind:   "log",
				Configs: map[string]interface{}{
					"replicas": "9",
				},
				Labels:    nil,
				Status:    resource.StatusCompleted,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			}, nil).Once()

		resourceService.EXPECT().
			UpdateResource(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, res *resource.Resource) {
				assert.Equal(t, resource.StatusPending, res.Status)
			}).
			Return(&resource.Resource{
				URN:    "p-testdata-gl-testname-log",
				Name:   "testname",
				Parent: "p-testdata-gl",
				Kind:   "log",
				Configs: map[string]interface{}{
					"replicas": "10",
				},
				Labels:    nil,
				Status:    resource.StatusPending,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil).Once()

		resourceService.EXPECT().
			UpdateResource(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, res *resource.Resource) {
				assert.Equal(t, resource.StatusCompleted, res.Status)
			}).
			Return(&resource.Resource{
				URN:    "p-testdata-gl-testname-log",
				Name:   "testname",
				Parent: "p-testdata-gl",
				Kind:   "log",
				Configs: map[string]interface{}{
					"replicas": "10",
				},
				Labels:    nil,
				Status:    resource.StatusCompleted,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(nil)
		moduleService.EXPECT().Sync(mock.Anything, mock.Anything).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"replicas": "10",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
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
			GetResource(mock.Anything, mock.Anything).
			Return(nil, resource.ErrResourceNotFound).Once()

		moduleService := &mocks.ModuleService{}

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.UpdateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test update resource with unknown kind", func(t *testing.T) {
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.UpdateResourceResponse)(nil)
		wantErr := status.Error(codes.InvalidArgument, "failed to find module to deploy this kind")

		ctx := context.Background()
		request := &entropyv1beta1.UpdateResourceRequest{
			Urn:     "p-testdata-gl-testname-log",
			Configs: configsStructValue,
		}

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			UpdateResource(mock.Anything, mock.Anything).
			Return(&resource.Resource{
				URN: "p-testdata-gl-testname-log",
			}, nil).Once()
		resourceService.EXPECT().
			GetResource(mock.Anything, mock.Anything).
			Return(&resource.Resource{
				URN: "p-testdata-gl-testname-log",
			}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(module.ErrModuleNotFound)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.UpdateResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("UpdateResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test update resource with validation failure", func(t *testing.T) {
		configsStructValue, _ := structpb.NewValue(map[string]interface{}{
			"replicas": "10",
		})
		want := (*entropyv1beta1.UpdateResourceResponse)(nil)
		wantErr := status.Error(codes.InvalidArgument, "failed to parse configs")

		ctx := context.Background()
		request := &entropyv1beta1.UpdateResourceRequest{
			Urn:     "p-testdata-gl-testname-log",
			Configs: configsStructValue,
		}

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			UpdateResource(mock.Anything, mock.Anything).
			Return(&resource.Resource{
				URN: "p-testdata-gl-testname-log",
			}, nil).Once()
		resourceService.EXPECT().
			GetResource(mock.Anything, mock.Anything).
			Return(&resource.Resource{
				URN: "p-testdata-gl-testname-log",
			}, nil).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Validate(mock.Anything, mock.Anything).Return(module.ErrModuleConfigParseFailed)

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
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

func TestAPIServer_GetResource(t *testing.T) {
	t.Run("test get resource", func(t *testing.T) {
		r := &resource.Resource{
			URN:       "p-testdata-gl-testname-mock",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "mock",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusCompleted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		rProto, _ := resourceToProto(r)
		argsRequest := &entropyv1beta1.GetResourceRequest{
			Urn: "p-testdata-gl-testname-mock",
		}
		want := &entropyv1beta1.GetResourceResponse{
			Resource: rProto,
		}
		wantErr := error(nil)

		mockResourceService := &mocks.ResourceService{}
		mockResourceService.EXPECT().GetResource(mock.Anything, mock.Anything).Return(r, nil).Once()

		mockModuleService := &mocks.ModuleService{}

		server := APIServer{
			resourceService: mockResourceService,
			moduleService:   mockModuleService,
		}
		got, err := server.GetResource(context.TODO(), argsRequest)
		if !errors.Is(err, wantErr) {
			t.Errorf("GetResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test get non existent resource", func(t *testing.T) {
		argsRequest := &entropyv1beta1.GetResourceRequest{
			Urn: "p-testdata-gl-testname-mock",
		}
		want := (*entropyv1beta1.GetResourceResponse)(nil)
		wantErr := status.Error(codes.NotFound, "could not find resource with given urn")

		mockResourceService := &mocks.ResourceService{}
		mockResourceService.EXPECT().GetResource(mock.Anything, mock.Anything).Return(nil, resource.ErrResourceNotFound).Once()

		mockModuleService := &mocks.ModuleService{}

		server := APIServer{
			resourceService: mockResourceService,
			moduleService:   mockModuleService,
		}
		got, err := server.GetResource(context.TODO(), argsRequest)
		if !errors.Is(err, wantErr) {
			t.Errorf("GetResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetResource() got = %v, want %v", got, want)
		}
	})
}

func TestAPIServer_ListResource(t *testing.T) {
	t.Run("test list resource", func(t *testing.T) {
		r := &resource.Resource{
			URN:       "p-testdata-gl-testname-mock",
			Name:      "testname",
			Parent:    "p-testdata-gl",
			Kind:      "mock",
			Configs:   map[string]interface{}{},
			Labels:    map[string]string{},
			Status:    resource.StatusCompleted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		rProto, _ := resourceToProto(r)
		argsRequest := &entropyv1beta1.ListResourcesRequest{
			Parent: "p-testdata-gl",
			Kind:   "mock",
		}
		want := &entropyv1beta1.ListResourcesResponse{
			Resources: []*entropyv1beta1.Resource{rProto},
		}
		wantErr := error(nil)

		mockResourceService := &mocks.ResourceService{}
		mockResourceService.EXPECT().ListResources(mock.Anything, r.Parent, r.Kind).Return([]*resource.Resource{r}, nil).Once()

		mockModuleService := &mocks.ModuleService{}

		server := APIServer{
			resourceService: mockResourceService,
			moduleService:   mockModuleService,
		}
		got, err := server.ListResources(context.TODO(), argsRequest)
		if !errors.Is(err, wantErr) {
			t.Errorf("ListResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ListResource() got = %v, want %v", got, want)
		}
	})
}

func TestAPIServer_DeleteResource(t *testing.T) {
	t.Run("test delete existing resource", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
		want := &entropyv1beta1.DeleteResourceResponse{}
		wantErr := error(nil)

		ctx := context.Background()
		request := &entropyv1beta1.DeleteResourceRequest{
			Urn: "p-testdata-gl-testname-log",
		}

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			GetResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(&resource.Resource{
				URN:    "p-testdata-gl-testname-log",
				Name:   "testname",
				Parent: "p-testdata-gl",
				Kind:   "log",
				Configs: map[string]interface{}{
					"replicas": "9",
				},
				Labels:    nil,
				Status:    resource.StatusCompleted,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil).Once()

		resourceService.EXPECT().
			DeleteResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(error(nil))

		moduleService := &mocks.ModuleService{}

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.DeleteResource(ctx, request)
		if !errors.Is(err, wantErr) {
			t.Errorf("DeleteResource() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DeleteResource() got = %v, want %v", got, want)
		}
	})

	t.Run("test delete non-existing resource", func(t *testing.T) {
		ctx := context.Background()
		request := &entropyv1beta1.DeleteResourceRequest{
			Urn: "p-testdata-gl-testname-log",
		}

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			GetResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(nil, resource.ErrResourceNotFound).Once()

		moduleService := &mocks.ModuleService{}

		providerService := &mocks.ProviderService{}
		server := NewApiServer(resourceService, moduleService, providerService)
		got, err := server.DeleteResource(ctx, request)
		if errors.Is(err, nil) {
			t.Errorf("DeleteResource() got nil error")
			return
		}
		if got != nil {
			t.Errorf("DeleteResource() got = %v, want nil", got)
		}
	})

}

func TestAPIServer_ApplyAction(t *testing.T) {
	t.Run("test applying action successfully", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Minute)
		r := &resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"log_level": "WARN",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		rDash := &resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"log_level": "INFO",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		rProto, _ := resourceToProto(r)
		want := &entropyv1beta1.ApplyActionResponse{
			Resource: rProto,
		}
		request := &entropyv1beta1.ApplyActionRequest{
			Urn:    "p-testdata-gl-testname-log",
			Action: "escalate",
		}
		wantErr := error(nil)

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			GetResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(rDash, nil).Once()

		resourceService.EXPECT().
			UpdateResource(mock.Anything, mock.Anything).
			Return(r, error(nil)).Once()

		moduleService := &mocks.ModuleService{}
		moduleService.EXPECT().Act(mock.Anything, rDash, "escalate", map[string]interface{}{}).Return(map[string]interface{}{
			"log_level": "WARN",
		}, nil).Once()
		moduleService.EXPECT().Sync(mock.Anything, mock.Anything).Run(func(_ context.Context, r *resource.Resource) {
			assert.Equal(t, "WARN", r.Configs["log_level"])
		}).Return(&resource.Resource{
			URN:    "p-testdata-gl-testname-log",
			Name:   "testname",
			Parent: "p-testdata-gl",
			Kind:   "log",
			Configs: map[string]interface{}{
				"log_level": "WARN",
			},
			Labels:    nil,
			Status:    resource.StatusCompleted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, nil).Once()

		server := APIServer{
			resourceService: resourceService,
			moduleService:   moduleService,
		}
		got, err := server.ApplyAction(context.TODO(), request)
		if !errors.Is(err, wantErr) {
			t.Errorf("ApplyAction() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ApplyAction() got = %v, want %v", got, want)
		}
	})

	t.Run("test applying action on non-existent resource", func(t *testing.T) {
		request := &entropyv1beta1.ApplyActionRequest{
			Urn:    "p-testdata-gl-testname-log",
			Action: "escalate",
		}
		want := (*entropyv1beta1.ApplyActionResponse)(nil)
		wantErr := status.Error(codes.NotFound, "could not find resource with given urn")

		resourceService := &mocks.ResourceService{}
		resourceService.EXPECT().
			GetResource(mock.Anything, "p-testdata-gl-testname-log").
			Return(nil, resource.ErrResourceNotFound).Once()

		moduleService := &mocks.ModuleService{}

		server := APIServer{
			resourceService: resourceService,
			moduleService:   moduleService,
		}
		got, err := server.ApplyAction(context.TODO(), request)
		if !errors.Is(err, wantErr) {
			t.Errorf("ApplyAction() error = %v, wantErr %v", err, wantErr)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ApplyAction() got = %v, want %v", got, want)
		}
	})
}
