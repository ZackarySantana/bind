package bind

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cacheVal struct {
	Custom Cache[string] `custom:"value"`
}

func TestCache(t *testing.T) {
	customReturn := "initial value"
	ranCount := 0
	destination := cacheVal{}
	require.NoError(t, Bind(t.Context(), &destination, []Supplier{
		NewFuncStringSupplier(func(ctx context.Context, name string, options []string) (string, error) {
			ranCount++
			return customReturn, nil
		}, "custom"),
	}))

	assert.Equal(t, 0, ranCount)
	custom, err := destination.Custom.Get(t.Context())
	require.NoError(t, err)
	assert.Equal(t, customReturn, custom)
	assert.Equal(t, 1, ranCount)

	t.Run("IsCached", func(t *testing.T) {
		customReturn = "new value"
		custom, err = destination.Custom.Get(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "initial value", custom)
		assert.Equal(t, 1, ranCount)
	})
}
