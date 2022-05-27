package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func parseFile(filePath string, v protoreflect.ProtoMessage) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := protojson.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := protojson.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type") //nolint
	}

	return nil
}

func prettyPrint(i interface{}) string {
	s, e := json.MarshalIndent(i, "", "\t")
	if e != nil {
		return fmt.Sprintln(e.Error())
	}
	return string(s)
}
