package handlersv1

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/entropy/domain/model"
	"github.com/odpf/entropy/domain/resource"
	"github.com/odpf/entropy/service"
	entropy "go.buf.build/odpf/gwv/rohilsurana/proton/odpf/entropy/v1beta1"
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
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()
	err := server.container.ResourceRepository.Create(res)
	if err != nil {
		if errors.Is(err, resource.ResourceAlreadyExistsError) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, err
	}
	createdResource, err := server.container.ResourceRepository.GetByURN(res.Urn)
	if err != nil {
		if errors.Is(err, resource.NoResourceFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
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
	response := entropy.UpdateResourceResponse{}
	return &response, nil
}
