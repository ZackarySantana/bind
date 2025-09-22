package bind

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lazyVal struct {
	String Lazy[string]  `json:"string"`
	Int    Lazy[int]     `json:"int"`
	Float  Lazy[float64] `json:"float"`
	Bool   Lazy[bool]    `json:"bool"`
}

func TestLazy(t *testing.T) {
	destination := lazyVal{}
	require.NoError(t, Bind(t.Context(), &destination, []Supplier{
		createJSONSupplier(t, `{
			"string": "hello",
			"int": 123,
			"float": 3.14,
			"bool": true
		}`),
	}))

	str, err := destination.String.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "hello", str)

	i, err := destination.Int.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 123, i)

	f, err := destination.Float.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 3.14, f)

	b, err := destination.Bool.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, true, b)
}
