package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/zackarysantana/bind"
	"github.com/zackarysantana/bind/parser"
)

const (
	TagSQL      = "sql"
	TagExactSQL = "exact-sql"
)

// NewSupplier creates a Supplier that looks up rows from the given SQL table.
// It uses the fields of the ref to build a WHERE clause for the lookup.
// An example struct tag usage is:
//
//	sql:"id=ID,name=Name"
//
// This would use the ID and Name fields of the ref struct to build a query like:
//
//	SELECT <struct fields> FROM users WHERE id = ? AND name = ? LIMIT 1
//
// The types of the fields will be passed as query parameters. The ref should be
// a pointer to the struct. The result row will be scanned into the output struct T.
func NewSupplier[T any](db *sql.DB, table string, ref any) (bind.Supplier, error) {
	return bind.NewSelfSupplier(func(ctx context.Context, filter map[string]any) (T, error) {
		// TODO: Meta options should allow for projections of specific fields.
		metaOptions := bind.GetMetaOptions(ctx)
		fmt.Println("Meta options:", metaOptions)
		projection := "*"
		// TODO: Use metaOptions to project specific fields.

		where, args := whereFromFilter(filter)
		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 1", projection, table, where)
		return execute[T](ctx, db, query, args...)
	}, TagSQL, ref)
}

// NewExactSupplier creates a Supplier that looks up rows from the given SQL table.
// It uses the options provided to build a WHERE clause for the lookup.
// An example struct tag usage is:
//
//	exact-sql:"id=Int(42),name=Name,age=Int(30)"
//
// This would build a query like:
//
//	SELECT <struct fields> FROM users WHERE id = 42 AND name = ? AND age = 30 LIMIT 1
//
// There are built-in parsers for Int32, Int64, Float64, Bool, TimeRFC3339 and
// their Slice variants. You can register your own parsers with RegisterParser.
// The result row will be scanned into the output struct T.
func NewExactSupplier[T any](db *sql.DB, table string) bind.Supplier {
	return bind.NewFuncSupplier(func(ctx context.Context, name string, options []string) (any, error) {
		// The 'name' is also an option.
		options = append(options, name)
		// TODO: Meta options should allow for projections of specific fields.
		metaOptions := bind.GetMetaOptions(ctx)
		fmt.Println("Meta options:", metaOptions)
		projection := "*"
		// TODO: Use metaOptions to project specific fields.

		filter, err := parser.BuildFilter(options)
		if err != nil {
			return nil, fmt.Errorf("building filter from options: %w", err)
		}
		where, args := whereFromFilter(filter)
		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 1", projection, table, where)
		return execute[T](ctx, db, query, args...)
	}, TagExactSQL)
}

func whereFromFilter(filter map[string]any) (string, []any) {
	clauses := make([]string, 0, len(filter))
	args := make([]any, 0, len(filter))
	for k, v := range filter {
		clauses = append(clauses, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}
	return strings.Join(clauses, " AND "), args
}

func execute[T any](ctx context.Context, db *sql.DB, query string, args ...any) (T, error) {
	var out T
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return out, fmt.Errorf("querying row: %w", err)
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return out, fmt.Errorf("iterating to next row: %w", err)
		}
		return out, sql.ErrNoRows
	}

	cols, err := rows.Columns()
	if err != nil {
		return out, err
	}

	dest, setters, err := buildScanDests(&out, cols)
	if err != nil {
		return out, err
	}
	if err := rows.Scan(dest...); err != nil {
		return out, err
	}
	for _, set := range setters {
		if err := set(); err != nil {
			return out, err
		}
	}
	return out, rows.Err()
}

// buildScanDests maps result columns to struct fields of *out by `db:"col"`
// (preferred) or lowercased field name. It returns []any for rows.Scan and a
// set of post-scan setters for conversions (e.g., RawBytes -> string).
func buildScanDests(out any, cols []string) ([]any, []func() error, error) {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return nil, nil, errors.New("scan target must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	index := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			// Unexported field, skip
			continue
		}
		if tag, ok := sf.Tag.Lookup("db"); ok && tag != "" && tag != "-" {
			index[tag] = i
		} else {
			index[strings.ToLower(sf.Name)] = i
		}
	}

	dest := make([]any, len(cols))
	var setters []func() error

	for i, col := range cols {
		key := col
		if _, ok := index[key]; !ok {
			key = strings.ToLower(col)
		}
		fi, ok := index[key]
		if !ok {
			// Column has no matching field, scan into dummy.
			var dummy any
			dest[i] = &dummy
			continue
		}

		fv := v.Field(fi)
		ptr := fv.Addr().Interface()

		switch ptr.(type) {
		case *string, *int, *int64, *int32, *float64, *bool, *[]byte,
			*sql.NullString, *sql.NullInt64, *sql.NullFloat64, *sql.NullBool, *sql.NullTime:
			// Types `database/sql` handles:
			dest[i] = ptr
		default:
			// Other types we scan into RawBytes and then convert.
			var raw sql.RawBytes
			dest[i] = &raw

			setters = append(setters, func(fi int, rawPtr *sql.RawBytes, fv reflect.Value) func() error {
				return func() error {
					if rawPtr == nil {
						return nil
					}
					switch fv.Kind() {
					case reflect.String:
						fv.SetString(string(*rawPtr))
						return nil
					case reflect.Slice:
						if fv.Type().Elem().Kind() == reflect.Uint8 {
							fv.SetBytes(append([]byte(nil), *rawPtr...))
							return nil
						}
					}
					//TODO: Extend here to handle JSON -> struct, time parsing, etc.
					return nil
				}
			}(fi, &raw, fv))
		}
	}
	return dest, setters, nil
}
