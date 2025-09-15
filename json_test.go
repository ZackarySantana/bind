package bind

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func createJSONSupplier(t *testing.T, s string) Supplier {
	t.Helper()
	js, err := NewJSONSupplier(bytes.NewBuffer([]byte(s)))
	require.NoError(t, err)
	return js
}

func TestJSONSupplier(t *testing.T) {
	m := map[string]any{
		keyStr: valStr,
		keyInt: valInt,
		keyNested: map[string]any{
			nestedKey: valNested,
		},
	}

	json, err := json.Marshal(m)
	require.NoError(t, err)

	s := createJSONSupplier(t, string(json))
	empty := createJSONSupplier(t, "")

	runSupplierTests(t, s, empty, TagJSON)
	runNestedSupplierTest(t, s)
}
