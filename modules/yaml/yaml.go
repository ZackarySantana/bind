package yaml

import (
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

const (
	TagYAML = "yaml"
)

type YAMLSupplier struct {
	raw map[string]yaml.Node
}

// New uses the given reader to supply YAML values.
func New(src io.Reader) (*YAMLSupplier, error) {
	if src == nil {
		// treat empty input as empty object (graceful)
		return &YAMLSupplier{raw: map[string]yaml.Node{}}, nil
	}
	var m map[string]yaml.Node
	if err := yaml.NewDecoder(src).Decode(&m); err != nil {
		if err == io.EOF {
			// treat empty input as empty object (graceful)
			return &YAMLSupplier{raw: map[string]yaml.Node{}}, nil
		}
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return &YAMLSupplier{raw: m}, nil
}

func (j *YAMLSupplier) Fill(_ context.Context, value string, options []string, dst any) (bool, error) {
	raw, ok := j.raw[value]
	if !ok {
		return false, nil
	}
	return true, raw.Decode(dst)
}

func (j *YAMLSupplier) Kind() string {
	return TagYAML
}

func (j *YAMLSupplier) IsKind(kind string) bool {
	return j.Kind() == kind
}
