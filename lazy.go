package bind

import (
	"context"
	"reflect"
	"sync"
)

var lazyFactories sync.Map // map[reflect.Type]func(loader any) any

type LazyLoader func(context.Context, any) error

func init() {
	// Registers the standard types.
	RegisterLazy(func(loader LazyLoader) Lazy[string] {
		return AsLazy[string](loader)
	})
	RegisterLazy(func(loader LazyLoader) Lazy[int] {
		return AsLazy[int](loader)
	})
	RegisterLazy(func(loader LazyLoader) Lazy[float64] {
		return AsLazy[float64](loader)
	})
	RegisterLazy(func(loader LazyLoader) Lazy[bool] {
		return AsLazy[bool](loader)
	})
}

func RegisterLazy[T any](factory func(LazyLoader) Lazy[T]) {
	it := reflect.TypeOf((*Lazy[T])(nil)).Elem()
	lazyFactories.Store(it, func(loader any) any {
		return factory(loader.(LazyLoader))
	})
	registerCache(factory)
}

type Lazy[T any] interface {
	Get(ctx context.Context) (T, error)
}

func AsLazy[T any](fn func(ctx context.Context, ptr any) error) Lazy[T] {
	return &lazyImpl[T]{fn: fn}
}

type lazyImpl[T any] struct {
	fn func(ctx context.Context, ptr any) error
}

func (l *lazyImpl[T]) Get(ctx context.Context) (T, error) {
	var out T
	return out, l.fn(ctx, any(&out))
}
