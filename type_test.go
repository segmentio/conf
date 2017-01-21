package conf

import (
	"reflect"
	"testing"
)

func TestFieldPath(t *testing.T) {
	tests := []struct {
		value  interface{}
		input  string
		output string
	}{
		{
			value:  struct{}{},
			input:  "",
			output: "",
		},
		{
			value:  struct{ A int }{},
			input:  "A",
			output: "A",
		},
		{
			value:  struct{ A int }{},
			input:  "1.2.3",
			output: "1.2.3",
		},
		{
			value: struct {
				A int `conf:"a"`
			}{},
			input:  "A",
			output: "a",
		},
		{
			value: struct {
				A int `conf:"a"`
			}{},
			input:  "a",
			output: "a",
		},
		{
			value: struct {
				A struct {
					B struct {
						C int `conf:"c"`
					} `conf:"b"`
				} `conf:"a"`
			}{},
			input:  "A.B.C",
			output: "a.b.c",
		},
		{
			value: struct {
				A struct {
					B struct {
						C int `conf:"c"`
					} `conf:"b"`
				} `conf:"a"`
			}{},
			input:  "A.B",
			output: "a.b",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if output := fieldPath(reflect.TypeOf(test.value), test.input); output != test.output {
				t.Error(output)
			}
		})
	}
}
