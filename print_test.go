package conf

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestPrettyType(t *testing.T) {
	tests := []struct {
		v interface{}
		s string
	}{
		{nil, "<nil>"},
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
		{[]byte{}, "bytes"},

		{[]int{}, "list"},
		{[1]int{}, "list"},

		{map[int]int{}, "object"},
		{struct{}{}, "object"},
	}

	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			if s := prettyType(test.v); s != test.s {
				t.Error(s)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	ld := Loader{}
	b := &bytes.Buffer{}

	ld.FprintError(b, errors.New("A: missing value"))

	const txt = "Error:\n  A: missing value\n"

	if s := b.String(); s != txt {
		t.Error(s)
	}
}

func TestPrintHelp(t *testing.T) {
	ld := Loader{
		Args:     []string{"-A=1", "-B=2", "-C=3"},
		Program:  "test",
		FileFlag: "F",
	}
	b := &bytes.Buffer{}

	ld.FprintHelp(b, struct {
		A int
		B int
		C int
		D bool `help:"Set D"`
		E bool `json:"enable" help:"Enable E"`
	}{A: 1})

	const txt = "Usage of test:\n" +
		"  -A int\n" +
		"    \t(default 1)\n" +
		"  -B int\n" +
		"  -C int\n" +
		"  -D\tSet D\n" +
		"  -F string\n" +
		"    \tPath to the configuration file\n" +
		"  -enable\n" +
		"    \tEnable E\n"

	if s := b.String(); s != txt {
		t.Error(s)
	}
}
