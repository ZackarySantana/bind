package testutil

import (
	"testing"

	"github.com/zackarysantana/bind"
)

var benchSink *BenchBindResult

var BenchBindData = map[string]any{
	"Int":    7,
	"Str":    "hello",
	"Float":  3.14,
	"Bool":   true,
	"Ptr":    "other",
	"Higher": "present",
	"Lazy":   "lazy",
}

type BenchBindResult struct {
	Int    int               `test-only:"Int"`
	Str    string            `test-only:"Str"`
	Float  float64           `test-only:"Float"`
	Bool   bool              `test-only:"Bool"`
	Ptr    *string           `test-only:"Ptr"`
	Lazy   bind.Lazy[string] `test-only:"Lazy"`
	Higher string            `test-only:"Higher" options:"level=2"`
}

func benchPaused(b *testing.B, fn func()) {
	b.StopTimer()
	fn()
	b.StartTimer()
}

func compareBase(b *testing.B, dst BenchBindResult) {
	if dst.Int != BenchBindData["Int"] || dst.Str != BenchBindData["Str"] ||
		dst.Float != BenchBindData["Float"] || dst.Bool != BenchBindData["Bool"] || dst.Ptr == nil ||
		*dst.Ptr != BenchBindData["Ptr"] || dst.Lazy == nil {
		b.Fatalf("bind produced wrong value: %+v", dst)
	}
	lazyVal, err := dst.Lazy.Get(b.Context())
	if err != nil {
		b.Fatalf("lazy get failed: %v", err)
	}
	if lazyVal != BenchBindData["Lazy"] {
		b.Fatalf("lazy produced wrong value: %q", lazyVal)
	}
}

func RunBindBenchmark(b *testing.B, getSuppliers func() []bind.Supplier) {
	suppliers := getSuppliers()

	b.Run("Bind", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var dst BenchBindResult
			if err := bind.Bind(b.Context(), &dst, suppliers, bind.WithTestOnly()); err != nil {
				b.Fatalf("bind failed: %v", err)
			}
			benchSink = &dst
		}

		benchPaused(b, func() {
			compareBase(b, *benchSink)
		})
	})

	b.Run("BindWithLevel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var dst BenchBindResult
			if err := bind.Bind(b.Context(), &dst, suppliers, bind.WithTestOnly(), bind.WithLevel(2)); err != nil {
				b.Fatalf("bind failed: %v", err)
			}
			benchSink = &dst
		}

		benchPaused(b, func() {
			compareBase(b, *benchSink)
			if benchSink.Higher != BenchBindData["Higher"] {
				b.Fatalf("bind produced wrong higher value: %+v", *benchSink)
			}
		})
	})

	b.Run("BindAndCreateSuppliers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			suppliers := getSuppliers()
			var dst BenchBindResult
			if err := bind.Bind(b.Context(), &dst, suppliers, bind.WithTestOnly()); err != nil {
				b.Fatalf("bind failed: %v", err)
			}
			benchSink = &dst
		}

		benchPaused(b, func() {
			compareBase(b, *benchSink)
		})
	})
}

// TODO: Make a supplier benchmark:
// func RunSupplierBenchmark[T any](b *testing.B, supplier bind.Supplier, opts ...bind.Option) {
// 	suppliers := []bind.Supplier{supplier}

// 	for i := 0; i < b.N; i++ {
// 		var dst T
// 		if err := bind.Bind(b.Context(), &dst, suppliers, opts...); err != nil {
// 			b.Fatalf("supplier failed: %v", err)
// 		}
// 	}
// }
