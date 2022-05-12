package module

import (
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/odpf/entropy/pkg/errors"
)

const (
	CreateAction = "create"
	UpdateAction = "update"
)

// ActionRequest describes an invocation of action on module.
type ActionRequest struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// ActionDesc is a descriptor for an action supported by a module.
type ActionDesc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParamSchema string `json:"param_schema"`

	schema *gojsonschema.Schema
}

func (ad ActionDesc) validateReq(req ActionRequest) error {
	result, err := ad.schema.Validate(gojsonschema.NewGoLoader(req.Params))
	if err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	} else if !result.Valid() {
		var errorStrings []string
		for _, resultErr := range result.Errors() {
			errorStrings = append(errorStrings, resultErr.String())
		}
		errorString := strings.Join(errorStrings, "\n")
		return errors.ErrInvalid.WithMsgf(errorString)
	}

	return nil
}
