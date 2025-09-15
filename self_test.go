package bind

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	ID    int
	Phone string
	Name  string `test:"id=ID"`
	Age   int    `test2:"num=Phone,other=ID"`
}

func TestSelfSupplier(t *testing.T) {
	test := testStruct{
		ID:    9001,
		Phone: "970-4133",
	}

	var givenTestFilter map[string]any
	testSupplier, err := NewSelfSupplier(func(ctx context.Context, filter map[string]any) (string, error) {
		givenTestFilter = filter
		return "found!", nil
	}, "test", &test)
	require.NoError(t, err)

	var givenTest2Filter map[string]any
	test2Supplier, err := NewSelfSupplier(func(ctx context.Context, filter map[string]any) (int, error) {
		givenTest2Filter = filter
		return 42, nil
	}, "test2", &test)
	require.NoError(t, err)

	err = Bind(t.Context(), &test, []Supplier{testSupplier, test2Supplier})
	require.NoError(t, err)

	assert.Equal(t, map[string]any{"id": 9001}, givenTestFilter)
	assert.Equal(t, "found!", test.Name)

	assert.Equal(t, map[string]any{"num": "970-4133", "other": 9001}, givenTest2Filter)
	assert.Equal(t, 42, test.Age)
}
