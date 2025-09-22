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

	destination := lazyVal{}
	require.NoError(t, Bind(t.Context(), &destination, []Supplier{
		createJSONSupplier(t, `{
			"string": "hello",
			"int": 123,
			"float": 3.14,
			"bool": true,
			"struct": {"SomeValue": "value", "Other": 42}
		}`),
	}))

	t.Run("String", func(t *testing.T) {
		str, err := destination.String.Get(t.Context())
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

	t.Run("UnregisteredStruct", func(t *testing.T) {
		destination := lazyValUnreg{}
		require.ErrorContains(t, Bind(t.Context(), &destination, []Supplier{
			createJSONSupplier(t, `{
				"unreg": {"SomeValue": "value", "Other": 42}
			}`),
		}), "did you forget to call bind.RegisterLazy")
	})
}
