package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Parser[T any] func(string) (T, error)
type dumbParser func(string) (any, error)

var parsers = map[string]dumbParser{}

func init() {
	Register("String", func(s string) (string, error) {
		return s, nil
	})
	Register("Int", func(s string) (int, error) {
		return strconv.Atoi(s)
	})
	Register("Int32", func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
	Register("Int64", func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	Register("Float64", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	Register("Bool", func(s string) (bool, error) {
		return strconv.ParseBool(s)
	})
	Register("Time", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339, s)
	})
}

// Register adds a new Parser for the given type name.
// It also registers `{typeName}Slice` for parsing comma-separated values of that type.
func Register[T any](typeName string, parser Parser[T]) {
	parsers[typeName] = func(s string) (any, error) {
		return parser(s)
	}
	sliceParser := asSliceParser(parser)
	parsers[typeName+"Slice"] = func(s string) (any, error) {
		return sliceParser(s)
	}
}

func asSliceParser[T any](parser Parser[T]) Parser[[]T] {
	return func(s string) ([]T, error) {
		parts := strings.Split(s, ",")
		out := make([]T, 0, len(parts))
		for _, part := range parts {
			value, err := parser(part)
			if err != nil {
				return nil, fmt.Errorf("parsing slice element %q: %w", part, err)
			}
			out = append(out, value)
		}
		return out, nil
	}
}

func parseValue(s string) (any, error) {
	i := strings.IndexByte(s, '(')
	if i <= 0 || !strings.HasSuffix(s, ")") {
		return s, nil
	}
	typ := s[:i]
	in := s[i+1 : len(s)-1] // inside parens
	if p, ok := parsers[typ]; ok {
		v, err := p(in)
		return v, err
	}
	return s, nil
}

// BuildFilter builds a filter map from the given options.
// It parses each string as `key=TypeName(value)` or for strings `key=value`.
// Built-in types are Int, Int32, Int64, Float64, Bool, Time (RFC3339 format)
// and their Slice variants, e.g. `IntSlice(0,-1,5)`.
// New types can be registered with Register.
func BuildFilter(options []string) (map[string]any, error) {
	filter := make(map[string]any)
	for _, opt := range options {
		parts := strings.SplitN(opt, "=", 2)
		if len(parts) == 2 {
			parsedValue, err := parseValue(parts[1])
			if err != nil {
				return nil, fmt.Errorf("parsing %s from %q: %w", parts[0], parts[1], err)
			}
			filter[parts[0]] = parsedValue
		} else {
			return nil, fmt.Errorf("invalid filter option %q", opt)
		}
	}
	return filter, nil
}
