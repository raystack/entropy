package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	outputJSON = "json"
	outputYAML = "yaml"
	outputYML  = "yml"
)

func parseFile(filePath string, v protoreflect.ProtoMessage) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type") //nolint
	}

	return nil
}

func formatOutput(i protoreflect.ProtoMessage, format string) (string, error) {
	marshalOpts := protojson.MarshalOptions{
		Indent:    "\t",
		Multiline: true,
	}

	b, e := marshalOpts.Marshal(i)
	if e != nil {
		return "", e
	}

	switch format {
	case outputJSON:
		return string(b), nil
	case outputYAML, outputYML:
		y, e := yaml.JSONToYAML(b)
		if e != nil {
			return "", e
		}
		return string(y), nil
	default:
		return "", errors.New("unsupported format") //nolint
	}
}
