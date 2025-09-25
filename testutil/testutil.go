package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zackarysantana/bind"
)

// Shared dataset keys and values used across all suppliers.
const (
	KeyMissing = "missing"
	KeyStr     = "str"
	KeyInt     = "integer"
	KeyNested  = "nested"

	ValStr    = "bar"
	ValInt    = 123
	NestedKey = "bar"
	ValNested = "baz"
)

// RunSupplierTests is a shared test util that suppliers use
// to test themselves.
func RunSupplierTests(t *testing.T, s bind.Supplier, emptySupplier bind.Supplier, wantKind string) {
	t.Helper()

	t.Run("Fill", func(t *testing.T) {
		t.Run("Missing", func(t *testing.T) {
			var v string
			ok, err := s.Fill(t.Context(), KeyMissing, nil, &v)
			require.NoError(t, err)
			require.False(t, ok)
			assert.Empty(t, v)
		})
		t.Run("Empty", func(t *testing.T) {
			var v string
			ok, err := emptySupplier.Fill(t.Context(), "empty", nil, &v)
			require.NoError(t, err)
			require.False(t, ok)
			assert.Empty(t, v)
		})
		t.Run("String", func(t *testing.T) {
			var v string
			ok, err := s.Fill(t.Context(), KeyStr, nil, &v)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, ValStr, v)
		})
		t.Run("Int", func(t *testing.T) {
			var v int
			ok, err := s.Fill(t.Context(), KeyInt, nil, &v)
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, 123, v)
		})
	})

	t.Run("Kind", func(t *testing.T) {
		assert.True(t, s.IsKind(wantKind))
	})
}

func RunNestedSupplierTest(t *testing.T, s bind.Supplier) {
	t.Run("Struct", func(t *testing.T) {
		var v struct {
			Bar string
		}
		ok, err := s.Fill(t.Context(), KeyNested, nil, &v)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "baz", v.Bar)
	})
}
