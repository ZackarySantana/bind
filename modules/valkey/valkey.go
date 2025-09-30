package valkey

import (
	"context"
	"fmt"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/zackarysantana/bind"
)

const (
	TagExactValKey = "exact-valkey"
)

// NewExactSupplier creates a Supplier that looks up documents from the given valkey client.
// It uses the options provided to build a filter for the lookup.
// An example struct tag usage is:
//
//	exact-valkey:"<key>"
//
// This would perform a Get operation with the given key.
//
// This supplier uses the [parser](https://pkg.go.dev/github.com/zackarysantana/bind/parser) package to parse the options into appropriate types.
func NewExactSupplier[T any](client *glide.Client) bind.Supplier {
	return bind.NewFuncSupplier(func(ctx context.Context, name string, options []string) (any, error) {
		// The 'name' is also an option.
		options = append(options, name)
		var out T
		// result, err := client.Get(ctx, name)
		result, err := client.CustomCommand(ctx, []string{"GET", name})
		if err != nil {
			return nil, err
		}
		fmt.Printf("Result: %v and type %T\n", result, result)

		return out, nil
		// return out, client.Get(ctx, filter).Decode(&out)
	}, TagExactValKey)
}
