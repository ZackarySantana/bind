package bind

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
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

		// Create one instance of the value to fill.
		val := reflect.New(fv.Type()).Interface()

		var filled bool
		for _, cand := range fb.Candidates {
			supplier := getSupplier(suppliers, cand.Kind)
			if supplier == nil {
				continue
			}

			ok, err := supplier.Fill(ctx, cand.Value, cand.Options, val)
			if err != nil {
				return fmt.Errorf("fill %s (%s): %w", cand.Value, cand.Kind, err)
			}
			if ok {
				fv.Set(reflect.ValueOf(val).Elem())
				filled = true
				break
			}
		}

		if !filled {
			options.logger.DebugContext(ctx, "no supplier matched",
				slog.String("field", dstReflectType.Field(fb.FieldIndex).Name),
				slog.String("dst_struct_type", dstReflectType.Name()),
			)
			if fb.Options.Required {
				return fmt.Errorf("required field '%s' (type '%s') not found",
					dstReflectType.Field(fb.FieldIndex).Name,
					dstReflectType.Field(fb.FieldIndex).Type.String(),
				)
			}
		}
	}
	return nil
}

func getSupplier(suppliers []Supplier, kind string) Supplier {
	for _, s := range suppliers {
		if s.IsKind(kind) {
			return s
		}
	}
	return nil
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
