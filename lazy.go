package bind

import (
	"context"
	"reflect"
	"sync"
)

var lazyFactories sync.Map // map[reflect.Type]func(loader any) any

func init() {
	// Registers the standard types.
	RegisterLazy(func(loader func(context.Context, any) error) Lazy[string] {
		return wrapLazy[string](loader)
	})
	RegisterLazy(func(loader func(context.Context, any) error) Lazy[int] {
		return wrapLazy[int](loader)
	})
	RegisterLazy(func(loader func(context.Context, any) error) Lazy[float64] {
		return wrapLazy[float64](loader)
	})
	RegisterLazy(func(loader func(context.Context, any) error) Lazy[bool] {
		return wrapLazy[bool](loader)
	})
}

func RegisterLazy[T any](factory func(func(context.Context, any) error) Lazy[T]) {
	it := reflect.TypeOf((*Lazy[T])(nil)).Elem()
	lazyFactories.Store(it, func(loader any) any {
		return factory(loader.(func(context.Context, any) error))
	})
}

type Lazy[T any] interface {
	Get(ctx context.Context) (T, error)
}

func wrapLazy[T any](fn func(ctx context.Context, ptr any) error) Lazy[T] {
	return &lazyImpl[T]{fn: fn}
}

type lazyImpl[T any] struct {
	fn func(ctx context.Context, ptr any) error
}

func (l *lazyImpl[T]) Get(ctx context.Context) (T, error) {
	var out T
	return out, l.fn(ctx, any(&out))
}
