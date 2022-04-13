package handlersv1

//go:generate mockery --name=ModuleService -r --case underscore --with-expecter --structname ModuleService  --filename=module_service.go --output=./mocks
//go:generate mockery --name=ResourceService -r --case underscore --with-expecter --structname ResourceService  --filename=resource_service.go --output=./mocks
//go:generate mockery --name=ProviderService -r --case underscore --with-expecter --structname ProviderService  --filename=provider_service.go --output=./mocks

import (
	"context"
	"errors"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/resource"
)

var ErrInternal = status.Error(codes.Internal, "internal server error")

type ResourceService interface {
	GetResource(ctx context.Context, urn string) (*resource.Resource, error)
	ListResources(ctx context.Context, parent string, kind string) ([]*resource.Resource, error)
	CreateResource(ctx context.Context, res *resource.Resource) (*resource.Resource, error)
	UpdateResource(ctx context.Context, res *resource.Resource) (*resource.Resource, error)
	DeleteResource(ctx context.Context, urn string) error
}

type ModuleService interface {
	Act(ctx context.Context, r resource.Resource, action string, params map[string]interface{}) (map[string]interface{}, error)
	Log(ctx context.Context, r resource.Resource, filter map[string]string) (<-chan module.LogChunk, error)
	Sync(ctx context.Context, r resource.Resource) (*resource.Resource, error)
	Validate(ctx context.Context, r resource.Resource) error
}

type ProviderService interface {
	CreateProvider(ctx context.Context, res *provider.Provider) (*provider.Provider, error)
	ListProviders(ctx context.Context, parent string, kind string) ([]*provider.Provider, error)
}

type APIServer struct {
	entropyv1beta1.UnimplementedResourceServiceServer
	entropyv1beta1.UnimplementedProviderServiceServer

	resourceService ResourceService
	moduleService   ModuleService
	providerService ProviderService
}

func NewApiServer(resourceService ResourceService, moduleService ModuleService, providerService ProviderService) *APIServer {
	return &APIServer{
		resourceService: resourceService,
		moduleService:   moduleService,
		providerService: providerService,
	}
}

func (server APIServer) CreateResource(ctx context.Context, request *entropyv1beta1.CreateResourceRequest) (*entropyv1beta1.CreateResourceResponse, error) {
	res := resourceFromProto(request.Resource)
	res.URN = resource.GenerateURN(*res)

	err := server.validateResource(ctx, *res)
	if err != nil {
		return nil, err
	}

	createdResource, err := server.resourceService.CreateResource(ctx, res)
	if err != nil {
		if errors.Is(err, resource.ErrResourceAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		}
		return nil, ErrInternal
	}

	syncedResource, err := server.syncResource(ctx, *createdResource)
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
		if errors.Is(err, resource.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}
	res.Configs = request.GetConfigs().GetStructValue().AsMap()
	res.Status = resource.StatusPending

	err = server.validateResource(ctx, *res)
	if err != nil {
		return nil, err
	}
	updatedResource, err := server.resourceService.UpdateResource(ctx, res)
	if err != nil {
		return nil, err
	}
	syncedResource, err := server.syncResource(ctx, *updatedResource)
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
		if errors.Is(err, resource.ErrResourceNotFound) {
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
		if errors.Is(err, resource.ErrResourceNotFound) {
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
		if errors.Is(err, resource.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return nil, ErrInternal
	}
	action := request.GetAction()
	params := request.GetParams().GetStructValue().AsMap()
	resultConfig, err := server.moduleService.Act(ctx, *res, action, params)
	if err != nil {
		return nil, ErrInternal
	}
	res.Configs = resultConfig
	syncedResource, err := server.syncResource(ctx, *res)
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

func (server APIServer) GetLog(request *entropyv1beta1.GetLogRequest, stream entropyv1beta1.ResourceService_GetLogServer) error {
	ctx := stream.Context()
	res, err := server.resourceService.GetResource(ctx, request.GetUrn())
	if err != nil {
		if errors.Is(err, resource.ErrResourceNotFound) {
			return status.Error(codes.NotFound, "could not find resource with given urn")
		}
		return ErrInternal
	}
	logChunks, err := server.moduleService.Log(ctx, *res, request.GetFilter())
	if err != nil {
		return ErrInternal
	}
	for logChunk := range logChunks {
		err := stream.Send(&entropyv1beta1.GetLogResponse{
			Chunk: &entropyv1beta1.LogChunk{
				Data:   logChunk.Data,
				Labels: logChunk.Labels,
			},
		})
		if err != nil {
			return ErrInternal
		}
	}
	return nil
}

func (server APIServer) CreateProvider(ctx context.Context, request *entropyv1beta1.CreateProviderRequest) (*entropyv1beta1.CreateProviderResponse, error) {
	pro := providerFromProto(request.Provider)
	pro.URN = provider.GenerateURN(*pro)
	// TODO: add provider validation

	createdProvider, err := server.providerService.CreateProvider(ctx, pro)
	if err != nil {
		if errors.Is(err, provider.ErrProviderAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "provider already exists")
		}
		return nil, ErrInternal
	}

	responseProvider, err := providerToProto(createdProvider)
	if err != nil {
		return nil, ErrInternal
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
		return nil, ErrInternal
	}

	for _, pro := range providers {
		responseProvider, err := providerToProto(pro)
		if err != nil {
			return nil, ErrInternal
		}
		responseProviders = append(responseProviders, responseProvider)
	}

	response := entropyv1beta1.ListProvidersResponse{
		Providers: responseProviders,
	}
	return &response, nil
}

func (server APIServer) syncResource(ctx context.Context, updatedResource resource.Resource) (*resource.Resource, error) {
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

func (server APIServer) validateResource(ctx context.Context, res resource.Resource) error {
	err := server.moduleService.Validate(ctx, res)
	if err != nil {
		if errors.Is(err, module.ErrModuleNotFound) {
			return status.Errorf(codes.InvalidArgument, "failed to find module to deploy this kind")
		}
		if errors.Is(err, module.ErrModuleConfigParseFailed) {
			return status.Errorf(codes.InvalidArgument, "failed to parse configs")
		}
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	return nil
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
