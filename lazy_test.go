package bind

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registeredStruct struct {
	SomeValue string
	Other     int
}

type lazyVal struct {
	String Lazy[string]           `json:"string"`
	Int    Lazy[int]              `json:"int"`
	Float  Lazy[float64]          `json:"float"`
	Bool   Lazy[bool]             `json:"bool"`
	Struct Lazy[registeredStruct] `json:"struct"`
	Custom Lazy[string]           `custom:"value"`
}

type unregisteredStruct struct {
	SomeValue string
	Other     int
}

type lazyValUnreg struct {
	Unreg Lazy[unregisteredStruct] `json:"unreg"`
}

func TestLazy(t *testing.T) {
	RegisterLazy(func(loader func(context.Context, any) error) Lazy[registeredStruct] {
		return wrapLazy[registeredStruct](loader)
	})

	var customReturn string
	customRanCount := 0
	destination := lazyVal{}
	require.NoError(t, Bind(t.Context(), &destination, []Supplier{
		createJSONSupplier(t, `{
			"string": "hello",
			"int": 123,
			"float": 3.14,
			"bool": true,
			"struct": {"SomeValue": "value", "Other": 42}
		}`),
		NewFuncStringSupplier(func(ctx context.Context, name string, options []string) (string, error) {
			customRanCount++
			return customReturn, nil
		}, "custom"),
	}))

	t.Run("String", func(t *testing.T) {
		str, err := destination.String.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "hello", str)

		str, err = destination.String.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "hello", str)
	})

	t.Run("Int", func(t *testing.T) {
		i, err := destination.Int.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 123, i)
	})

	t.Run("Float", func(t *testing.T) {
		f, err := destination.Float.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 3.14, f)
	})

	t.Run("Bool", func(t *testing.T) {
		b, err := destination.Bool.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, true, b)
	})

	t.Run("RegisteredStruct", func(t *testing.T) {
		s, err := destination.Struct.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, registeredStruct{SomeValue: "value", Other: 42}, s)
	})

	t.Run("Custom", func(t *testing.T) {
		assert.Equal(t, 0, customRanCount)
		custom, err := destination.Custom.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, customReturn, custom)
		assert.Equal(t, 1, customRanCount)

		t.Run("NotCached", func(t *testing.T) {
			customReturn = "new value"
			custom, err = destination.Custom.Get(t.Context())
			require.NoError(t, err)
			assert.Equal(t, customReturn, custom)
			assert.Equal(t, 2, customRanCount)
		})
	})

	t.Run("UnregisteredStruct", func(t *testing.T) {
		destination := lazyValUnreg{}
		require.ErrorContains(t, Bind(t.Context(), &destination, []Supplier{
			createJSONSupplier(t, `{
				"unreg": {"SomeValue": "value", "Other": 42}
			}`),
		}), "did you forget to call bind.RegisterLazy")
	})
}
