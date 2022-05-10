package handlersv1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/internal/server/v1/mocks"
	"github.com/odpf/entropy/pkg/errors"
)

func TestAPIServer_CreateResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	configsStructValue, _ := structpb.NewValue(map[string]interface{}{
		"replicas": "10",
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.CreateResourceRequest
		want    *entropyv1beta1.CreateResourceResponse
		wantErr error
	}{
		{
			name: "Duplicate",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					CreateResource(mock.Anything, mock.Anything).
					Return(nil, errors.ErrConflict).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Parent:  "p-testdata-gl",
					Kind:    "log",
					Configs: configsStructValue,
					Labels:  nil,
				},
			},
			want:    nil,
			wantErr: status.Error(codes.AlreadyExists, "an entity with conflicting identifier exists"),
		},
		{
			name: "InvalidRequest",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					CreateResource(mock.Anything, mock.Anything).
					Return(nil, errors.ErrInvalid).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Parent:  "p-testdata-gl",
					Kind:    "log",
					Configs: configsStructValue,
					Labels:  nil,
				},
			},
			want:    nil,
			wantErr: status.Errorf(codes.InvalidArgument, "request is not valid"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					CreateResource(mock.Anything, mock.Anything).
					Return(&resource.Resource{
						URN:       "p-testdata-gl-testname-log",
						Kind:      "log",
						Name:      "testname",
						Project:   "p-testdata-gl",
						Labels:    nil,
						CreatedAt: createdAt,
						UpdatedAt: createdAt,
						Spec: resource.Spec{
							Configs: map[string]interface{}{
								"replicas": "10",
							},
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Parent:  "p-testdata-gl",
					Kind:    "log",
					Configs: configsStructValue,
					Labels:  nil,
				},
			},
			want: &entropyv1beta1.CreateResourceResponse{
				Resource: &entropyv1beta1.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Name:      "testname",
					Parent:    "p-testdata-gl",
					Kind:      "log",
					Configs:   configsStructValue,
					Labels:    nil,
					Status:    entropyv1beta1.Resource_STATUS_PENDING,
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(createdAt),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.CreateResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestAPIServer_UpdateResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)
	configsStructValue, _ := structpb.NewValue(map[string]interface{}{
		"replicas": "10",
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.UpdateResourceRequest
		want    *entropyv1beta1.UpdateResourceResponse
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					UpdateResource(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrNotFound).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn:     "p-testdata-gl-testname-log",
				Configs: configsStructValue,
			},
			want:    nil,
			wantErr: status.Error(codes.NotFound, "requested entity not found"),
		},
		{
			name: "InvalidRequest",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					UpdateResource(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrInvalid).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn:     "p-testdata-gl-testname-log",
				Configs: configsStructValue,
			},
			want:    nil,
			wantErr: status.Errorf(codes.InvalidArgument, "request is not valid"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					UpdateResource(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(&resource.Resource{
						URN:       "p-testdata-gl-testname-log",
						Kind:      "log",
						Name:      "testname",
						Project:   "p-testdata-gl",
						Labels:    nil,
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Spec: resource.Spec{
							Configs: map[string]interface{}{
								"replicas": "10",
							},
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn:     "p-testdata-gl-testname-log",
				Configs: configsStructValue,
			},
			want: &entropyv1beta1.UpdateResourceResponse{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.UpdateResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestAPIServer_GetResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)
	configsStructValue, _ := structpb.NewValue(map[string]interface{}{
		"replicas": "10",
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.GetResourceRequest
		want    *entropyv1beta1.GetResourceResponse
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					GetResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(nil, errors.ErrNotFound).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.GetResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want:    nil,
			wantErr: status.Error(codes.NotFound, "requested entity not found"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					GetResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(&resource.Resource{
						URN:       "p-testdata-gl-testname-log",
						Kind:      "log",
						Name:      "testname",
						Project:   "p-testdata-gl",
						Labels:    nil,
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Spec: resource.Spec{
							Configs: map[string]interface{}{
								"replicas": "10",
							},
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.GetResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want: &entropyv1beta1.GetResourceResponse{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.GetResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestAPIServer_ListResources(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)
	configsStructValue, _ := structpb.NewValue(map[string]interface{}{
		"replicas": "10",
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.ListResourcesRequest
		want    *entropyv1beta1.ListResourcesResponse
		wantErr error
	}{
		{
			name: "UnhandledError",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ListResources(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("failed")).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.ListResourcesRequest{
				Parent: "p-testdata-gl",
				Kind:   "log",
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "some unexpected error occurred"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ListResources(mock.Anything, mock.Anything, mock.Anything).
					Return([]resource.Resource{
						{
							URN:       "p-testdata-gl-testname-log",
							Kind:      "log",
							Name:      "testname",
							Project:   "p-testdata-gl",
							Labels:    nil,
							CreatedAt: createdAt,
							UpdatedAt: updatedAt,
							Spec: resource.Spec{
								Configs: map[string]interface{}{
									"replicas": "10",
								},
							},
							State: resource.State{
								Status: resource.StatusPending,
							},
						},
					}, nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.ListResourcesRequest{
				Parent: "p-testdata-gl",
				Kind:   "log",
			},
			want: &entropyv1beta1.ListResourcesResponse{
				Resources: []*entropyv1beta1.Resource{
					{
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
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.ListResources(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestAPIServer_DeleteResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.DeleteResourceRequest
		want    *entropyv1beta1.DeleteResourceResponse
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					DeleteResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(errors.ErrNotFound).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.DeleteResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want:    nil,
			wantErr: status.Error(codes.NotFound, "requested entity not found"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					DeleteResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.DeleteResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want: &entropyv1beta1.DeleteResourceResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.DeleteResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestAPIServer_ApplyAction(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)
	configsStructValue, _ := structpb.NewValue(map[string]interface{}{
		"replicas": "10",
	})

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.ApplyActionRequest
		want    *entropyv1beta1.ApplyActionResponse
		wantErr error
	}{
		{
			name: "ResourceNotFound",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ApplyAction(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrNotFound).Once()
				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.ApplyActionRequest{
				Urn:    "p-testdata-gl-testname-log",
				Action: "scale",
			},
			want:    nil,
			wantErr: status.Error(codes.NotFound, "requested entity not found"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ApplyAction(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(&resource.Resource{
						URN:       "p-testdata-gl-testname-log",
						Kind:      "log",
						Name:      "testname",
						Project:   "p-testdata-gl",
						Labels:    nil,
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Spec: resource.Spec{
							Configs: map[string]interface{}{
								"replicas": "10",
							},
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewApiServer(resourceService, nil)
			},
			request: &entropyv1beta1.ApplyActionRequest{
				Urn:    "p-testdata-gl-testname-log",
				Action: "scale",
				Params: configsStructValue,
			},
			want: &entropyv1beta1.ApplyActionResponse{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.setup(t)

			got, err := srv.ApplyAction(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}
