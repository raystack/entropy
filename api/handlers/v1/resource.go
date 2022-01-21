package handlersv1

import (
	"context"
	"errors"
	"github.com/odpf/entropy/domain/model"
	"github.com/odpf/entropy/domain/resource"
	"github.com/odpf/entropy/service"
	entropy "go.buf.build/odpf/gwv/whoabhisheksah/proton/odpf/entropy/v1beta1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type APIServer struct {
	entropy.UnimplementedResourceServiceServer
	container *service.Container
}

func NewApiServer(container *service.Container) *APIServer {
	return &APIServer{
		container: container,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropy.CreateResourceRequest) (*entropy.CreateResourceResponse, error) {
	res := model.ResourceFromProto(request.Resource)
	res.Urn = model.GenerateResourceUrn(res)
	res.Status = "PENDING"
	err := server.container.ResourceRepository.Create(res)
	if err != nil {
		if errors.Is(err, resource.ResourceAlreadyExistsError) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create resource in db")
	}
	createdResource, err := server.container.ResourceRepository.GetByURN(res.Urn)
	if err != nil {
		if errors.Is(err, resource.NoResourceFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get resource from db")
	}

	resource, err := model.ResourceToProto(createdResource)
	if err != nil {
		return nil, err
	}

	response := entropy.CreateResourceResponse{
		Resource: resource,
	}
	return &response, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropy.UpdateResourceRequest) (*entropy.UpdateResourceResponse, error) {
	updatePayload := &model.Resource{
		Urn:     request.GetUrn(),
		Configs: request.GetConfigs().GetStructValue().AsMap(),
	}
	res, err := server.container.ResourceRepository.GetByURN(updatePayload.Urn)
	if err != nil {
		if errors.Is(err, resource.NoResourceFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get resource from db")
	}
	res.Configs = updatePayload.Configs
	err = server.container.ResourceRepository.Update(res)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update resource in db")
	}
	updatedRes, err := server.container.ResourceRepository.GetByURN(res.Urn)
	updatedResponse, err := model.ResourceToProto(updatedRes)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get resource from db")
	}
	response := entropy.UpdateResourceResponse{
		Resource: updatedResponse,
	}
	return &response, nil
}
