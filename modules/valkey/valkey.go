package valkey

import (
	"context"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/zackarysantana/bind"
)

const (
	TagValKey = "valkey"
)

func NewSupplier[T any](client *glide.Client, ref any) (bind.Supplier, error) {
	return bind.NewSelfSupplier(func(ctx context.Context, filter map[string]any) (T, error) {
		var out T
		return out, nil
		// return out, client.Get(ctx, filter).Decode(&out)
	}, TagValKey, ref)
}
