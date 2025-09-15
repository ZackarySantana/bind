package bind

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

// setStringInto attempts to set the string value in to the given reflect.Value.
// If the reflect.Value is a type that string can parse in to, e.g. int, float, bool, etc
// it will attempt to parse it.
func setStringInto(p reflect.Value, s string) error {
	// p is a pointer; walk through pointers, allocating as needed.
	t := p.Type().Elem()
	if t.Kind() == reflect.Ptr {
		if p.Elem().IsNil() {
			p.Elem().Set(reflect.New(t.Elem()))
		}
		return setStringInto(p.Elem(), s)
	}

	// Try TextUnmarshaler on *T (common for custom types).
	if tu, ok := p.Interface().(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText([]byte(s))
	}

	ev := p.Elem()
	switch ev.Kind() {
	case reflect.String:
		ev.SetString(s)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		ev.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, int(ev.Type().Bits()))
		if err != nil {
			return err
		}
		ev.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, int(ev.Type().Bits()))
		if err != nil {
			return err
		}
		ev.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, int(ev.Type().Bits()))
		if err != nil {
			return err
		}
		ev.SetFloat(f)
		return nil
	default:
		return fmt.Errorf("unsupported kind %s for path binding (implement encoding.TextUnmarshaler to support custom types)", ev.Kind())
	}
}
