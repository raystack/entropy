package module

import (
	"encoding/json"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/goto/entropy/pkg/errors"
)

const (
	CreateAction = "create"
	UpdateAction = "update"
	DeleteAction = "delete"
)

// ActionRequest describes an invocation of action on module.
type ActionRequest struct {
	Name   string            `json:"name"`
	Params json.RawMessage   `json:"params"`
	Labels map[string]string `json:"labels"`
	UserID string
}

// ActionDesc is a descriptor for an action supported by a module.
type ActionDesc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParamSchema string `json:"param_schema"`

	schema *gojsonschema.Schema
}

// Sanitise cleans action description by removing whitespaces, checking
// parameter-schema etc.
func (ad *ActionDesc) Sanitise() error {
	if ad.ParamSchema == "" {
		return nil
	}

	var err error
	ad.schema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(ad.ParamSchema))
	if err != nil {
		return errors.ErrInvalid.
			WithMsgf("parameter schema for action '%s' is not valid", ad.Name).
			WithCausef(err.Error())
	}
	return nil
}

func (ad ActionDesc) validateReq(req ActionRequest) error {
	if ad.schema == nil {
		return nil
	}

	result, err := ad.schema.Validate(gojsonschema.NewBytesLoader(req.Params))
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
