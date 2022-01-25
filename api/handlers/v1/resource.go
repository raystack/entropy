package handlersv1

import (
	"context"
	"errors"
	"github.com/odpf/entropy/container"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
	entropy "go.buf.build/odpf/gwv/whoabhisheksah/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type APIServer struct {
	entropy.UnimplementedResourceServiceServer
	container *container.Container
}

func NewApiServer(container *container.Container) *APIServer {
	return &APIServer{
		container: container,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropy.CreateResourceRequest) (*entropy.CreateResourceResponse, error) {
	res := resourceFromProto(request.Resource)
	createdResource, err := server.container.ResourceService.CreateResource(ctx, res)
	if err != nil {
		if errors.Is(err, store.ResourceAlreadyExistsError) {
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create resource in db")
	}
	createdResponse, err := resourceToProto(createdResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize resource")
	}
	response := entropy.CreateResourceResponse{
		Resource: createdResponse,
	}
	return &response, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropy.UpdateResourceRequest) (*entropy.UpdateResourceResponse, error) {
	updatedResource, err := server.container.ResourceService.UpdateResource(ctx, request.GetUrn(), request.GetConfigs().GetStructValue().AsMap())
	if err != nil {
		if errors.Is(err, store.ResourceNotFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update resource in db")
	}
	updatedResponse, err := resourceToProto(updatedResource)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize resource")
	}
	response := entropy.UpdateResourceResponse{
		Resource: updatedResponse,
	}
	return &response, nil
}

func resourceToProto(res *domain.Resource) (*entropy.Resource, error) {
	conf, err := structpb.NewValue(res.Configs)
	if err != nil {
		return nil, err
	}
	return &entropy.Resource{
		Urn:       res.Urn,
		Name:      res.Name,
		Parent:    res.Parent,
		Kind:      res.Kind,
		Configs:   conf,
		Labels:    res.Labels,
		Status:    res.Status,
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}, nil
}

func resourceFromProto(res *entropy.Resource) *domain.Resource {
	return &domain.Resource{
		Urn:     res.GetUrn(),
		Name:    res.GetName(),
		Parent:  res.GetParent(),
		Kind:    res.GetKind(),
		Configs: res.GetConfigs().GetStructValue().AsMap(),
		Labels:  res.GetLabels(),
	}
}
