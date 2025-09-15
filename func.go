package bind

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
)

const (
	TagPath   = "path"
	TagQuery  = "query"
	TagHeader = "header"
	TagForm   = "form"
	TagEnv    = "env"
	TagFlag   = "flag"
)

func NewRequestSuppliers(req *http.Request) ([]Supplier, error) {
	ps, err := NewPathSupplier(req)
	if err != nil {
		return nil, fmt.Errorf("creating path supplier: %w", err)
	}

	qs, err := NewQuerySupplier(req.URL)
	if err != nil {
		return nil, fmt.Errorf("creating query supplier: %w", err)
	}

	hs, err := NewHeaderSupplier(req.Header)
	if err != nil {
		return nil, fmt.Errorf("creating header supplier: %w", err)
	}

	fs, err := NewFormSupplier(req)
	if err != nil {
		return nil, fmt.Errorf("creating form supplier: %w", err)
	}

	return []Supplier{ps, qs, hs, fs}, nil
}

// NewPathSupplier uses the given http.Request to supply path values.
// An example path value is the "org_id" in "/orgs/{org_id}/users".
func NewPathSupplier(req *http.Request) (Supplier, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	return NewFuncStringSupplier(func(_ context.Context, name string, options []string) (string, error) {
		return req.PathValue(name), nil
	}, TagPath), nil
}

// NewQuerySupplier uses the given url.URL to supply query values.
// An example query value is the "hello" in "/search?q=hello".
func NewQuerySupplier(url *url.URL) (Supplier, error) {
	if url == nil {
		return nil, errors.New("url is nil")
	}
	query := url.Query()
	return NewFuncStringSupplier(func(_ context.Context, name string, options []string) (string, error) {
		return query.Get(name), nil
	}, TagQuery), nil
}

// NewHeaderSupplier uses the given http.Header to supply header values.
func NewHeaderSupplier(h http.Header) (Supplier, error) {
	if h == nil {
		return nil, errors.New("header is nil")
	}
	return NewFuncStringSupplier(func(_ context.Context, name string, options []string) (string, error) {
		return h.Get(name), nil
	}, TagHeader), nil
}

// NewFormSupplier uses the given http.Request to supply form values.
// For documentation on the form values, look at http.Request's FormValue
// documentation.
func NewFormSupplier(r *http.Request) (Supplier, error) {
	if r == nil {
		return nil, errors.New("request is nil")
	}
	return NewFuncStringSupplier(func(_ context.Context, name string, options []string) (string, error) {
		return r.FormValue(name), nil
	}, TagForm), nil
}

// NewEnvSupplier grabs values from environment variables.
func NewEnvSupplier() Supplier {
	return NewFuncStringSupplier(func(_ context.Context, name string, options []string) (string, error) {
		return os.Getenv(name), nil
	}, TagEnv)
}

// NewFlagSupplier grabs values from a given FlagSet.
// If provided nil, it will default to the flag.CommandLine flags.
func NewFlagSupplier(fs *flag.FlagSet) Supplier {
	if fs == nil {
		fs = flag.CommandLine
	}
	m := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		m[f.Name] = f.Value.String()
	})
	return NewFuncStringSupplier(func(ctx context.Context, name string, options []string) (string, error) {
		return m[name], nil
	}, TagFlag)
}

type FuncStringSupplier struct {
	fn   func(ctx context.Context, name string, options []string) (string, error)
	kind string
}

// NewFuncStringSupplier is a helper supplier, it is used to construct other suppliers.
// It will attempt to set the string returned from the function
// into the type you are binding to, using parsing if needed.
func NewFuncStringSupplier(fn func(ctx context.Context, name string, options []string) (string, error), kind string) *FuncStringSupplier {
	return &FuncStringSupplier{
		fn:   fn,
		kind: kind,
	}
}

func (p *FuncStringSupplier) Fill(ctx context.Context, value string, options []string, dst any) (bool, error) {
	if p.fn == nil {
		return false, nil
	}
	result, err := p.fn(ctx, value, options)
	if err != nil {
		return false, fmt.Errorf("supplier %q: %w", p.kind, err)
	}
	if result == "" {
		return false, nil
	}

	setStringInto(reflect.ValueOf(dst), result)

	return true, nil
}

func (p *FuncStringSupplier) Kind() string {
	return p.kind
}

func (p *FuncStringSupplier) IsKind(kind string) bool {
	return p.kind == kind
}

type FuncSupplier struct {
	fn   func(ctx context.Context, name string, options []string) (any, error)
	kind string
}

// NewFuncSupplier is a helper supplier, it is used to construct other suppliers.
// It requires that the type you are binding is the same
// as the type the supplier function returns.
func NewFuncSupplier(fn func(ctx context.Context, name string, options []string) (any, error), kind string) *FuncSupplier {
	return &FuncSupplier{
		fn:   fn,
		kind: kind,
	}
}

func (p *FuncSupplier) Fill(ctx context.Context, value string, options []string, dst any) (bool, error) {
	if p.fn == nil {
		return false, nil
	}
	result, err := p.fn(ctx, value, options)
	if err != nil {
		return false, fmt.Errorf("supplier %q: %w", p.kind, err)
	}
	if result == nil {
		return false, nil
	}

	ev := reflect.ValueOf(dst)
	if ev.Kind() != reflect.Ptr || ev.IsNil() {
		return false, fmt.Errorf("supplier %q: dst must be a non-nil pointer", p.kind)
	}
	rv := reflect.ValueOf(result)
	if !rv.Type().AssignableTo(ev.Elem().Type()) {
		return false, fmt.Errorf("supplier %q: cannot assign %s to %s", p.kind, rv.Type().String(), ev.Elem().Type().String())
	}
	ev.Elem().Set(rv)

	return true, nil
}

func (p *FuncSupplier) Kind() string {
	return p.kind
}

func (p *FuncSupplier) IsKind(kind string) bool {
	return p.kind == kind
}
