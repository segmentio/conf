package conf

import (
	"encoding"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestTypes(t *testing.T) {
	for _, v := range [...]interface{}{
		Duration(0),
		Duration(1 * time.Second),
		Duration(1 * time.Hour),

		NetAddr{},
		NetAddr{IP: net.IPv4(127, 0, 0, 1)},
		NetAddr{Port: 80},
		NetAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80},
		NetAddr{IP: net.ParseIP("::1"), Port: 80, Zone: "11"},
	} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			testType(t, v)
		})
	}
}

func testType(t *testing.T, v interface{}) {
	var m encoding.TextMarshaler
	var u encoding.TextUnmarshaler
	var w interface{}

	switch x := v.(type) {
	case encoding.TextMarshaler:
		m = x
	default:
		t.Errorf("%#v doesn't implement encoding.TextMarshaler", v)
		return
	}

	w = reflect.New(reflect.TypeOf(v)).Interface()

	switch x := w.(type) {
	case encoding.TextUnmarshaler:
		u = x
	default:
		t.Errorf("%#v doesn't implement encoding.TextUnmarshaler", v)
		return
	}

	b, err := m.MarshalText()
	if err != nil {
		t.Errorf("%#v: MarshalText: %s", err)
		return
	}

	if err := u.UnmarshalText(b); err != nil {
		t.Errorf("%#v: UnmarshalText: %s", err)
		return
	}

	w = reflect.ValueOf(w).Elem().Interface()

	if !reflect.DeepEqual(v, w) {
		t.Errorf("\n<<< %#v\n>>> %#v", v, w)
		return
	}
}
