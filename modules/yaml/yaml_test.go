package yaml

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/testutil"
	"gopkg.in/yaml.v3"
)

func createYAMLSupplier(t *testing.T, s string) bind.Supplier {
	t.Helper()
	js, err := New(bytes.NewBuffer([]byte(s)))
	require.NoError(t, err)
	return js
}

func TestYAMLSupplier(t *testing.T) {
	m := map[string]any{
		testutil.KeyStr: testutil.ValStr,
		testutil.KeyInt: testutil.ValInt,
		testutil.KeyNested: map[string]any{
			testutil.NestedKey: testutil.ValNested,
		},
	}

	yaml, err := yaml.Marshal(m)
	require.NoError(t, err)

	s := createYAMLSupplier(t, string(yaml))
	empty := createYAMLSupplier(t, "")

	testutil.RunSupplierTests(t, s, empty, TagYAML)
	testutil.RunNestedSupplierTest(t, s)
}

func BenchmarkYAML(b *testing.B) {
	data, err := yaml.Marshal(testutil.BenchBindData)
	if err != nil {
		b.Fatal(err)
	}

	testutil.RunBindBenchmark(b, func() []bind.Supplier {
		sup, err := New(bytes.NewReader(data))
		if err != nil {
			b.Fatal(err)
		}
		return []bind.Supplier{sup}
	})
}
