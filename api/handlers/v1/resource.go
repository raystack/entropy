package handlersv1

import (
	"context"
	"errors"

	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/pkg/module"
	"github.com/odpf/entropy/pkg/resource"
	"github.com/odpf/entropy/store"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrInternal = status.Error(codes.Internal, "internal server error")

type APIServer struct {
	entropyv1beta1.UnimplementedResourceServiceServer
	resourceService resource.ServiceInterface
	moduleService   module.ServiceInterface
}

func NewApiServer(resourceService resource.ServiceInterface, moduleService module.ServiceInterface) *APIServer {
	return &APIServer{
		resourceService: resourceService,
		moduleService:   moduleService,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropyv1beta1.CreateResourceRequest) (*entropyv1beta1.CreateResourceResponse, error) {
	res := resourceFromProto(request.Resource)
	res.Urn = domain.GenerateResourceUrn(res)
	err := server.validateResource(ctx, res)
	if err != nil {
		return nil, err
	}
	createdResource, err := server.resourceService.CreateResource(ctx, res)
	if err != nil {
		if errors.Is(err, store.ResourceAlreadyExistsError) {
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		}
		return nil, ErrInternal
	}
	syncedResource, err := server.syncResource(ctx, createdResource)
	if err != nil {
		return nil, err
	}
	responseResource, err := resourceToProto(syncedResource)
	if err != nil {
		return nil, ErrInternal
	}
	response := entropyv1beta1.CreateResourceResponse{
		Resource: responseResource,
	}
	return &response, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropyv1beta1.UpdateResourceRequest) (*entropyv1beta1.UpdateResourceResponse, error) {
	res, err := server.resourceService.GetResource(ctx, request.GetUrn())
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}
	res.Configs = request.GetConfigs().GetStructValue().AsMap()
	res.Status = domain.ResourceStatusPending
	err = server.validateResource(ctx, res)
	if err != nil {
		return nil, err
	}
	updatedResource, err := server.resourceService.UpdateResource(ctx, res)
	if err != nil {
		return nil, err
	}
	syncedResource, err := server.syncResource(ctx, updatedResource)
	if err != nil {
		return nil, err
	}
	responseResource, err := resourceToProto(syncedResource)
	if err != nil {
		return nil, ErrInternal
	}
	response := entropyv1beta1.UpdateResourceResponse{
		Resource: responseResource,
	}
	return &response, nil
}

func (server APIServer) GetResource(ctx context.Context, request *entropyv1beta1.GetResourceRequest) (*entropyv1beta1.GetResourceResponse, error) {
	res, err := server.resourceService.GetResource(ctx, request.GetUrn())
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}
	responseResource, err := resourceToProto(res)
	if err != nil {
		return nil, ErrInternal
	}
	response := entropyv1beta1.GetResourceResponse{
		Resource: responseResource,
	}
	return &response, nil
}

func (server APIServer) ListResources(ctx context.Context, request *entropyv1beta1.ListResourcesRequest) (*entropyv1beta1.ListResourcesResponse, error) {
	var responseResources []*entropyv1beta1.Resource
	resources, err := server.resourceService.ListResources(ctx, request.GetParent(), request.GetKind())
	if err != nil {
		return nil, ErrInternal
	}
	for _, res := range resources {
		responseResource, err := resourceToProto(res)
		if err != nil {
			return nil, ErrInternal
		}
		responseResources = append(responseResources, responseResource)
	}
	if err != nil {
		return nil, ErrInternal
	}
	response := entropyv1beta1.ListResourcesResponse{
		Resources: responseResources,
	}
	return &response, nil
}

func (server APIServer) DeleteResource(ctx context.Context, request *entropyv1beta1.DeleteResourceRequest) (*entropyv1beta1.DeleteResourceResponse, error) {
	urn := request.GetUrn()
	_, err := server.resourceService.GetResource(ctx, urn)
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}

	err = server.resourceService.DeleteResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	response := entropyv1beta1.DeleteResourceResponse{}
	return &response, nil
}

func (server APIServer) ApplyAction(ctx context.Context, request *entropyv1beta1.ApplyActionRequest) (*entropyv1beta1.ApplyActionResponse, error) {
	res, err := server.resourceService.GetResource(ctx, request.GetUrn())
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}
	action := request.GetAction()
	params := request.GetParams().GetStructValue().AsMap()
	resultConfig, err := server.moduleService.Act(ctx, res, action, params)
	if err != nil {
		return nil, ErrInternal
	}
	res.Configs = resultConfig
	syncedResource, err := server.syncResource(ctx, res)
	if err != nil {
		return nil, err
	}
	responseResource, err := resourceToProto(syncedResource)
	if err != nil {
		return nil, ErrInternal
	}
	response := &entropyv1beta1.ApplyActionResponse{
		Resource: responseResource,
	}
	return response, nil
}

func (server APIServer) syncResource(ctx context.Context, updatedResource *domain.Resource) (*domain.Resource, error) {
	syncedResource, err := server.moduleService.Sync(ctx, updatedResource)
	if err != nil {
		return nil, ErrInternal
	}
	responseResource, err := server.resourceService.UpdateResource(ctx, syncedResource)
	if err != nil {
		return nil, ErrInternal
	}
	return responseResource, nil
}

func (server APIServer) validateResource(ctx context.Context, res *domain.Resource) error {
	err := server.moduleService.Validate(ctx, res)
	if err != nil {
		if errors.Is(err, store.ModuleNotFoundError) {
			return status.Errorf(codes.InvalidArgument, "failed to find module to deploy this kind")
		}
		if errors.Is(err, domain.ModuleConfigParseFailed) {
			return status.Errorf(codes.InvalidArgument, "failed to parse configs")
		}
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	return nil
}

func resourceToProto(res *domain.Resource) (*entropyv1beta1.Resource, error) {
	conf, err := structpb.NewValue(res.Configs)
	if err != nil {
		return nil, err
	}
	return &entropyv1beta1.Resource{
		Urn:       res.Urn,
		Name:      res.Name,
		Parent:    res.Parent,
		Kind:      res.Kind,
		Configs:   conf,
		Labels:    res.Labels,
		Status:    resourceStatusToProto(string(res.Status)),
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}, nil
}

func resourceStatusToProto(status string) entropyv1beta1.Resource_Status {
	if resourceStatus, ok := entropyv1beta1.Resource_Status_value[status]; ok {
		return entropyv1beta1.Resource_Status(resourceStatus)
	}
	return entropyv1beta1.Resource_STATUS_UNSPECIFIED
}

func resourceFromProto(res *entropyv1beta1.Resource) *domain.Resource {
	return &domain.Resource{
		Urn:     res.GetUrn(),
		Name:    res.GetName(),
		Parent:  res.GetParent(),
		Kind:    res.GetKind(),
		Configs: res.GetConfigs().GetStructValue().AsMap(),
		Labels:  res.GetLabels(),
	}
}
