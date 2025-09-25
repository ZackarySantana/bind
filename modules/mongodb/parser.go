package mongodb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Parser func(string) (any, error)

var parsers = map[string]Parser{
	"ObjectID": func(s string) (any, error) {
		return bson.ObjectIDFromHex(s)
	},
	"Int": func(s string) (any, error) {
		return strconv.Atoi(s)
	},
	"Int32": func(s string) (any, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	},
	"Int64": func(s string) (any, error) {
		return strconv.ParseInt(s, 10, 64)
	},
	"Float64": func(s string) (any, error) {
		return strconv.ParseFloat(s, 64)
	},
	"Bool": func(s string) (any, error) {
		return strconv.ParseBool(s)
	},
	"TimeRFC3339": func(s string) (any, error) {
		return time.Parse(time.RFC3339, s)
	},
}

func init() {
	for k, v := range parsers {
		parsers[k+"Slice"] = makeSliceParser(v)
	}
}

// RegisterParser adds a new Parser for the given type name.
// It also registers `{typeName}Slice` for parsing comma-separated values of that type.
func RegisterParser(typeName string, parser Parser) {
	parsers[typeName] = parser
	parsers[typeName+"Slice"] = makeSliceParser(parser)
}

// makeSliceParser builds a Parser for comma-separated values.
func makeSliceParser(parser Parser) Parser {
	return func(s string) (any, error) {
		parts := strings.Split(s, ",")
		out := make([]any, 0, len(parts))
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

// parseValue tries TypeName(value) via the registry.
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
