package modules

import (
	"encoding/json"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/module"
)

func moduleToProto(mod module.Module) (*entropyv1beta1.Module, error) {
	spec, err := moduleSpecToProto(mod.Spec)
	if err != nil {
		return nil, err
	}

	return &entropyv1beta1.Module{
		Urn:       mod.URN,
		Name:      mod.Name,
		Spec:      spec,
		Project:   mod.Project,
		CreatedAt: timestamppb.New(mod.CreatedAt),
		UpdatedAt: timestamppb.New(mod.UpdatedAt),
	}, nil
}

func moduleSpecToProto(spec module.Spec) (*entropyv1beta1.ModuleSpec, error) {
	conf := structpb.Value{}
	if err := json.Unmarshal(spec.Configs, &conf); err != nil {
		return nil, err
	}

	var loaderType entropyv1beta1.ModuleSpec_LoaderType
	switch spec.Loader {
	case "go":
		loaderType = entropyv1beta1.ModuleSpec_LOADER_TYPE_GO

	default:
		loaderType = entropyv1beta1.ModuleSpec_LOADER_TYPE_UNSPECIFIED
	}

	return &entropyv1beta1.ModuleSpec{
		Path:    spec.Path,
		Loader:  loaderType,
		Configs: &conf,
	}, nil
}

func moduleFromProto(res *entropyv1beta1.Module) (*module.Module, error) {
	spec, err := moduleSpecFromProto(res.Spec)
	if err != nil {
		return nil, err
	}

	return &module.Module{
		URN:       res.GetUrn(),
		Name:      res.GetName(),
		Spec:      *spec,
		Project:   res.GetProject(),
		CreatedAt: res.GetCreatedAt().AsTime(),
		UpdatedAt: res.GetUpdatedAt().AsTime(),
	}, nil
}

func moduleSpecFromProto(spec *entropyv1beta1.ModuleSpec) (*module.Spec, error) {
	confJSON, err := spec.GetConfigs().MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &module.Spec{
		Path:    spec.Path,
		Loader:  spec.Loader.String(),
		Configs: confJSON,
	}, nil
}
