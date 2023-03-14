package resources

import (
	"encoding/json"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

const decimalBase = 10

func resourceToProto(res resource.Resource) (*entropyv1beta1.Resource, error) {
	protoState, err := resourceStateToProto(res.State)
	if err != nil {
		return nil, err
	}

	spec, err := resourceSpecToProto(res.Spec)
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
		Spec:      spec,
		State:     protoState,
	}, nil
}

func resourceStateToProto(state resource.State) (*entropyv1beta1.ResourceState, error) {
	var outputVal *structpb.Value
	if len(state.Output) > 0 {
		outputVal = &structpb.Value{}
		if err := json.Unmarshal(state.Output, outputVal); err != nil {
			return nil, err
		}
	}

	protoStatus := entropyv1beta1.ResourceState_STATUS_UNSPECIFIED
	if resourceStatus, ok := entropyv1beta1.ResourceState_Status_value[state.Status]; ok {
		protoStatus = entropyv1beta1.ResourceState_Status(resourceStatus)
	}

	return &entropyv1beta1.ResourceState{
		Status:     protoStatus,
		Output:     outputVal,
		ModuleData: state.ModuleData,
	}, nil
}

func resourceSpecToProto(spec resource.Spec) (*entropyv1beta1.ResourceSpec, error) {
	conf := structpb.Value{}
	if err := json.Unmarshal(spec.Configs, &conf); err != nil {
		return nil, err
	}

	var deps []*entropyv1beta1.ResourceDependency
	for key, ref := range spec.Dependencies {
		deps = append(deps, &entropyv1beta1.ResourceDependency{
			Key:   key,
			Value: ref,
		})
	}

	return &entropyv1beta1.ResourceSpec{
		Configs:      &conf,
		Dependencies: deps,
	}, nil
}

func resourceFromProto(res *entropyv1beta1.Resource) (*resource.Resource, error) {
	spec, err := resourceSpecFromProto(res.Spec)
	if err != nil {
		return nil, err
	}

	jsonData, err := res.State.GetOutput().GetStructValue().MarshalJSON()
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
			Output:     jsonData,
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

	confJSON, err := spec.GetConfigs().MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &resource.Spec{
		Configs:      confJSON,
		Dependencies: deps,
	}, nil
}

func revisionToProto(revision resource.Revision) (*entropyv1beta1.ResourceRevision, error) {
	spec, err := resourceSpecToProto(revision.Spec)
	if err != nil {
		return nil, err
	}

	return &entropyv1beta1.ResourceRevision{
		Id:        strconv.FormatInt(revision.ID, decimalBase),
		Urn:       revision.URN,
		Reason:    revision.Reason,
		Labels:    revision.Labels,
		CreatedAt: timestamppb.New(revision.CreatedAt),
		Spec:      spec,
	}, nil
}
