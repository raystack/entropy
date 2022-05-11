package handlersv1

//go:generate mockery --name=ResourceService -r --case underscore --with-expecter --structname ResourceService  --filename=resource_service.go --output=./mocks

import (
	"context"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

type ResourceService interface {
	GetResource(ctx context.Context, urn string) (*resource.Resource, error)
	ListResources(ctx context.Context, parent string, kind string) ([]resource.Resource, error)
	CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error)
	UpdateResource(ctx context.Context, urn string, newSpec resource.Spec) (*resource.Resource, error)
	DeleteResource(ctx context.Context, urn string) error

	ApplyAction(ctx context.Context, urn string, action module.ActionRequest) (*resource.Resource, error)
	GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan module.LogChunk, error)
}

type APIServer struct {
	entropyv1beta1.UnimplementedResourceServiceServer

	resourceService ResourceService
}

func NewApiServer(resourceService ResourceService) *APIServer {
	return &APIServer{
		resourceService: resourceService,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropyv1beta1.CreateResourceRequest) (*entropyv1beta1.CreateResourceResponse, error) {
	res := resourceFromProto(request.Resource)

	result, err := server.resourceService.CreateResource(ctx, *res)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	responseResource, err := resourceToProto(result)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	return &entropyv1beta1.CreateResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropyv1beta1.UpdateResourceRequest) (*entropyv1beta1.UpdateResourceResponse, error) {
	newSpec := resource.Spec{
		Configs: request.GetConfigs().GetStructValue().AsMap(),
	}

	res, err := server.resourceService.UpdateResource(ctx, request.GetUrn(), newSpec)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	responseResource, err := resourceToProto(res)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	return &entropyv1beta1.UpdateResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) GetResource(ctx context.Context, request *entropyv1beta1.GetResourceRequest) (*entropyv1beta1.GetResourceResponse, error) {
	res, err := server.resourceService.GetResource(ctx, request.GetUrn())
	if err != nil {
		return nil, generateRPCErr(err)
	}

	responseResource, err := resourceToProto(res)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	return &entropyv1beta1.GetResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) ListResources(ctx context.Context, request *entropyv1beta1.ListResourcesRequest) (*entropyv1beta1.ListResourcesResponse, error) {
	resources, err := server.resourceService.ListResources(ctx, request.GetParent(), request.GetKind())
	if err != nil {
		return nil, generateRPCErr(err)
	}

	var responseResources []*entropyv1beta1.Resource
	for _, res := range resources {
		responseResource, err := resourceToProto(&res)
		if err != nil {
			return nil, generateRPCErr(err)
		}
		responseResources = append(responseResources, responseResource)
	}

	return &entropyv1beta1.ListResourcesResponse{
		Resources: responseResources,
	}, nil
}

func (server APIServer) DeleteResource(ctx context.Context, request *entropyv1beta1.DeleteResourceRequest) (*entropyv1beta1.DeleteResourceResponse, error) {
	err := server.resourceService.DeleteResource(ctx, request.GetUrn())
	if err != nil {
		return nil, generateRPCErr(err)
	}

	return &entropyv1beta1.DeleteResourceResponse{}, nil
}

func (server APIServer) ApplyAction(ctx context.Context, request *entropyv1beta1.ApplyActionRequest) (*entropyv1beta1.ApplyActionResponse, error) {
	action := module.ActionRequest{
		Name:   request.GetAction(),
		Params: request.GetParams().GetStructValue().AsMap(),
	}

	updatedRes, err := server.resourceService.ApplyAction(ctx, request.GetUrn(), action)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	responseResource, err := resourceToProto(updatedRes)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	return &entropyv1beta1.ApplyActionResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) GetLog(request *entropyv1beta1.GetLogRequest, stream entropyv1beta1.ResourceService_GetLogServer) error {
	ctx := stream.Context()

	logStream, err := server.resourceService.GetLog(ctx, request.GetUrn(), request.GetFilter())
	if err != nil {
		return generateRPCErr(err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case chunk, open := <-logStream:
			if !open {
				return nil
			}

			resp := &entropyv1beta1.GetLogResponse{
				Chunk: &entropyv1beta1.LogChunk{
					Data:   chunk.Data,
					Labels: chunk.Labels,
				},
			}

			if err := stream.Send(resp); err != nil {
				return generateRPCErr(err)
			}
		}
	}
}

func resourceToProto(res *resource.Resource) (*entropyv1beta1.Resource, error) {
	conf, err := structpb.NewValue(res.Spec.Configs)
	if err != nil {
		return nil, err
	}
	return &entropyv1beta1.Resource{
		Urn:       res.URN,
		Name:      res.Name,
		Parent:    res.Project,
		Kind:      res.Kind,
		Configs:   conf,
		Labels:    res.Labels,
		Providers: nil,
		Status:    resourceStatusToProto(string(res.State.Status)),
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

func resourceFromProto(res *entropyv1beta1.Resource) *resource.Resource {
	return &resource.Resource{
		URN:     res.GetUrn(),
		Kind:    res.GetKind(),
		Name:    res.GetName(),
		Project: res.GetParent(),
		Spec: resource.Spec{
			Configs: res.GetConfigs().GetStructValue().AsMap(),
		},
		Labels: res.GetLabels(),
	}
}

func generateRPCErr(e error) error {
	err := errors.E(e)

	var code codes.Code
	switch {
	case errors.Is(err, errors.ErrNotFound):
		code = codes.NotFound

	case errors.Is(err, errors.ErrConflict):
		code = codes.AlreadyExists

	case errors.Is(err, errors.ErrInvalid):
		code = codes.InvalidArgument

	default:
		code = codes.Internal
	}
	return status.Error(code, err.Error())
}
