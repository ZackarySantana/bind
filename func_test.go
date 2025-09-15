package bind

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Shared dataset keys and values used across all suppliers.
const (
	keyMissing = "missing"
	keyStr     = "str"
	keyInt     = "integer"
	keyNested  = "nested"

	valStr    = "bar"
	valInt    = 123
	nestedKey = "bar"
	valNested = "baz"
)

// runCommonSupplierTests runs the same assertions against any Supplier.

func TestPathSupplier(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.org/resource", nil)
	req.SetPathValue(keyStr, valStr)
	req.SetPathValue(keyInt, fmt.Sprint(valInt))

	s, err := NewPathSupplier(req)
	require.NoError(t, err)

	empty, err := NewPathSupplier(httptest.NewRequest(http.MethodGet, "http://example.org/resource", nil))
	require.NoError(t, err)

	runSupplierTests(t, s, empty, TagPath)
}

func TestQuerySupplier(t *testing.T) {
	q := url.Values{}
	q.Set(keyStr, valStr)             // ?str=bar
	q.Set(keyInt, fmt.Sprint(valInt)) // ?integer=123

	u, err := url.Parse("http://example.org/search?" + q.Encode())
	require.NoError(t, err)

	s, err := NewQuerySupplier(u)
	require.NoError(t, err)

	empty, err := NewQuerySupplier(&url.URL{})
	require.NoError(t, err)

	runSupplierTests(t, s, empty, TagQuery)
}

func TestHeaderSupplier(t *testing.T) {
	h := http.Header{}
	h.Set(keyStr, valStr)
	h.Set(keyInt, fmt.Sprint(valInt))

	s, err := NewHeaderSupplier(h)
	require.NoError(t, err)

	empty, err := NewHeaderSupplier(http.Header{})
	require.NoError(t, err)

	runSupplierTests(t, s, empty, TagHeader)
}

func TestFormSupplier(t *testing.T) {
	form := url.Values{}
	form.Set(keyStr, valStr)
	form.Set(keyInt, fmt.Sprint(valInt))

	req := httptest.NewRequest(http.MethodPost, "http://example.org/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s, err := NewFormSupplier(req)
	require.NoError(t, err)

	empty, err := NewFormSupplier(httptest.NewRequest(http.MethodPost, "http://example.org/form", nil))
	require.NoError(t, err)

	runSupplierTests(t, s, empty, TagForm)
}

func TestEnvSupplier(t *testing.T) {
	// Ensure environment matches the shared dataset.
	t.Setenv(keyStr, valStr)
	t.Setenv(keyInt, fmt.Sprint(valInt))

	s := NewEnvSupplier()
	runSupplierTests(t, s, s, TagEnv)
}

func TestFlagSupplier(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String(keyStr, "", "string")
	_ = fs.String(keyInt, "", "integer")
	_ = fs.String(keyNested, "", "nested json")
	// Note: missing is intentionally not defined or set.

	args := []string{
		"-" + keyStr, valStr,
		"-" + keyInt, fmt.Sprint(valInt),
	}
	require.NoError(t, fs.Parse(args))

	s := NewFlagSupplier(fs)
	empty := NewFlagSupplier(flag.NewFlagSet("test2", flag.ContinueOnError))

	runSupplierTests(t, s, empty, TagFlag)
}
