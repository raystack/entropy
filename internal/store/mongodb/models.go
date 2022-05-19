package mongodb

import (
	"time"

	"github.com/odpf/entropy/core/resource"
)

type resourceModel struct {
	URN       string            `bson:"urn"`
	Kind      string            `bson:"kind"`
	Name      string            `bson:"name"`
	Project   string            `bson:"project"`
	Labels    map[string]string `bson:"labels"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
	Spec      specModel         `bson:"spec"`
	State     stateModel        `bson:"state"`
}

type specModel struct {
	Configs      []byte            `bson:"configs"`
	Dependencies map[string]string `bson:"dependencies"`
}

type stateModel struct {
	Status     string                 `bson:"status"`
	Output     map[string]interface{} `bson:"output"`
	ModuleData []byte                 `bson:"module_data"`
}

func modelFromResource(res resource.Resource) resourceModel {
	return resourceModel{
		URN:       res.URN,
		Kind:      res.Kind,
		Name:      res.Name,
		Project:   res.Project,
		Labels:    res.Labels,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Spec: specModel{
			Configs:      res.Spec.Configs,
			Dependencies: res.Spec.Dependencies,
		},
		State: stateModel{
			Status:     res.State.Status,
			Output:     res.State.Output,
			ModuleData: res.State.ModuleData,
		},
	}
}

func modelToResource(m resourceModel) *resource.Resource {
	return &resource.Resource{
		URN:       m.URN,
		Kind:      m.Kind,
		Name:      m.Name,
		Project:   m.Project,
		Labels:    m.Labels,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Spec: resource.Spec{
			Configs:      m.Spec.Configs,
			Dependencies: m.Spec.Dependencies,
		},
		State: resource.State{
			Status:     m.State.Status,
			Output:     m.State.Output,
			ModuleData: m.State.ModuleData,
		},
	}
}
