package bind

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func createYAMLSupplier(t *testing.T, s string) Supplier {
	t.Helper()
	js, err := NewYAMLSupplier(bytes.NewBuffer([]byte(s)))
	require.NoError(t, err)
	return js
}

func TestYAMLSupplier(t *testing.T) {
	m := map[string]any{
		keyStr: valStr,
		keyInt: valInt,
		keyNested: map[string]any{
			nestedKey: valNested,
		},
	}

	yaml, err := yaml.Marshal(m)
	require.NoError(t, err)

	s := createYAMLSupplier(t, string(yaml))
	empty := createYAMLSupplier(t, "")

	runSupplierTests(t, s, empty, TagYAML)
	runNestedSupplierTest(t, s)
}
