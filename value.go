package conf

import (
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"time"
)

// Creates a copy of v1 as a dynamically generated configuration value usable
// for loading configs.
func makeConfValue(v1 reflect.Value) reflect.Value {
	v2 := reflect.New(makeType(v1.Type())).Elem()
	setValue(v2, v1)
	return v2
}

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
	t1 := v1.Type()
	v3 := reflect.Value{}

	switch x := v2.Interface().(type) {
	case Duration:
		v3 = convertConfDuration(t1, x)

	case NetAddr:
		v3 = convertConfNetAddr(t1, x)

	case URL:
		v3 = convertConfURL(t1, x)

	case Email:
		v3 = convertConfEmail(t1, x)

	case time.Duration:
		v3 = convertDuration(t1, x)

	case time.Time:
		v3 = v2

	case net.TCPAddr:
		v3 = convertTCPAddr(t1, x)

	case net.UDPAddr:
		v3 = convertUDPAddr(t1, x)

	case url.URL:
		v3 = convertURL(t1, x)

	case mail.Address:
		v3 = convertMailAddress(t1, x)
	}

	if v3.IsValid() {
		v1.Set(v3)
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

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func convertConfDuration(t reflect.Type, d Duration) (v reflect.Value) {
	switch t {
	case timeDurationType:
		v = reflect.ValueOf(time.Duration(d))
	}
	return
}

func convertConfNetAddr(t reflect.Type, a NetAddr) (v reflect.Value) {
	switch t {
	case netTCPAddrType:
		v = reflect.ValueOf(net.TCPAddr{a.IP, a.Port, a.Zone})
	case netUDPAddrType:
		v = reflect.ValueOf(net.UDPAddr{a.IP, a.Port, a.Zone})
	}
	return
}

func convertConfURL(t reflect.Type, u URL) (v reflect.Value) {
	switch t {
	case urlURLType:
		v = reflect.ValueOf(url.URL(u))
	}
	return
}

func convertConfEmail(t reflect.Type, e Email) (v reflect.Value) {
	switch t {
	case mailAddressType:
		v = reflect.ValueOf(mail.Address(e))
	}
	return
}

func convertDuration(t reflect.Type, d time.Duration) (v reflect.Value) {
	switch t {
	case confDurationType:
		v = reflect.ValueOf(Duration(d))
	}
	return
}

func convertTCPAddr(t reflect.Type, a net.TCPAddr) (v reflect.Value) {
	switch t {
	case confNetAddrType:
		v = reflect.ValueOf(NetAddr{a.IP, a.Port, a.Zone})
	}
	return
}

func convertUDPAddr(t reflect.Type, a net.UDPAddr) (v reflect.Value) {
	switch t {
	case confNetAddrType:
		v = reflect.ValueOf(NetAddr{a.IP, a.Port, a.Zone})
	}
	return
}

func convertURL(t reflect.Type, u url.URL) (v reflect.Value) {
	switch t {
	case confURLType:
		v = reflect.ValueOf(URL(u))
	}
	return
}

func convertMailAddress(t reflect.Type, a mail.Address) (v reflect.Value) {
	switch t {
	case confEmailType:
		v = reflect.ValueOf(Email(a))
	}
	return
}
