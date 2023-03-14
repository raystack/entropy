package modules

import (
	"encoding/json"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/pkg/errors"
)

func moduleToProto(mod module.Module) (*entropyv1beta1.Module, error) {
	var conf *structpb.Value
	if len(mod.Configs) > 0 {
		conf = &structpb.Value{}
		if err := json.Unmarshal(mod.Configs, &conf); err != nil {
			return nil, err
		}
	}

	return &entropyv1beta1.Module{
		Urn:       mod.URN,
		Name:      mod.Name,
		Configs:   conf,
		Project:   mod.Project,
		CreatedAt: timestamppb.New(mod.CreatedAt),
		UpdatedAt: timestamppb.New(mod.UpdatedAt),
	}, nil
}

func moduleFromProto(res *entropyv1beta1.Module) (*module.Module, error) {
	confJSON, err := getConfigsAsRawJSON(res)
	if err != nil {
		return nil, err
	}

	return &module.Module{
		URN:       res.GetUrn(),
		Name:      res.GetName(),
		Configs:   confJSON,
		Project:   res.GetProject(),
		CreatedAt: res.GetCreatedAt().AsTime(),
		UpdatedAt: res.GetUpdatedAt().AsTime(),
	}, nil
}

func getConfigsAsRawJSON(v interface{ GetConfigs() *structpb.Value }) ([]byte, error) {
	errInvalidJSON := errors.ErrInvalid.WithMsgf("'configs' field must be specified and must be valid JSON")

	confVal := v.GetConfigs()
	if confVal == nil {
		return nil, errInvalidJSON
	}

	confJSON, err := confVal.MarshalJSON()
	if err != nil {
		return nil, errInvalidJSON.WithCausef(err.Error())
	}
	return confJSON, nil
}
