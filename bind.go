package bind

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
)

// Supplier fills a specific field from a named key (per tag kind).
type Supplier interface {
	// Fill tries to set dst from the supplier using `value`.
	// Returns (true, nil) if it set the value, (false, nil) if not present, or error.
	Fill(ctx context.Context, value string, options []string, dst any) (bool, error)
	// IsKind reports whether this supplier matches the given kind.
	IsKind(kind string) bool
}

// Bind fills the struct pointed to by dst using the given suppliers.
func Bind(ctx context.Context, dst any, suppliers []Supplier, opts ...Option) error {
	options := &options{
		logger: slog.New(slog.DiscardHandler),
		level:  DefaultLevel,
	}

	for _, opt := range opts {
		opt(options)
	}

	dstReflectValue, dstReflectType, err := requireStructPtr(dst)
	if err != nil {
		return err
	}

	fields := collectFieldBindings(dstReflectType)
	for _, fb := range fields {
		// If the minimum level is higher, skip this field.
		if fb.Options.MinimumLevel > options.level {
			continue
		}
		// The reflect.Value of the field to fill.
		fv := dstReflectValue.Field(fb.FieldIndex)

		// Check if the field was already set.
		if !fv.IsZero() {
			options.logger.DebugContext(ctx, "field already set, skipping",
				slog.String("field", dstReflectType.Field(fb.FieldIndex).Name),
				slog.String("dst_struct_type", dstReflectType.Name()),
			)
			continue
		}

		val := reflect.New(fv.Type()).Interface()

		if f, ok := lazyFactories.Load(fv.Type()); ok {
			var loader LazyLoader = func(ctx context.Context, t any) error {
				_, err := applySuppliers(ctx, suppliers, dstReflectType, fb, t, options)
				return err
			}
			v := f.(func(any) any)(loader)

			fv.Set(reflect.ValueOf(v))
			continue
		}

		if f, ok := cacheFactories.Load(fv.Type()); ok {
			var loader LazyLoader = func(ctx context.Context, t any) error {
				_, err := applySuppliers(ctx, suppliers, dstReflectType, fb, t, options)
				return err
			}
			v := f.(func(any) any)(loader)

			fv.Set(reflect.ValueOf(v))
			continue
		}

		foundVal, err := applySuppliers(ctx, suppliers, dstReflectType, fb, val, options)
		if err != nil {
			if strings.Contains(err.Error(), "bind.Lazy[") {
				err = fmt.Errorf("field '%s' (type '%s') uses unregistered Lazy type; did you forget to call bind.RegisterLazy?",
					dstReflectType.Field(fb.FieldIndex).Name,
					dstReflectType.Field(fb.FieldIndex).Type.String(),
				)
			}
			return err
		}
		if foundVal == nil {
			continue
		}

		fv.Set(*foundVal)
	}
	return nil
}

func applySuppliers(ctx context.Context, suppliers []Supplier, fullValue reflect.Type, fb FieldBinding, val any, options *options) (*reflect.Value, error) {
	ctx = setMetaOptions(ctx, fb.Options.Options)
	for _, cand := range fb.Candidates {
		sups := getSuppliers(suppliers, cand.Kind)
		if len(sups) == 0 {
			if !options.testOnly {
				continue
			}
			// In test mode, try all suppliers.
			sups = suppliers
		}

		for _, supplier := range sups {
			ok, err := supplier.Fill(ctx, cand.Value, cand.Options, val)
			if err != nil {
				return nil, fmt.Errorf("fill %s (%s): %w", cand.Value, cand.Kind, err)
			}
			if ok {
				val := reflect.ValueOf(val).Elem()
				return &val, nil
			}
		}
	}

	options.logger.DebugContext(ctx, "no supplier matched",
		slog.String("field", fullValue.Field(fb.FieldIndex).Name),
		slog.String("dst_struct_type", fullValue.Name()),
	)
	if fb.Options.Required {
		return nil, fmt.Errorf("required field '%s' (type '%s') not found",
			fullValue.Field(fb.FieldIndex).Name,
			fullValue.Field(fb.FieldIndex).Type.String(),
		)
	}

	return nil, nil
}

func getSuppliers(suppliers []Supplier, kind string) []Supplier {
	var result []Supplier
	for _, s := range suppliers {
		if s.IsKind(kind) {
			result = append(result, s)
		}
	}
	return result
}

var (
	ErrNilDestination = errors.New("dst is nil")
	ErrNonPointer     = errors.New("dst is not a pointer")
	ErrNonStructPtr   = errors.New("dst is not a struct pointer")
)

func requireStructPtr(dst any) (reflect.Value, reflect.Type, error) {
	if dst == nil {
		return reflect.Value{}, nil, ErrNilDestination
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, nil, ErrNonPointer
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, nil, ErrNonStructPtr
	}
	return rv, rv.Type(), nil
}
