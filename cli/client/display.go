package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/goto/entropy/pkg/errors"
)

type FormatFn func(w io.Writer, v any) error

// Display formats the given value 'v' using format specified as value for --format
// flag and writes to STDOUT. If --format=pretty/human, custom-formatter passed will
// be used.
func Display(cmd *cobra.Command, v any, prettyFormatter FormatFn) error {
	format, _ := cmd.Flags().GetString("format")
	format = strings.ToLower(strings.TrimSpace(format))

	var formatter FormatFn
	switch format {
	case "json":
		formatter = JSONFormat

	case "yaml", "yml":
		formatter = YAMLFormat

	case "toml":
		formatter = TOMLFormat

	case "pretty", "human":
		if prettyFormatter != nil {
			formatter = prettyFormatter
		} else {
			formatter = GoFormat
		}
	}

	if formatter == nil {
		return errors.Errorf("--format value '%s' is not valid", format)
	}

	return formatter(os.Stdout, v)
}

// JSONFormat outputs 'v' formatted as indented JSON.
func JSONFormat(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// TOMLFormat outputs 'v' formatted as per TOML spec.
func TOMLFormat(w io.Writer, v any) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(v)
}

// YAMLFormat outputs 'v' formatted as per YAML spec.
func YAMLFormat(w io.Writer, v any) error {
	// note: since most values are json tagged but may not be
	// yaml tagged, we do this to ensure keys are snake-cased.
	val, err := jsonConvert(v)
	if err != nil {
		return err
	}
	return yaml.NewEncoder(w).Encode(val)
}

// GoFormat outputs 'v' formatted using pp package.
func GoFormat(w io.Writer, v any) error {
	_, err := fmt.Fprintln(w, v)
	return err
}

func jsonConvert(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var val any
	if err := json.Unmarshal(b, &val); err != nil {
		return nil, err
	}
	return val, nil
}
