package conf

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	type point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	tests := []struct {
		val interface{}
		env []string
	}{
		{
			val: struct{ A bool }{true},
			env: []string{"TEST_A=true"},
		},
		{
			val: struct{ A bool }{false},
			env: []string{"TEST_A=false"},
		},
		{
			val: struct{ A int }{42},
			env: []string{"TEST_A=42"},
		},
		{
			val: struct{ A int }{0},
			env: []string{}, // missing => zero value
		},
		{
			val: struct{ P *point }{&point{1, 2}},
			env: []string{"TEST_P_X=1", "TEST_P_Y=2"},
		},
		{
			val: struct{ P *point }{&point{1, 2}},
			env: []string{"TEST_P={ 'x': 1, 'y': 2 }"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()

			if err := loadEnv(v2, "test", test.env); err != nil {
				t.Error(err)
			}

			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("%#v", x2)
			}
		})
	}
}

func TestLoadArgs(t *testing.T) {
	type point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	tests := []struct {
		val  interface{}
		args []string
	}{
		{
			val:  struct{ A bool }{true},
			args: []string{"-A"},
		},
		{
			val:  struct{ A bool }{false},
			args: []string{},
		},
		{
			val:  struct{ A int }{42},
			args: []string{"-A", "42"},
		},
		{
			val:  struct{ A int }{0},
			args: []string{}, // missing => zero value
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			args: []string{"-P.x", "1", "-P.y", "2"},
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			args: []string{"-P", "{ 'x': 1, 'y': 2 }"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()

			if _, err := loadArgs(v2, "test", "", test.args); err != nil {
				t.Error(err)
			}

			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("%#v", x2)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	type point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	tests := []struct {
		val  interface{}
		file string
	}{
		{
			val:  struct{ A bool }{true},
			file: `A: true`,
		},
		{
			val:  struct{ A bool }{false},
			file: `A: false`,
		},
		{
			val:  struct{ A int }{42},
			file: `A: 42`,
		},
		{
			val:  struct{ A int }{0},
			file: ``, // missing => zero value
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			file: `P: { 'x': 1, 'y': 2 }`,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()
			args := []string{"-config-file", "test.yml"}

			if err := loadFile(v2, "test", "config-file", args, func(file string) (b []byte, err error) {
				if file != "test.yml" {
					t.Error(file)
				}
				b = []byte(test.file)
				return
			}); err != nil {
				t.Error(err)
			}

			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("%#v\n", x2)
			}
		})
	}
}

func TestLoader(t *testing.T) {
	const configFile = "/tmp/conf-test.yml"
	ioutil.WriteFile(configFile, []byte(`---
points:
 - { 'x': 0, 'y': 0 }
 - { 'x': 1, 'y': 2 }
 - { 'x': 42, 'y': 42 }
`), 0644)
	defer os.Remove(configFile)

	loaders := []Loader{
		Loader{
			Program: "test",
			Args:    []string{"-points", `[{'x':0,'y':0},{'x':1,'y':2},{'x':42,'y':42}]`, "A", "B", "C"},
			Env:     []string{},
		},
		Loader{
			Program: "test",
			Args:    []string{"A", "B", "C"},
			Env:     []string{"TEST_POINTS=[{'x':0,'y':0},{'x':1,'y':2},{'x':42,'y':42}]"},
		},
		Loader{
			Program:  "test",
			Args:     []string{"-f", configFile, "A", "B", "C"},
			Env:      []string{},
			FileFlag: "f",
		},
		Loader{
			Program:  "test",
			Args:     []string{"-f", configFile, "-points", `[{'x':0,'y':0},{'x':1,'y':2},{'x':42,'y':42}]`, "A", "B", "C"},
			Env:      []string{"TEST_POINTS=[{'x':0,'y':0},{'x':1,'y':2},{'x':42,'y':42}]"},
			FileFlag: "f",
		},
	}

	type point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	type config struct {
		Points []point `json:"points"`
	}

	for _, ld := range loaders {
		t.Run("", func(t *testing.T) {
			var cfg config
			args, err := ld.Load(&cfg)

			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(args, []string{"A", "B", "C"}) {
				t.Error("bad args:", args)
			}

			if !reflect.DeepEqual(cfg, config{[]point{{0, 0}, {1, 2}, {42, 42}}}) {
				t.Error("bad config:", cfg)
			}
		})
	}
}
