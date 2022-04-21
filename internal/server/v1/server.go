package handlersv1

//go:generate mockery --name=ResourceService -r --case underscore --with-expecter --structname ResourceService  --filename=resource_service.go --output=./mocks
//go:generate mockery --name=ProviderService -r --case underscore --with-expecter --structname ProviderService  --filename=provider_service.go --output=./mocks

import (
	"context"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

type ResourceService interface {
	GetResource(ctx context.Context, urn string) (*resource.Resource, error)
	ListResources(ctx context.Context, parent string, kind string) ([]resource.Resource, error)
	CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error)
	UpdateResource(ctx context.Context, urn string, updates resource.Updates) (*resource.Resource, error)
	DeleteResource(ctx context.Context, urn string) error

	ApplyAction(ctx context.Context, urn string, action resource.Action) (*resource.Resource, error)
	GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan resource.LogChunk, error)
}

type ProviderService interface {
	CreateProvider(ctx context.Context, res provider.Provider) (*provider.Provider, error)
	ListProviders(ctx context.Context, parent string, kind string) ([]*provider.Provider, error)
}

type APIServer struct {
	entropyv1beta1.UnimplementedResourceServiceServer
	entropyv1beta1.UnimplementedProviderServiceServer

	resourceService ResourceService
	providerService ProviderService
}

func NewApiServer(resourceService ResourceService, providerService ProviderService) *APIServer {
	return &APIServer{
		resourceService: resourceService,
		providerService: providerService,
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
	updates := resource.Updates{
		Configs: request.GetConfigs().GetStructValue().AsMap(),
	}

	res, err := server.resourceService.UpdateResource(ctx, request.GetUrn(), updates)
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
	action := resource.Action{
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

		case chunk := <-logStream:
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

func (server APIServer) CreateProvider(ctx context.Context, request *entropyv1beta1.CreateProviderRequest) (*entropyv1beta1.CreateProviderResponse, error) {
	pro := providerFromProto(request.Provider)
	pro.URN = provider.GenerateURN(*pro)
	// TODO: add provider validation

	createdProvider, err := server.providerService.CreateProvider(ctx, *pro)
	if err != nil {
		return nil, generateRPCErr(err)
	}

	responseProvider, err := providerToProto(createdProvider)
	if err != nil {
		return nil, generateRPCErr(err)
	}
	response := entropyv1beta1.CreateProviderResponse{
		Provider: responseProvider,
	}
	return &response, nil
}

func (server APIServer) ListProviders(ctx context.Context, request *entropyv1beta1.ListProvidersRequest) (*entropyv1beta1.ListProvidersResponse, error) {
	var responseProviders []*entropyv1beta1.Provider
	providers, err := server.providerService.ListProviders(ctx, request.GetParent(), request.GetKind())
	if err != nil {
		return nil, generateRPCErr(err)
	}

	for _, pro := range providers {
		responseProvider, err := providerToProto(pro)
		if err != nil {
			return nil, generateRPCErr(err)
		}
		responseProviders = append(responseProviders, responseProvider)
	}

	response := entropyv1beta1.ListProvidersResponse{
		Providers: responseProviders,
	}
	return &response, nil
}

func resourceToProto(res *resource.Resource) (*entropyv1beta1.Resource, error) {
	conf, err := structpb.NewValue(res.Configs)
	if err != nil {
		return nil, err
	}
	return &entropyv1beta1.Resource{
		Urn:       res.URN,
		Name:      res.Name,
		Parent:    res.Parent,
		Kind:      res.Kind,
		Configs:   conf,
		Labels:    res.Labels,
		Providers: resourceProvidersToProto(res.Providers),
		Status:    resourceStatusToProto(string(res.Status)),
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}, nil
}

func resourceProvidersToProto(ps []resource.ProviderSelector) []*entropyv1beta1.ProviderSelector {
	var providerSelectors []*entropyv1beta1.ProviderSelector

	for _, p := range ps {
		selector := &entropyv1beta1.ProviderSelector{
			Urn:    p.URN,
			Target: p.Target,
		}
		providerSelectors = append(providerSelectors, selector)
	}
	return providerSelectors
}

func providerToProto(pro *provider.Provider) (*entropyv1beta1.Provider, error) {
	conf, err := structpb.NewValue(pro.Configs)
	if err != nil {
		return nil, err
	}
	return &entropyv1beta1.Provider{
		Urn:       pro.URN,
		Name:      pro.Name,
		Parent:    pro.Parent,
		Kind:      pro.Kind,
		Configs:   conf,
		Labels:    pro.Labels,
		CreatedAt: timestamppb.New(pro.CreatedAt),
		UpdatedAt: timestamppb.New(pro.UpdatedAt),
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
		URN:       res.GetUrn(),
		Name:      res.GetName(),
		Parent:    res.GetParent(),
		Kind:      res.GetKind(),
		Configs:   res.GetConfigs().GetStructValue().AsMap(),
		Labels:    res.GetLabels(),
		Providers: providerSelectorFromProto(res.GetProviders()),
	}
}

func providerSelectorFromProto(ps []*entropyv1beta1.ProviderSelector) []resource.ProviderSelector {
	var providerSelectors []resource.ProviderSelector

	for _, p := range ps {
		selector := resource.ProviderSelector{
			URN:    p.GetUrn(),
			Target: p.GetTarget(),
		}
		providerSelectors = append(providerSelectors, selector)
	}
	return providerSelectors
}

func providerFromProto(pro *entropyv1beta1.Provider) *provider.Provider {
	return &provider.Provider{
		URN:     pro.GetUrn(),
		Name:    pro.GetName(),
		Parent:  pro.GetParent(),
		Kind:    pro.GetKind(),
		Configs: pro.GetConfigs().GetStructValue().AsMap(),
		Labels:  pro.GetLabels(),
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
