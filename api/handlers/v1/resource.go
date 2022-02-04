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
	createdResource, err := server.resourceService.CreateResource(ctx, res)
	if err != nil {
		if errors.Is(err, store.ResourceAlreadyExistsError) {
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create resource in db")
	}
	err = server.moduleService.TriggerSync(ctx, createdResource.Urn)
	if err != nil {
		if errors.Is(err, store.ModuleNotFoundError) {
			return nil, status.Errorf(codes.NotFound, "failed to find module to deploy this kind")
		}
		return nil, status.Error(codes.Internal, "failed to sync created resource")
	}
	createdResponse, err := resourceToProto(createdResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize resource")
	}
	response := entropyv1beta1.CreateResourceResponse{
		Resource: createdResponse,
	}
	return &response, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropyv1beta1.UpdateResourceRequest) (*entropyv1beta1.UpdateResourceResponse, error) {
	updatedResource, err := server.resourceService.UpdateResource(ctx, request.GetUrn(), request.GetConfigs().GetStructValue().AsMap())
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, status.Error(codes.Internal, "failed to update resource in db")
	}
	err = server.moduleService.TriggerSync(ctx, updatedResource.Urn)
	if err != nil {
		if errors.Is(err, store.ModuleNotFoundError) {
			return nil, status.Errorf(codes.NotFound, "failed to find module to deploy this kind")
		}
		return nil, status.Error(codes.Internal, "failed to sync updated resource")
	}
	updatedResponse, err := resourceToProto(updatedResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize resource")
	}
	response := entropyv1beta1.UpdateResourceResponse{
		Resource: updatedResponse,
	}
	return &response, nil
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
