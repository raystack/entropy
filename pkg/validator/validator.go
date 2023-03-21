package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/xeipuuv/gojsonschema"

	"github.com/goto/entropy/pkg/errors"
)

// FromJSONSchema returns a validator that can validate using a JSON schema.
func FromJSONSchema(schemaVal []byte) func(jsonVal []byte) error {
	schema, schemaErr := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaVal))

	return func(jsonVal []byte) error {
		if schemaErr != nil {
			return errors.ErrInternal.WithCausef(schemaErr.Error())
		}

		result, err := schema.Validate(gojsonschema.NewBytesLoader(jsonVal))
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
}

// TaggedStruct validates the given struct-value using go-validate package
// based on 'validate' tags.
func TaggedStruct(structVal any) error {
	err := validator.New().Struct(structVal)
	if err != nil {
		var fields []string

		var valErr *validator.ValidationErrors
		if errors.As(err, &valErr) {
			for _, fieldError := range *valErr {
				fields = append(fields, fmt.Sprintf("%s: %s", fieldError.Field(), fieldError.Tag()))
			}
			return errors.ErrInvalid.WithMsgf("invalid values for fields").WithCausef(strings.Join(fields, ", "))
		} else {
			return errors.ErrInvalid.WithCausef(err.Error())
		}
	}
	return nil
}
