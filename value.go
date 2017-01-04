package conf

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/segmentio/objconv"
	"github.com/segmentio/objconv/json"
	"github.com/segmentio/objconv/yaml"
)

// Creates a copy of v1 as a dynamically generated configuration value usable
// for loading configs.
func makeValue(v1 reflect.Value) reflect.Value {
	var t2 = makeType(v1.Type())
	var v2 reflect.Value

	if t2 == specialValueType {
		v2 = reflect.ValueOf(specialValue{v1})
	} else {
		v2 = reflect.New(t2).Elem()
	}

	setValue(v2, v1)
	return v2
}

// copyValue creates a shallow copy of the value referenced by v.
func copyValue(v reflect.Value) reflect.Value {
	c := reflect.New(v.Type()).Elem()
	c.Set(v)
	return c
}

// Sets v to its zero value.
func setZero(v reflect.Value) {
	v.Set(reflect.Zero(v.Type()))
}

// Performs a deep copy of v2 into v1.
func setValue(v1 reflect.Value, v2 reflect.Value) {
	if v1.Type() == specialValueType {
		v1.Set(reflect.ValueOf(specialValue{copyValue(v2)}))
		return
	}
	if v2.Type() == specialValueType {
		v1.Set(v2.Interface().(specialValue).v)
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

// flagValue is a generic value wrapper that implement the flag.Value interface
// to bridge between the conf and flag packages.
type flagValue struct {
	v reflect.Value
	s string
}

func makeFlagValue(v reflect.Value) flagValue {
	f := flagValue{
		v: v,
		s: makeFlagString(v),
	}
	return f
}

func makeFlagString(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	switch x := v.Interface().(type) {
	case specialValue:
		return makeFlagString(x.v)
	case fmt.Stringer:
		return fmt.Sprint(x)
	default:
		if v.Kind() == reflect.Struct {
			return ""
		}
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func (f flagValue) String() string {
	return f.s
}

func (f flagValue) Set(s string) error {
	return yaml.Unmarshal([]byte(s), f.v.Addr().Interface())
}

func (f flagValue) IsBoolFlag() bool {
	return f.v.IsValid() && f.v.Kind() == reflect.Bool
}

// specialValue is a wrapper for special cases handled by the package that
// augment the default capabilities of the objconv decoder.
type specialValue struct {
	v reflect.Value
}

func (s specialValue) DecodeValue(d objconv.Decoder) error {
	v := s.v.Addr().Interface()

	switch x := v.(type) {
	case *net.TCPAddr:
		ip, port, zone, err := decodeNetAddr(d)
		if err != nil {
			return err
		}
		*x = net.TCPAddr{IP: ip, Port: port, Zone: zone}
		return nil

	case *net.UDPAddr:
		ip, port, zone, err := decodeNetAddr(d)
		if err != nil {
			return err
		}
		*x = net.UDPAddr{IP: ip, Port: port, Zone: zone}
		return nil

	case *url.URL:
		u, err := decodeURL(d)
		if err != nil {
			return err
		}
		*x = *u
		return nil

	case *mail.Address:
		a, err := decodeEmailAddr(d)
		if err != nil {
			return err
		}
		*x = *a
		return nil

	default:
		return d.Decode(v)
	}
}

func decodeURL(d objconv.Decoder) (u *url.URL, err error) {
	var s string
	if err = d.Decode(&s); err != nil {
		return
	}
	return url.Parse(s)
}

func decodeEmailAddr(d objconv.Decoder) (a *mail.Address, err error) {
	var s string
	if err = d.Decode(&s); err != nil {
		return
	}
	return mail.ParseAddress(s)
}

func decodeNetAddr(d objconv.Decoder) (ip net.IP, port int, zone string, err error) {
	var s string
	if err = d.Decode(&s); err != nil {
		return
	}
	return parseNetAddr(s)
}

func parseNetAddr(s string) (ip net.IP, port int, zone string, err error) {
	var h string
	var p string

	if h, p, err = net.SplitHostPort(s); err != nil {
		h, p = s, ""
	}

	if len(h) != 0 {
		if off := strings.IndexByte(h, '%'); off >= 0 {
			h, zone = h[:off], h[off+1:]
		}
		if ip = net.ParseIP(h); ip == nil {
			err = errors.New(s + ": bad IP address")
			return
		}
	}

	if len(p) != 0 {
		if port, err = strconv.Atoi(p); err != nil || port < 0 || port > 65535 {
			err = errors.New(s + ": bad port number")
			return
		}
	}

	return
}
