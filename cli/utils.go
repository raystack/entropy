package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	outputJSON = "json"
	outputYAML = "yaml"
	outputYML  = "yml"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func parseFile(filePath string, v protoreflect.ProtoMessage) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	unmarshalOpts := protojson.UnmarshalOptions{}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := unmarshalOpts.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}

	case ".yaml", ".yml":
		j, err := yaml.YAMLToJSON(b)
		if err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
		if err := unmarshalOpts.Unmarshal(j, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}

	default:
		return errors.New("unsupported file type") // nolint
	}

	return nil
}

func formatOutput(i protoreflect.ProtoMessage, format string) (string, error) {
	marshalOpts := protojson.MarshalOptions{
		Indent:        "\t",
		Multiline:     true,
		UseProtoNames: true,
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
		return "", errors.New("unsupported format") // nolint
	}
}

func fatalExitf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func handleErr(fn RunEFunc) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			fatalExitf(err.Error())
		}
		return nil
	}
}
