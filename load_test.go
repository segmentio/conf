package conf

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/mail"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/objconv/yaml"
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

type point struct {
	X int `conf:"x"`
	Y int `conf:"y"`
}

var (
	testTime = time.Date(2016, 12, 6, 1, 1, 42, 123456789, time.UTC)

	loadTests = []struct {
		val  interface{}
		file string
		args []string
		env  []string
	}{
		{
			val:  struct{ A bool }{true},
			file: `A: true`,
			args: []string{"-A"},
			env:  []string{"TEST_A=true"},
		},

		{
			val:  struct{ A bool }{false},
			file: `A: false`,
			args: []string{},
			env:  []string{"TEST_A=false"},
		},

		{
			val:  struct{ A int }{42},
			file: `A: 42`,
			args: []string{"-A", "42"},
			env:  []string{"TEST_A=42"},
		},

		{
			val:  struct{ A int }{0}, // missing => zero value
			file: ``,
			args: []string{},
			env:  []string{},
		},

		{
			val:  struct{ A string }{"42"}, // convert digit sequence to string
			file: `A: '42'`,
			args: []string{"-A", "42"},
			env:  []string{"TEST_A=42"},
		},

		{
			val:  struct{ S string }{"Hello World!"},
			file: `S: Hello World!`,
			args: []string{"-S", "Hello World!"},
			env:  []string{"TEST_S=Hello World!"},
		},

		{
			val:  struct{ L []int }{[]int{1, 2, 3}},
			file: `L: [1, 2, 3]`,
			args: []string{"-L", "[1,2,3]"},
			env:  []string{"TEST_L=[1, 2, 3]"},
		},

		{
			val:  struct{ L []string }{[]string{"A", "42"}},
			file: `L: [A, 42]`,
			args: []string{"-L", "[A, 42]"},
			env:  []string{"TEST_L=[A, 42]"},
		},

		{
			val:  struct{ L []string }{[]string{"A", "B", "C"}},
			file: `L: [A,B,C]`,
			args: []string{"-L", "[A,B,C]"},
			env:  []string{"TEST_L=[A,B,C]"},
		},

		{
			val:  struct{ L []string }{[]string{"A", "B", "C"}},
			file: `L: [A,B,C]`,
			args: []string{"-L", `["A","B","C"]`},
			env:  []string{`TEST_L=["A","B","C"]`},
		},

		{
			val:  struct{ P *point }{&point{1, 2}},
			file: `P: { 'x': 1, 'y': 2 }`,
			args: []string{"-P.x", "1", "-P.y", "2"},
			env:  []string{"TEST_P_X=1", "TEST_P_Y=2"},
		},

		{
			val:  struct{ P *point }{&point{1, 2}},
			file: `P: { 'x': 1, 'y': 2 }`,
			args: []string{"-P", "{ 'x': 1, 'y': 2 }"},
			env:  []string{"TEST_P={ 'x': 1, 'y': 2 }"},
		},

		{
			val:  struct{ D time.Duration }{10 * time.Second},
			file: `D: 10s`,
			args: []string{"-D=10s"},
			env:  []string{"TEST_D=10s"},
		},

		{
			val:  struct{ T time.Time }{testTime},
			file: `T: 2016-12-06T01:01:42.123456789Z`,
			args: []string{"-T=2016-12-06T01:01:42.123456789Z"},
			env:  []string{"TEST_T=2016-12-06T01:01:42.123456789Z"},
		},

		{
			val:  struct{ T *time.Time }{&testTime},
			file: `T: 2016-12-06T01:01:42.123456789Z`,
			args: []string{"-T=2016-12-06T01:01:42.123456789Z"},
			env:  []string{"TEST_T=2016-12-06T01:01:42.123456789Z"},
		},

		{
			val:  struct{ M map[string]int }{map[string]int{"answer": 42}},
			file: `M: { answer: 42 }`,
			args: []string{"-M={ answer: 42 }"},
			env:  []string{"TEST_M={ answer: 42 }"},
		},

		{
			val:  struct{ A net.TCPAddr }{net.TCPAddr{IP: net.ParseIP("::1"), Port: 80, Zone: "11"}},
			file: `A: '[::1%11]:80'`,
			args: []string{"-A", "[::1%11]:80"},
			env:  []string{"TEST_A=[::1%11]:80"},
		},

		{
			val:  struct{ A net.UDPAddr }{net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53, Zone: ""}},
			file: `A: 127.0.0.1:53`,
			args: []string{"-A", "127.0.0.1:53"},
			env:  []string{"TEST_A=127.0.0.1:53"},
		},

		{
			val:  struct{ U url.URL }{parseURL("http://localhost:8080/hello/world?answer=42#OK")},
			file: `U: http://localhost:8080/hello/world?answer=42#OK`,
			args: []string{"-U", "http://localhost:8080/hello/world?answer=42#OK"},
			env:  []string{"TEST_U=http://localhost:8080/hello/world?answer=42#OK"},
		},

		{
			val:  struct{ E mail.Address }{parseEmail("Bob <bob@domain.com>")},
			file: `E: Bob <bob@domain.com>`,
			args: []string{"-E", "Bob <bob@domain.com>"},
			env:  []string{"TEST_E=Bob <bob@domain.com>"},
		},
	}
)

func TestLoad(t *testing.T) {
	for _, test := range loadTests {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			ld := Loader{
				Name: "test",
				Args: test.args,
				Sources: []Source{
					SourceFunc(func(dst Map) (err error) { return yaml.Unmarshal([]byte(test.file), dst) }),
					NewEnvSource("test", test.env...),
				},
			}

			val := reflect.New(reflect.TypeOf(test.val))

			if _, _, err := ld.Load(val.Interface()); err != nil {
				t.Error(err)
				t.Log("<<<", test.val)
				t.Log(">>>", val.Elem().Interface())
				return
			}

			if v := val.Elem().Interface(); !reflect.DeepEqual(test.val, v) {
				t.Errorf("bad value:\n<<< %#v\n>>> %#v", test.val, v)
			}
		})
	}
}

func TestDefaultLoader(t *testing.T) {
	const configFile = "/tmp/conf-test.yml"
	ioutil.WriteFile(configFile, []byte(`---
points:
 - { 'x': 0, 'y': 0 }
 - { 'x': 1, 'y': 2 }
 - { 'x': {{ .X }}, 'y': {{ .Y }} }
`), 0644)
	defer os.Remove(configFile)

	tests := []struct {
		args []string
		env  []string
	}{
		{
			args: []string{"test", "-points", `[{'x':0,'y':0},{'x':1,'y':2},{'x':21,'y':42}]`, "A", "B", "C"},
			env:  []string{},
		},
		{
			args: []string{"test", "A", "B", "C"},
			env:  []string{"TEST_POINTS=[{'x':0,'y':0},{'x':1,'y':2},{'x':21,'y':42}]"},
		},
		{
			args: []string{"test", "-config-file", configFile, "A", "B", "C"},
			env:  []string{"X=21", "Y=42"},
		},
		{
			args: []string{"test", "-config-file", configFile, "-points", `[{'x':0,'y':0},{'x':1,'y':2},{'x':21,'y':42}]`, "A", "B", "C"},
			env:  []string{"TEST_POINTS=[{'x':0,'y':0},{'x':1,'y':2},{'x':21,'y':42}]", "X=3", "Y=4"},
		},
	}

	type point struct {
		X int `conf:"x"`
		Y int `conf:"y"`
	}

	type extra struct {
		Dummy []map[string]string
	}

	type config struct {
		// should not impact loading configuration
		unexported bool
		Ignored    string `conf:"-"`

		// these fields only are getting configured
		Points []point `conf:"points"`
		Extra  *extra
		Time   time.Time
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			var cfg config
			_, args, err := defaultLoader(test.args, test.env).Load(&cfg)

			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(args, []string{"A", "B", "C"}) {
				t.Error("bad args:", args)
			}

			if !reflect.DeepEqual(cfg, config{Points: []point{{0, 0}, {1, 2}, {21, 42}}, Extra: &extra{}}) {
				t.Errorf("bad config: %#v", cfg)
			}
		})
	}
}

func TestTemplateFunc(t *testing.T) {
	const configFile = "/tmp/conf-json-test.yml"
	ioutil.WriteFile(configFile, []byte(`---
hello: {{ .NAME | json }}
`), 0644)
	defer os.Remove(configFile)

	var cfg struct {
		Hello string `conf:"hello"`
	}

	_, _, err := defaultLoader([]string{"test", "-config-file", configFile}, []string{
		"NAME=first: Luke, second: Leia",
	}).Load(&cfg)

	if err != nil {
		t.Error(err)
	}

	if cfg.Hello != "first: Luke, second: Leia" {
		t.Error("bad value:", cfg.Hello)
	}
}

func TestCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ld := Loader{
			Name:     "test",
			Args:     []string{"run", "A", "B", "C"},
			Commands: []Command{{"run", ""}, {"version", ""}},
		}

		config := struct{}{}
		cmd, args, err := ld.Load(&config)

		if err != nil {
			t.Error(err)
		}
		if cmd != "run" {
			t.Error("bad command:", cmd)
		}
		if !reflect.DeepEqual(args, []string{"A", "B", "C"}) {
			t.Error("bad arguments:", args)
		}
	})

	t.Run("Missing Command", func(t *testing.T) {
		ld := Loader{
			Name:     "test",
			Args:     []string{},
			Commands: []Command{{"run", ""}, {"version", ""}},
		}

		config := struct{}{}
		_, _, err := ld.Load(&config)

		if err == nil || err.Error() != "missing command" {
			t.Error("bad error:", err)
		}
	})

	t.Run("Unknown Command", func(t *testing.T) {
		ld := Loader{
			Name:     "test",
			Args:     []string{"test"},
			Commands: []Command{{"run", ""}, {"version", ""}},
		}

		config := struct{}{}
		_, _, err := ld.Load(&config)

		if err == nil || err.Error() != "unknown command: test" {
			t.Error("bad error:", err)
		}
	})
}

func TestValidator(t *testing.T) {
	config := struct {
		A struct {
			Bind string `conf:"bind" validate:"nonzero"`
		}
	}{}

	_, _, err := (Loader{}).Load(&config)

	if err == nil {
		t.Error("bad error:", err)
	} else {
		t.Log(err)
	}
}

func TestModifiers(t *testing.T) {
	config := struct {
		Email string `conf:"email" validate:"nonzero" mod:"trim,lcase"`
	}{
		Email: " Test.Email@email.com",
	}

	_, _, err := (Loader{}).Load(&config)
	if err != nil {
		t.Error("bad error:", err)
	}
	if config.Email != "test.email@email.com" {
		t.Error("bad mod value:", config.Email)
	}
}

func parseURL(s string) url.URL {
	u, _ := url.Parse(s)
	return *u
}

func parseEmail(s string) mail.Address {
	a, _ := mail.ParseAddress(s)
	return *a
}

func TestMakeEnvVars(t *testing.T) {
	envList := []string{
		"A=123",
		"B=456",
		"C=789",
		"Hello=World",
		"Answer=42",
		"Key=",
		"Key=Value",
		"Other",
	}

	envVars := makeEnvVars(envList)

	if !reflect.DeepEqual(envVars, map[string]string{
		"A":      "123",
		"B":      "456",
		"C":      "789",
		"Hello":  "World",
		"Answer": "42",
		"Key":    "Value",
		"Other":  "",
	}) {
		t.Error(envVars)
	}
}
