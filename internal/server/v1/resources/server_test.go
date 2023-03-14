package resources

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/internal/server/v1/mocks"
	"github.com/goto/entropy/pkg/errors"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

func TestAPIServer_CreateResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()

	configsStructValue := &structpb.Value{}
	require.NoError(t, json.Unmarshal([]byte(`{"replicas": "10"}`), &configsStructValue))

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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					CreateResource(mock.Anything, mock.Anything).
					Return(nil, errors.ErrConflict).Once()
				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Project: "p-testdata-gl",
					Kind:    "log",
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					Labels: nil,
				},
			},
			want:    nil,
			wantErr: status.Error(codes.AlreadyExists, "an entity with conflicting identifier exists"),
		},
		{
			name: "InvalidRequest",
			setup: func(t *testing.T) *APIServer {
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					CreateResource(mock.Anything, mock.Anything).
					Return(nil, errors.ErrInvalid).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Project: "p-testdata-gl",
					Kind:    "log",
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					Labels: nil,
				},
			},
			want:    nil,
			wantErr: status.Errorf(codes.InvalidArgument, "request is not valid"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				t.Helper()
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
							Configs: []byte(`{"replicas": "10"}`),
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.CreateResourceRequest{
				Resource: &entropyv1beta1.Resource{
					Name:    "testname",
					Project: "p-testdata-gl",
					Kind:    "log",
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					Labels: nil,
				},
			},
			want: &entropyv1beta1.CreateResourceResponse{
				Resource: &entropyv1beta1.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Kind:      "log",
					Name:      "testname",
					Labels:    nil,
					Project:   "p-testdata-gl",
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(createdAt),
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					State: &entropyv1beta1.ResourceState{
						Status: entropyv1beta1.ResourceState_STATUS_PENDING,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.CreateResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAPIServer_UpdateResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)

	configsStructValue := &structpb.Value{}
	require.NoError(t, json.Unmarshal([]byte(`{"replicas": "10"}`), &configsStructValue))

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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					UpdateResource(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrNotFound).Once()
				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn: "p-testdata-gl-testname-log",
				NewSpec: &entropyv1beta1.ResourceSpec{
					Configs: configsStructValue,
				},
			},
			want:    nil,
			wantErr: status.Error(codes.NotFound, "requested entity not found"),
		},
		{
			name: "InvalidRequest",
			setup: func(t *testing.T) *APIServer {
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					UpdateResource(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrInvalid).Once()
				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn: "p-testdata-gl-testname-log",
				NewSpec: &entropyv1beta1.ResourceSpec{
					Configs: configsStructValue,
				},
			},
			want:    nil,
			wantErr: status.Errorf(codes.InvalidArgument, "request is not valid"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				t.Helper()
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
							Configs: []byte(`{"replicas": "10"}`),
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.UpdateResourceRequest{
				Urn: "p-testdata-gl-testname-log",
				NewSpec: &entropyv1beta1.ResourceSpec{
					Configs: configsStructValue,
				},
			},
			want: &entropyv1beta1.UpdateResourceResponse{
				Resource: &entropyv1beta1.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Kind:      "log",
					Name:      "testname",
					Labels:    nil,
					Project:   "p-testdata-gl",
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(updatedAt),
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					State: &entropyv1beta1.ResourceState{
						Status: entropyv1beta1.ResourceState_STATUS_PENDING,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.UpdateResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAPIServer_GetResource(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)

	configsStructValue := &structpb.Value{}
	require.NoError(t, json.Unmarshal([]byte(`{"replicas": "10"}`), &configsStructValue))

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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					GetResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(nil, errors.ErrNotFound).Once()
				return NewAPIServer(resourceService)
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
				t.Helper()
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
							Configs: []byte(`{"replicas": "10"}`),
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.GetResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want: &entropyv1beta1.GetResourceResponse{
				Resource: &entropyv1beta1.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Kind:      "log",
					Name:      "testname",
					Labels:    nil,
					Project:   "p-testdata-gl",
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(updatedAt),
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					State: &entropyv1beta1.ResourceState{
						Status: entropyv1beta1.ResourceState_STATUS_PENDING,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.GetResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAPIServer_ListResources(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)

	configsStructValue := &structpb.Value{}
	require.NoError(t, json.Unmarshal([]byte(`{"replicas": "10"}`), &configsStructValue))

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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ListResources(mock.Anything, mock.Anything).
					Return(nil, errors.New("failed")).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.ListResourcesRequest{
				Project: "p-testdata-gl",
				Kind:    "log",
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "some unexpected error occurred"),
		},
		{
			name: "Success",
			setup: func(t *testing.T) *APIServer {
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ListResources(mock.Anything, mock.Anything).
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
								Configs: []byte(`{"replicas": "10"}`),
							},
							State: resource.State{
								Status: resource.StatusPending,
							},
						},
					}, nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.ListResourcesRequest{
				Project: "p-testdata-gl",
				Kind:    "log",
			},
			want: &entropyv1beta1.ListResourcesResponse{
				Resources: []*entropyv1beta1.Resource{
					{
						Urn:       "p-testdata-gl-testname-log",
						Kind:      "log",
						Name:      "testname",
						Labels:    nil,
						Project:   "p-testdata-gl",
						CreatedAt: timestamppb.New(createdAt),
						UpdatedAt: timestamppb.New(updatedAt),
						Spec: &entropyv1beta1.ResourceSpec{
							Configs: configsStructValue,
						},
						State: &entropyv1beta1.ResourceState{
							Status: entropyv1beta1.ResourceState_STATUS_PENDING,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.ListResources(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					DeleteResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(errors.ErrNotFound).Once()
				return NewAPIServer(resourceService)
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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					DeleteResource(mock.Anything, "p-testdata-gl-testname-log").
					Return(nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.DeleteResourceRequest{
				Urn: "p-testdata-gl-testname-log",
			},
			want: &entropyv1beta1.DeleteResourceResponse{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.DeleteResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAPIServer_ApplyAction(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Minute)

	configsStructValue := &structpb.Value{}
	require.NoError(t, json.Unmarshal([]byte(`{"replicas": "10"}`), &configsStructValue))

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
				t.Helper()
				resourceService := &mocks.ResourceService{}
				resourceService.EXPECT().
					ApplyAction(mock.Anything, "p-testdata-gl-testname-log", mock.Anything).
					Return(nil, errors.ErrNotFound).Once()
				return NewAPIServer(resourceService)
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
				t.Helper()
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
							Configs: []byte(`{"replicas": "10"}`),
						},
						State: resource.State{
							Status: resource.StatusPending,
						},
					}, nil).Once()

				return NewAPIServer(resourceService)
			},
			request: &entropyv1beta1.ApplyActionRequest{
				Urn:    "p-testdata-gl-testname-log",
				Action: "scale",
				Params: configsStructValue,
			},
			want: &entropyv1beta1.ApplyActionResponse{
				Resource: &entropyv1beta1.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Kind:      "log",
					Name:      "testname",
					Labels:    nil,
					Project:   "p-testdata-gl",
					CreatedAt: timestamppb.New(createdAt),
					UpdatedAt: timestamppb.New(updatedAt),
					Spec: &entropyv1beta1.ResourceSpec{
						Configs: configsStructValue,
					},
					State: &entropyv1beta1.ResourceState{
						Status: entropyv1beta1.ResourceState_STATUS_PENDING,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.ApplyAction(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
