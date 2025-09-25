package bind_test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/testutil"
)

// runCommonSupplierTests runs the same assertions against any Supplier.

func TestPathSupplier(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.org/resource", nil)
	req.SetPathValue(testutil.KeyStr, testutil.ValStr)
	req.SetPathValue(testutil.KeyInt, fmt.Sprint(testutil.ValInt))

	s, err := bind.NewPathSupplier(req)
	require.NoError(t, err)

	empty, err := bind.NewPathSupplier(httptest.NewRequest(http.MethodGet, "http://example.org/resource", nil))
	require.NoError(t, err)

	testutil.RunSupplierTests(t, s, empty, bind.TagPath)
}

func TestQuerySupplier(t *testing.T) {
	q := url.Values{}
	q.Set(testutil.KeyStr, testutil.ValStr)             // ?str=bar
	q.Set(testutil.KeyInt, fmt.Sprint(testutil.ValInt)) // ?integer=123

	u, err := url.Parse("http://example.org/search?" + q.Encode())
	require.NoError(t, err)

	s, err := bind.NewQuerySupplier(u)
	require.NoError(t, err)

	empty, err := bind.NewQuerySupplier(&url.URL{})
	require.NoError(t, err)

	testutil.RunSupplierTests(t, s, empty, bind.TagQuery)
}

func TestHeaderSupplier(t *testing.T) {
	h := http.Header{}
	h.Set(testutil.KeyStr, testutil.ValStr)
	h.Set(testutil.KeyInt, fmt.Sprint(testutil.ValInt))

	s, err := bind.NewHeaderSupplier(h)
	require.NoError(t, err)

	empty, err := bind.NewHeaderSupplier(http.Header{})
	require.NoError(t, err)

	testutil.RunSupplierTests(t, s, empty, bind.TagHeader)
}

func TestFormSupplier(t *testing.T) {
	form := url.Values{}
	form.Set(testutil.KeyStr, testutil.ValStr)
	form.Set(testutil.KeyInt, fmt.Sprint(testutil.ValInt))

	req := httptest.NewRequest(http.MethodPost, "http://example.org/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s, err := bind.NewFormSupplier(req)
	require.NoError(t, err)

	empty, err := bind.NewFormSupplier(httptest.NewRequest(http.MethodPost, "http://example.org/form", nil))
	require.NoError(t, err)

	testutil.RunSupplierTests(t, s, empty, bind.TagForm)
}

func TestEnvSupplier(t *testing.T) {
	// Ensure environment matches the shared dataset.
	t.Setenv(testutil.KeyStr, testutil.ValStr)
	t.Setenv(testutil.KeyInt, fmt.Sprint(testutil.ValInt))

	s := bind.NewEnvSupplier()
	testutil.RunSupplierTests(t, s, s, bind.TagEnv)
}

func TestFlagSupplier(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = fs.String(testutil.KeyStr, "", "desc: string")
	_ = fs.String(testutil.KeyInt, "", "desc: integer")
	_ = fs.String(testutil.KeyNested, "", "desc: nested json")
	// Note: missing is intentionally not defined or set.

	args := []string{
		"-" + testutil.KeyStr, testutil.ValStr,
		"-" + testutil.KeyInt, fmt.Sprint(testutil.ValInt),
	}
	require.NoError(t, fs.Parse(args))

	s := bind.NewFlagSupplier(fs)
	empty := bind.NewFlagSupplier(flag.NewFlagSet("test2", flag.ContinueOnError))

	testutil.RunSupplierTests(t, s, empty, bind.TagFlag)
}
