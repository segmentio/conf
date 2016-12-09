package conf

import (
	"encoding"
	"encoding/json"
	"errors"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// This Duration type is intended to be used in configuration structs to provide
// automatic loading an type checking of duration values.
//
// The main need for this type comes from the fact that the standard
// time.Duration doesn't implement encoding.TextUnmarshaler or json.Unmarshaler.
//
// Note that the use of this type is optional since the package will implicitly
// do the conversion between time.Duration and conf.Duration values.
type Duration time.Duration

// String satisifies the fmt.Stringer interface.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// MarshalText satisifies the encoding.TextMarshaler interface.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalText satisifies the encoding.TextUnmarshaler interface.
func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	*d = Duration(v)
	return err
}

// Email is a type alisa to mail.Address that implements the
// encoding.TextMarshaler and encoding.TextUnmarshaler interfaces so it can be
// loaded from the program configuration.
//
// Note that the use of this type is optional since the package will implicitly
// do the conversion between url.URL and conf.URL values.
type Email mail.Address

// MarshalText satisifies the encoding.TextMarshaler interface.
func (e Email) MarshalText() ([]byte, error) {
	a := mail.Address(e)
	return []byte(a.String()), nil
}

// UnmarshalText satisifies the encoding.TextUnmarshaler interface.
func (e *Email) UnmarshalText(b []byte) error {
	a, err := mail.ParseAddress(string(b))
	if a != nil {
		*e = Email(*a)
	}
	return err
}

// URL is a type alias to url.URL that implements the encoding.TextMarshaler and
// encoding.TextUnmarshaler interfaces so it can be loaded from the program
// configuration.
//
// Note that the use of this type is optional since the package will implicitly
// do the conversion between url.URL and conf.URL values.
type URL url.URL

// MarshalText satisifies the encoding.TextMarshaler interface.
func (u URL) MarshalText() ([]byte, error) {
	v := url.URL(u)
	return []byte(v.String()), nil
}

// UnmarshalText satisifies the encoding.TextUnmarshaler interface.
func (u *URL) UnmarshalText(b []byte) error {
	v, err := url.Parse(string(b))
	if v != nil {
		*u = URL(*v)
	}
	return err
}

// NetAddr represents a network address which can be loaded from a program
// configuration.
//
// Note that the use of this type is optional since the package will implicitly
// do the conversion between net.TCPAddr or net.UDPAddr and conf.NetAddr.
type NetAddr struct {
	IP   net.IP // IPv4 or IPv6 address
	Port int    // the port number
	Zone string // the IPv6 address zone
}

// ParseNetAddr parses a network address representation from s into a.
func ParseNetAddr(s string) (NetAddr, error) {
	var a NetAddr
	var h string
	var p string
	var err error

	if h, p, err = net.SplitHostPort(s); err != nil {
		h, p = s, ""
	}

	if len(h) != 0 {
		if off := strings.IndexByte(h, '%'); off >= 0 {
			h, a.Zone = h[:off], h[off+1:]
		}
		if a.IP = net.ParseIP(h); a.IP == nil {
			return NetAddr{}, errors.New(s + ": bad IP address")
		}
	}

	if len(p) != 0 {
		if a.Port, err = strconv.Atoi(p); err != nil || a.Port < 0 || a.Port > 65535 {
			return NetAddr{}, errors.New(s + ": bad port number")
		}
	}

	return a, nil
}

// Network returns the network type used by the address, it always is an empty
// string because NetAddr may represent TCP or UDP addresses.
//
// This method matches the net.Addr interface.
func (a NetAddr) Network() string {
	return ""
}

// String returns a string representation of the address, which is usually made
// of a host and port separated by a column.
//
// This method matches the net.Addr interface.
func (a NetAddr) String() string {
	var host string
	var port string

	if len(a.IP) != 0 {
		host = a.IP.String()

		if len(a.Zone) != 0 {
			host += "%" + a.Zone
		}
	}

	if a.Port != 0 {
		port = strconv.Itoa(a.Port)
	}

	switch {
	case len(host) == 0:
		return ":" + port

	case len(port) == 0:
		return host

	default:
		return net.JoinHostPort(host, port)
	}
}

// MarshalText satisifies the encoding.TextMarshaler interface.
func (a NetAddr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText satisifies the encoding.TextUnmarshaler interface.
func (a *NetAddr) UnmarshalText(b []byte) (err error) {
	*a, err = ParseNetAddr(string(b))
	return
}

// The makeType dynamically rebuilds a Go type, replacing values that are not
// serializable to JSON with equivalents that can.
//
// This has proven useful to handle the time.Duration type which doesn't
// implement json.Unmarshaler.
func makeType(t reflect.Type) reflect.Type {
	switch {
	case t.Implements(jsonUnmarshalerType):
		return t

	case t.Implements(textUnmarshalerType):
		return t
	}

	switch t {
	case timeTimeType:
		return t

	case timeDurationType:
		return confDurationType

	case netTCPAddrType:
		return confNetAddrType

	case netUDPAddrType:
		return confNetAddrType

	case urlURLType:
		return confURLType

	case mailAddressType:
		return confEmailType
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
		Tag:       reflect.StructTag(strings.Replace(string(f.Tag), "conf:", "json:", -1)),
		Anonymous: f.Anonymous,
	}
}

// Cache of various types used by the package.
var (
	// encoding
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	// time
	timeTimeType     = reflect.TypeOf(time.Time{})
	timeDurationType = reflect.TypeOf(time.Duration(0))
	confDurationType = reflect.TypeOf(Duration(0))

	// net
	netTCPAddrType  = reflect.TypeOf(net.TCPAddr{})
	netUDPAddrType  = reflect.TypeOf(net.UDPAddr{})
	confNetAddrType = reflect.TypeOf(NetAddr{})

	// net/url
	urlURLType  = reflect.TypeOf(url.URL{})
	confURLType = reflect.TypeOf(URL{})

	// net/mail
	mailAddressType = reflect.TypeOf(mail.Address{})
	confEmailType   = reflect.TypeOf(Email{})
)
