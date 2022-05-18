package handlersv1

import (
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func resourceToProto(res resource.Resource) (*entropyv1beta1.Resource, error) {
	protoState, err := resourceStateToProto(res.State)
	if err != nil {
		return nil, err
	}

	return &entropyv1beta1.Resource{
		Urn:       res.URN,
		Kind:      res.Kind,
		Project:   res.Project,
		Name:      res.Name,
		Labels:    res.Labels,
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
		Spec:      resourceSpecToProto(res.Spec),
		State:     protoState,
	}, nil
}

func resourceStateToProto(state resource.State) (*entropyv1beta1.ResourceState, error) {
	var outputVal *structpb.Value
	if len(state.Output) > 0 {
		out, err := structpb.NewValue(map[string]interface{}(state.Output))
		if err != nil {
			return nil, err
		}
		outputVal = out
	}

	var protoStatus = entropyv1beta1.ResourceState_STATUS_UNSPECIFIED
	if resourceStatus, ok := entropyv1beta1.ResourceState_Status_value[state.Status]; ok {
		protoStatus = entropyv1beta1.ResourceState_Status(resourceStatus)
	}

	return &entropyv1beta1.ResourceState{
		Status:     protoStatus,
		Output:     outputVal,
		ModuleData: state.ModuleData,
	}, nil
}

func resourceSpecToProto(spec resource.Spec) *entropyv1beta1.ResourceSpec {
	conf, err := structpb.NewValue(spec.Configs)
	if err != nil {
		return nil
	}

	var deps []*entropyv1beta1.ResourceDependency
	for key, ref := range spec.Dependencies {
		deps = append(deps, &entropyv1beta1.ResourceDependency{
			Key:   key,
			Value: ref,
		})
	}

	return &entropyv1beta1.ResourceSpec{Configs: conf, Dependencies: deps}
}

func resourceFromProto(res *entropyv1beta1.Resource) (*resource.Resource, error) {
	spec, err := resourceSpecFromProto(res.Spec)
	if err != nil {
		return nil, err
	}

	return &resource.Resource{
		URN:       res.GetUrn(),
		Kind:      res.GetKind(),
		Name:      res.GetName(),
		Labels:    res.GetLabels(),
		Project:   res.GetProject(),
		CreatedAt: res.GetCreatedAt().AsTime(),
		UpdatedAt: res.GetUpdatedAt().AsTime(),
		Spec:      *spec,
		State: resource.State{
			Status:     res.State.GetStatus().String(),
			Output:     res.State.GetOutput().GetStructValue().AsMap(),
			ModuleData: res.State.GetModuleData(),
		},
	}, nil
}

func resourceSpecFromProto(spec *entropyv1beta1.ResourceSpec) (*resource.Spec, error) {
	deps := map[string]string{}

	for _, dep := range spec.GetDependencies() {
		key, value := dep.GetKey(), dep.GetValue()
		if _, alreadySet := deps[key]; alreadySet {
			return nil, errors.ErrInvalid.WithMsgf("dependency key '%s' is set more than once", dep.GetKey())
		}
		deps[key] = value
	}

	return &resource.Spec{
		Configs:      spec.GetConfigs().GetStructValue().AsMap(),
		Dependencies: deps,
	}, nil
}
