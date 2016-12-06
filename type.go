package conf

import (
	"encoding"
	"encoding/json"
	"reflect"
	"time"
)

// The makeType dynamically rebuilds a Go type, replacing values that are not
// serializable to JSON with equivalents that can.
//
// This has proven useful to handle the time.Duration type which doesn't
// implement json.Unmarshaler.
func makeType(t reflect.Type) reflect.Type {
	switch {
	case t.Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()):
		return t

	case t.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()):
		return t

	case t == reflect.TypeOf(time.Time{}):
		return t

	case t == reflect.TypeOf(time.Duration(0)):
		return reflect.TypeOf(duration(0))
	}

	switch t.Kind() {
	case reflect.Struct:
		return makeStructType(t)

	case reflect.Map:
		return reflect.MapOf(makeType(t.Key()), makeType(t.Elem()))

	case reflect.Slice:
		return reflect.SliceOf(makeType(t.Elem()))

	case reflect.Array:
		return reflect.ArrayOf(t.Len(), makeType(t.Elem()))

	case reflect.Ptr:
		return reflect.PtrTo(makeType(t.Elem()))

	default:
		return t
	}
}

func makeStructType(t reflect.Type) reflect.Type {
	n := t.NumField()
	f := make([]reflect.StructField, n)

	for i := 0; i != n; i++ {
		f[i] = makeStructField(t.Field(i))
	}

	return reflect.StructOf(f)
}

func makeStructField(f reflect.StructField) reflect.StructField {
	return reflect.StructField{
		Name:      f.Name,
		PkgPath:   f.PkgPath,
		Type:      makeType(f.Type),
		Tag:       f.Tag,
		Anonymous: f.Anonymous,
	}
}

// This type is used to replace time.Duration when building the reflective
// representation of a configuration object because time.Duration doesn't
// support the json.Unmarshaler interface.
type duration time.Duration

func (d *duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	*d = duration(v)
	return err
}

func (d duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}
