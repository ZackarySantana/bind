package bind

import (
	"context"
	"reflect"
	"sync"
)

var cacheFactories sync.Map // map[reflect.Type]func(loader any) any

type Cache[T any] Lazy[T]

func registerCache[T any](factory func(LazyLoader) Lazy[T]) {
	it := reflect.TypeOf((*Cache[T])(nil)).Elem()
	cacheFactories.Store(it, func(loader any) any {
		return &cacheImpl[T]{
			Lazy: factory(loader.(LazyLoader)),
		}
	})
}

type cacheImpl[T any] struct {
	Lazy[T]

	val *T
}

func (c *cacheImpl[T]) Get(ctx context.Context) (T, error) {
	if c.val != nil {
		return *c.val, nil
	}

	v, err := c.Lazy.Get(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	c.val = &v
	return v, nil
}
