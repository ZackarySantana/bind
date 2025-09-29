package bind_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/testutil"
)

func createJSONSupplier(t *testing.T, s string) bind.Supplier {
	t.Helper()
	js, err := bind.NewJSONSupplier(bytes.NewBuffer([]byte(s)))
	require.NoError(t, err)
	return js
}

func TestJSONSupplier(t *testing.T) {
	m := map[string]any{
		testutil.KeyStr: testutil.ValStr,
		testutil.KeyInt: testutil.ValInt,
		testutil.KeyNested: map[string]any{
			testutil.NestedKey: testutil.ValNested,
		},
	}

	json, err := json.Marshal(m)
	require.NoError(t, err)

	s := createJSONSupplier(t, string(json))
	empty := createJSONSupplier(t, "")

	testutil.RunSupplierTests(t, s, empty, bind.TagJSON)
	testutil.RunNestedSupplierTest(t, s)
}

func BenchmarkJSON(b *testing.B) {
	data, err := json.Marshal(testutil.BenchBindData)
	if err != nil {
		b.Fatal(err)
	}

	testutil.RunBindBenchmark(b, func() []bind.Supplier {
		sup, err := bind.NewJSONSupplier(bytes.NewReader(data))
		if err != nil {
			b.Fatal(err)
		}
		return []bind.Supplier{sup}
	})
}
