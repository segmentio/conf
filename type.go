package conf

import (
	"encoding"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/segmentio/objconv"
)

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	timeTimeType     = reflect.TypeOf(time.Time{})
	netTCPAddrType   = reflect.TypeOf(net.TCPAddr{})
	netUDPAddrType   = reflect.TypeOf(net.UDPAddr{})
	urlURLType       = reflect.TypeOf(url.URL{})
	mailAddressType  = reflect.TypeOf(mail.Address{})
	specialValueType = reflect.TypeOf(specialValue{})

	textUnmarshalerInterface     = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	objconvValueDecoderInterface = reflect.TypeOf((*objconv.ValueDecoder)(nil)).Elem()
)

// The makeType dynamically rebuilds a Go type, replacing values that are not
// serializable to JSON with equivalents that can.
//
// This has proven useful to handle the time.Duration type which doesn't
// implement json.Unmarshaler.
func makeType(t reflect.Type) reflect.Type {
	if specialType(t) {
		return specialValueType
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

func specialType(t reflect.Type) bool {
	switch {
	case t.Implements(objconvValueDecoderInterface), t.Implements(textUnmarshalerInterface):
		return true
	}
	switch t {
	case timeTimeType, timeDurationType, netTCPAddrType, netUDPAddrType, urlURLType, mailAddressType:
		return true
	}
	return false
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
		Tag:       reflect.StructTag(strings.Replace(string(f.Tag), "conf:", "objconv:", -1)),
		Anonymous: f.Anonymous,
	}
}
