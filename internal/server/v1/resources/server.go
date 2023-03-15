package resources

//go:generate mockery --name=ResourceService -r --case underscore --with-expecter --structname ResourceService  --filename=resource_service.go --output=../mocks

import (
	"context"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/internal/server/serverutils"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

type ResourceService interface {
	GetResource(ctx context.Context, urn string) (*resource.Resource, error)
	ListResources(ctx context.Context, filter resource.Filter) ([]resource.Resource, error)
	CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error)
	UpdateResource(ctx context.Context, urn string, req resource.UpdateRequest) (*resource.Resource, error)
	DeleteResource(ctx context.Context, urn string) error

	ApplyAction(ctx context.Context, urn string, action module.ActionRequest) (*resource.Resource, error)
	GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan module.LogChunk, error)

	GetRevisions(ctx context.Context, selector resource.RevisionsSelector) ([]resource.Revision, error)
}

type APIServer struct {
	entropyv1beta1.UnimplementedResourceServiceServer
	resourceSvc ResourceService
}

func NewAPIServer(resourceService ResourceService) *APIServer {
	return &APIServer{
		resourceSvc: resourceService,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropyv1beta1.CreateResourceRequest) (*entropyv1beta1.CreateResourceResponse, error) {
	res, err := resourceFromProto(request.Resource)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	result, err := server.resourceSvc.CreateResource(ctx, *res)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	responseResource, err := resourceToProto(*result)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.CreateResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) UpdateResource(ctx context.Context, request *entropyv1beta1.UpdateResourceRequest) (*entropyv1beta1.UpdateResourceResponse, error) {
	newSpec, err := resourceSpecFromProto(request.GetNewSpec())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	updateRequest := resource.UpdateRequest{
		Spec:   *newSpec,
		Labels: request.Labels,
	}

	res, err := server.resourceSvc.UpdateResource(ctx, request.GetUrn(), updateRequest)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	responseResource, err := resourceToProto(*res)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.UpdateResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) GetResource(ctx context.Context, request *entropyv1beta1.GetResourceRequest) (*entropyv1beta1.GetResourceResponse, error) {
	res, err := server.resourceSvc.GetResource(ctx, request.GetUrn())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	responseResource, err := resourceToProto(*res)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.GetResourceResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) ListResources(ctx context.Context, request *entropyv1beta1.ListResourcesRequest) (*entropyv1beta1.ListResourcesResponse, error) {
	filter := resource.Filter{
		Kind:    request.GetKind(),
		Project: request.GetProject(),
		Labels:  nil,
	}

	resources, err := server.resourceSvc.ListResources(ctx, filter)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	var responseResources []*entropyv1beta1.Resource
	for _, res := range resources {
		responseResource, err := resourceToProto(res)
		if err != nil {
			return nil, serverutils.ToRPCError(err)
		}
		responseResources = append(responseResources, responseResource)
	}

	return &entropyv1beta1.ListResourcesResponse{
		Resources: responseResources,
	}, nil
}

func (server APIServer) DeleteResource(ctx context.Context, request *entropyv1beta1.DeleteResourceRequest) (*entropyv1beta1.DeleteResourceResponse, error) {
	err := server.resourceSvc.DeleteResource(ctx, request.GetUrn())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.DeleteResourceResponse{}, nil
}

func (server APIServer) ApplyAction(ctx context.Context, request *entropyv1beta1.ApplyActionRequest) (*entropyv1beta1.ApplyActionResponse, error) {
	paramsJSON, err := request.GetParams().GetStructValue().MarshalJSON()
	if err != nil {
		return nil, err
	}

	action := module.ActionRequest{
		Name:   request.GetAction(),
		Params: paramsJSON,
		Labels: request.Labels,
	}

	updatedRes, err := server.resourceSvc.ApplyAction(ctx, request.GetUrn(), action)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	responseResource, err := resourceToProto(*updatedRes)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.ApplyActionResponse{
		Resource: responseResource,
	}, nil
}

func (server APIServer) GetLog(request *entropyv1beta1.GetLogRequest, stream entropyv1beta1.ResourceService_GetLogServer) error {
	ctx := stream.Context()

	logStream, err := server.resourceSvc.GetLog(ctx, request.GetUrn(), request.GetFilter())
	if err != nil {
		return serverutils.ToRPCError(err)
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
				return serverutils.ToRPCError(err)
			}
		}
	}
}

func (server APIServer) GetResourceRevisions(ctx context.Context, request *entropyv1beta1.GetResourceRevisionsRequest) (*entropyv1beta1.GetResourceRevisionsResponse, error) {
	revisions, err := server.resourceSvc.GetRevisions(ctx, resource.RevisionsSelector{URN: request.GetUrn()})
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	var responseRevisions []*entropyv1beta1.ResourceRevision
	for _, res := range revisions {
		responseRevision, err := revisionToProto(res)
		if err != nil {
			return nil, serverutils.ToRPCError(err)
		}

		responseRevisions = append(responseRevisions, responseRevision)
	}

	return &entropyv1beta1.GetResourceRevisionsResponse{
		Revisions: responseRevisions,
	}, nil
}
