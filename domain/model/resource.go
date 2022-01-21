package model

import (
	"strings"
	"time"

	entropy "go.buf.build/odpf/gwv/whoabhisheksah/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Resource struct {
	Urn       string                 `bson:"urn"`
	Name      string                 `bson:"name"`
	Parent    string                 `bson:"parent"`
	Kind      string                 `bson:"kind"`
	Configs   map[string]interface{} `bson:"configs"`
	Labels    map[string]string      `bson:"labels"`
	Status    string                 `bson:"status"`
	CreatedAt time.Time              `bson:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at"`
}

func ResourceToProto(res *Resource) (*entropy.Resource, error) {
	conf, err := structpb.NewValue(res.Configs)
	if err != nil {
		return nil, err
	}
	return &entropy.Resource{
		Urn:       res.Urn,
		Name:      res.Name,
		Parent:    res.Parent,
		Kind:      res.Kind,
		Configs:   conf,
		Labels:    res.Labels,
		Status:    res.Status,
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}, nil
}

func ResourceFromProto(res *entropy.Resource) *Resource {
	return &Resource{
		Urn:     res.GetUrn(),
		Name:    res.GetName(),
		Parent:  res.GetParent(),
		Kind:    res.GetKind(),
		Configs: res.GetConfigs().GetStructValue().AsMap(),
		Labels:  res.GetLabels(),
	}
}

func GenerateResourceUrn(res *Resource) string {
	return strings.Join([]string{
		sanitizeString(res.Parent),
		sanitizeString(res.Name),
		sanitizeString(res.Kind),
	}, "-")
}

func sanitizeString(s string) string {
	return strings.Replace(s, " ", "_", -1)
}
