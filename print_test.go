package conf

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type Bytes uint64

func TestPrettyType(t *testing.T) {
	tests := []struct {
		v interface{}
		s string
	}{
		{nil, "unknown"},
		{false, "bool"},

		{int(0), "int"},
		{int8(0), "int8"},
		{int16(0), "int16"},
		{int32(0), "int32"},
		{int64(0), "int64"},

		{uint(0), "uint"},
		{uint8(0), "uint8"},
		{uint16(0), "uint16"},
		{uint32(0), "uint32"},
		{uint64(0), "uint64"},

		{float32(0), "float32"},
		{float64(0), "float64"},

		{time.Duration(0), "duration"},
		{time.Time{}, "time"},

		{"", "string"},
		{[]byte{}, "base64"},

		{[]int{}, "list"},
		{[1]int{}, "list"},

		{map[int]int{}, "object"},
		{struct{}{}, "object"},
		{&struct{}{}, "object"},

		{Bytes(0), "bytes"},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			if s := prettyType(reflect.TypeOf(test.v)); s != test.s {
				t.Error(s)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	ld := Loader{}
	b := &bytes.Buffer{}

	ld.FprintError(b, errors.New("A: missing value"))

	const txt = "Error:\n  A: missing value\n\n"

	if s := b.String(); s != txt {
		t.Error(s)
	}
}

func TestPrintHelp(t *testing.T) {
	ld := Loader{
		Name:     "test",
		Args:     []string{"-A=1", "-B=2", "-C=3"},
		Commands: []Command{{"run", "Run something"}, {"version", "Print the version"}},
	}
	b := &bytes.Buffer{}

	ld.FprintHelp(b, struct {
		A int
		B int
		C int
		D bool `help:"Set D"`
		E bool `conf:"enable" help:"Enable E"`
		F bool `help:"run in super mode"`
		T time.Duration
	}{A: 1, F: true, T: time.Second})

	const txt = "Usage:\n" +
		"  test [command] [options...]\n" +
		"\n" +
		"Commands:\n" +
		"  run      Run something\n" +
		"  version  Print the version\n" +
		"\n" +
		"Options:\n" +
		"  -A int\n" +
		"    \t(default 1)\n" +
		"\n" +
		"  -B int\n" +
		"\n" +
		"  -C int\n" +
		"\n" +
		"  -D\tSet D\n" +
		"\n" +
		"  -F\trun in super mode (default true)\n" +
		"\n" +
		"  -T duration\n" +
		"    \t(default 1s)\n" +
		"\n" +
		"  -enable\n" +
		"    \tEnable E\n" +
		"\n"

	if diff := cmp.Diff(txt, b.String()); diff != "" {
		t.Error(diff)
	}
}
