package bind

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

const (
	TagJSON = "json"
)

type JSONSupplier struct {
	raw map[string]json.RawMessage
}

// NewJSONSupplier uses the given reader to supply JSON values.
func NewJSONSupplier(src io.Reader) (*JSONSupplier, error) {
	if src == nil {
		// treat empty input as empty object (graceful)
		return &JSONSupplier{raw: map[string]json.RawMessage{}}, nil
	}
	var m map[string]json.RawMessage
	if err := json.NewDecoder(src).Decode(&m); err != nil {
		if err == io.EOF {
			// treat empty input as empty object (graceful)
			return &JSONSupplier{raw: map[string]json.RawMessage{}}, nil
		}
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return &JSONSupplier{raw: m}, nil
}

func (j *JSONSupplier) Fill(_ context.Context, value string, options []string, dst any) (bool, error) {
	raw, ok := j.raw[value]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(raw, dst)
}

func (j *JSONSupplier) Kind() string {
	return TagJSON
}

func (j *JSONSupplier) IsKind(kind string) bool {
	return j.Kind() == kind
}
