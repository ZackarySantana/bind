package bind_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/testutil"
)

func BenchmarkJSON(b *testing.B) {
	jsonData, err := json.Marshal(testutil.BenchBindData)
	if err != nil {
		b.Fatalf("marshal JSON: %v", err)
	}

	testutil.RunBindBenchmark(b, func() []bind.Supplier {
		sup, err := bind.NewJSONSupplier(bytes.NewReader(jsonData))
		if err != nil {
			b.Fatalf("new JSON supplier: %v", err)
		}
		return []bind.Supplier{sup}
	})
}
