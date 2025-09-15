package bind

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Store[T any] func(ctx context.Context, filter map[string]any) (T, error)

func NewSelfSupplier[T any](store Store[T], kind string, self any) (*StoreSupplier[T], error) {
	if self == nil {
		return nil, errors.New("self is nil")
	}
	fieldVals, err := topLevelFieldMap(self)
	if err != nil {
		return nil, fmt.Errorf("getting top level fields from struct: %w", err)
	}
	return &StoreSupplier[T]{
		store:     store,
		kind:      kind,
		fieldVals: fieldVals,
	}, nil
}

type StoreSupplier[T any] struct {
	store Store[T]
	kind  string
	// a struct or pointer to struct whose fields can be used in filters
	fieldVals map[string]reflect.Value
}

func (b *StoreSupplier[T]) Fill(ctx context.Context, value string, options []string, dst any) (bool, error) {
	options = append(options, value)
	filter := map[string]any{}

	for _, option := range options {
		var key, fieldName string
		if split := strings.Split(option, "="); len(split) == 2 {
			key = strings.TrimSpace(split[0])
			fieldName = strings.TrimSpace(split[1])
		} else {
			continue
		}

		var fieldValue any
		if fv, ok := b.fieldVals[fieldName]; ok {
			fieldValue = fv.Interface()
		} else {
			return false, fmt.Errorf("field %q not found in struct", fieldName)
		}
		filter[key] = fieldValue
	}

	val, err := b.store(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("error finding document: %w", err)
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false, fmt.Errorf("dst must be a non-nil pointer")
	}
	rv = rv.Elem()
	if reflect.TypeOf(val) != rv.Type() {
		return false, fmt.Errorf("type mismatch: store returned %T but dst is %T", val, dst)
	}
	rv.Set(reflect.ValueOf(val))

	return true, nil
}

func (b *StoreSupplier[T]) IsKind(kind string) bool {
	return b.kind == kind
}

// topLevelFieldMap returns a map of exported top-level field names to their interface values.
// If b.val is a pointer, it is dereferenced. Only struct fields are considered; unexported are skipped.
func topLevelFieldMap(src any) (map[string]reflect.Value, error) {
	v := reflect.ValueOf(src)
	if !v.IsValid() {
		return nil, errors.New("source value (b.val) is invalid")
	}
	// Deref pointers
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, errors.New("source pointer (b.val) is nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source must be a struct or pointer to struct; got %s", v.Kind())
	}

	t := v.Type()
	out := make(map[string]reflect.Value, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// skip unexported
		if f.PkgPath != "" {
			continue
		}
		fv := v.Field(i)
		// safely get interface; if it can't, skip
		if fv.IsValid() && fv.CanInterface() {
			out[f.Name] = fv
		}
	}
	return out, nil
}
