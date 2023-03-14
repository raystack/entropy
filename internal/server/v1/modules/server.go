package modules

//go:generate mockery --name=ModuleService -r --case underscore --with-expecter --structname ModuleService  --filename=module_service.go --output=../mocks

import (
	"context"
	"encoding/json"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/internal/server/serverutils"
)

type ModuleService interface {
	GetModule(ctx context.Context, urn string) (*module.Module, error)
	ListModules(ctx context.Context, project string) ([]module.Module, error)
	CreateModule(ctx context.Context, mod module.Module) (*module.Module, error)
	UpdateModule(ctx context.Context, urn string, newConfigs json.RawMessage) (*module.Module, error)
	DeleteModule(ctx context.Context, urn string) error
}

type APIServer struct {
	entropyv1beta1.UnimplementedModuleServiceServer

	moduleService ModuleService
}

func NewAPIServer(moduleService ModuleService) *APIServer {
	return &APIServer{
		moduleService: moduleService,
	}
}

func (srv *APIServer) ListModules(ctx context.Context, request *entropyv1beta1.ListModulesRequest) (*entropyv1beta1.ListModulesResponse, error) {
	mods, err := srv.moduleService.ListModules(ctx, request.GetProject())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	var responseModules []*entropyv1beta1.Module
	for _, mod := range mods {
		rm, err := moduleToProto(mod)
		if err != nil {
			return nil, serverutils.ToRPCError(err)
		}
		responseModules = append(responseModules, rm)
	}

	return &entropyv1beta1.ListModulesResponse{
		Modules: responseModules,
	}, nil
}

func (srv *APIServer) GetModule(ctx context.Context, request *entropyv1beta1.GetModuleRequest) (*entropyv1beta1.GetModuleResponse, error) {
	mod, err := srv.moduleService.GetModule(ctx, request.GetUrn())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	resp, err := moduleToProto(*mod)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}
	return &entropyv1beta1.GetModuleResponse{Module: resp}, nil
}

func (srv *APIServer) CreateModule(ctx context.Context, request *entropyv1beta1.CreateModuleRequest) (*entropyv1beta1.CreateModuleResponse, error) {
	mod, err := moduleFromProto(request.GetModule())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	createdMod, err := srv.moduleService.CreateModule(ctx, *mod)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	resp, err := moduleToProto(*createdMod)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}
	return &entropyv1beta1.CreateModuleResponse{Module: resp}, nil
}

func (srv *APIServer) UpdateModule(ctx context.Context, request *entropyv1beta1.UpdateModuleRequest) (*entropyv1beta1.UpdateModuleResponse, error) {
	newConfigs, err := getConfigsAsRawJSON(request)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	updatedMod, err := srv.moduleService.UpdateModule(ctx, request.GetUrn(), newConfigs)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	resp, err := moduleToProto(*updatedMod)
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}

	return &entropyv1beta1.UpdateModuleResponse{Module: resp}, nil
}

func (srv *APIServer) DeleteModule(ctx context.Context, request *entropyv1beta1.DeleteModuleRequest) (*entropyv1beta1.DeleteModuleResponse, error) {
	err := srv.moduleService.DeleteModule(ctx, request.GetUrn())
	if err != nil {
		return nil, serverutils.ToRPCError(err)
	}
	return &entropyv1beta1.DeleteModuleResponse{}, nil
}
