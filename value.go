package conf

import (
	"reflect"
	"time"
)

// Sets v to its zero value.
func setZero(v reflect.Value) {
	v.Set(reflect.Zero(v.Type()))
}

// Performs a deep copy of v2 into v1, converting time.Duration to duration and
// duration to time.Duration.
//
// This function is useful when paired with makeType because it translates
// between values of an original type to their dynamically generated
// counterparts.
func setValue(v1 reflect.Value, v2 reflect.Value) {
	switch x := v2.Interface().(type) {
	case duration:
		v1.Set(reflect.ValueOf(time.Duration(x)))
		return

	case time.Duration:
		v1.Set(reflect.ValueOf(duration(x)))
		return

	case time.Time:
		v1.Set(v2)
		return
	}

	switch v2.Kind() {
	case reflect.Struct:
		setStructValue(v1, v2)

	case reflect.Map:
		setMapValue(v1, v2)

	case reflect.Slice:
		setSliceValue(v1, v2)

	case reflect.Array:
		setArrayValue(v1, v2)

	case reflect.Ptr:
		setPtrValue(v1, v2)

	default:
		v1.Set(v2)
	}
}

func setStructValue(v1 reflect.Value, v2 reflect.Value) {
	n2 := v2.NumField()

	for i := 0; i != n2; i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)
		setValue(f1, f2)
	}
}

func setMapValue(v1 reflect.Value, v2 reflect.Value) {
	t1 := v1.Type()
	v1.Set(reflect.MakeMap(t1))

	for _, k2 := range v2.MapKeys() {
		k1 := reflect.New(t1.Key()).Elem()
		e1 := reflect.New(t1.Elem()).Elem()
		setValue(k1, k2)
		setValue(e1, v2.MapIndex(k2))
		v1.SetMapIndex(k1, e1)
	}
}

func setSliceValue(v1 reflect.Value, v2 reflect.Value) {
	n2 := v2.Len()
	t1 := v1.Type()
	v1.Set(reflect.MakeSlice(t1, n2, n2))

	for i := 0; i != n2; i++ {
		e1 := reflect.New(t1.Elem()).Elem()
		e2 := v2.Index(i)
		setValue(e1, e2)
		v1.Index(i).Set(e1)
	}
}

func setArrayValue(v1 reflect.Value, v2 reflect.Value) {
	n2 := v2.Len()
	t1 := v1.Type()

	for i := 0; i != n2; i++ {
		e1 := reflect.New(t1.Elem()).Elem()
		e2 := v2.Index(i)
		setValue(e1, e2)
		v1.Index(i).Set(e1)
	}
}

func setPtrValue(v1 reflect.Value, v2 reflect.Value) {
	if v2.IsNil() {
		v1.Set(reflect.Zero(v1.Type())) // set nil
		return
	}

	e1 := reflect.New(v1.Type().Elem())
	setValue(e1.Elem(), v2.Elem())
	v1.Set(e1)
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
