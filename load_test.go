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
)

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
			val:  struct{ S string }{"Hello World!"},
			file: `S: Hello World!`,
			args: []string{"-S", "Hello World!"},
			env:  []string{"TEST_S=Hello World!"},
		},

		{
			val:  struct{ S []byte }{[]byte("Hello World!")},
			file: `S: SGVsbG8gV29ybGQh`,
			args: []string{"-S", "SGVsbG8gV29ybGQh\n"},
			env:  []string{"TEST_S=SGVsbG8gV29ybGQh"},
		},

		{
			val:  struct{ L []int }{[]int{1, 2, 3}},
			file: `L: [1, 2, 3]`,
			args: []string{"-L", "[1,2,3]"},
			env:  []string{"TEST_L=[1, 2, 3]"},
		},

		{
			val:  struct{ L [3]int }{[3]int{1, 2, 3}},
			file: `L: [1, 2, 3]`,
			args: []string{"-L", "[1,2,3]"},
			env:  []string{"TEST_L=[1, 2, 3]"},
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
			args: []string{"-A", "'[::1%11]:80'"},
			env:  []string{"TEST_A='[::1%11]:80'"},
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

func TestLoadEnv(t *testing.T) {
	for _, test := range loadTests {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()
			v3 := reflect.New(makeType(v1.Type())).Elem()
			setValue(v3, v2)

			if err := loadEnv(v3, "test", test.env); err != nil {
				t.Error(err)
			}

			setValue(v2, v3)
			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("\n<<< %#v\n>>> %#v", x1, x2)
			}
		})
	}
}

func TestLoadArgs(t *testing.T) {
	for _, test := range loadTests {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()
			v3 := reflect.New(makeType(v1.Type())).Elem()
			setValue(v3, v2)

			if _, err := loadArgs(v3, "test", "", test.args); err != nil {
				t.Error(err)
			}

			setValue(v2, v3)
			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("\n<<< %#v\n>>> %#v", x1, x2)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	for _, test := range loadTests {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			v1 := reflect.ValueOf(test.val)
			v2 := reflect.New(v1.Type()).Elem()
			v3 := reflect.New(makeType(v1.Type())).Elem()
			setValue(v3, v2)

			readFile := func(file string) (b []byte, err error) {
				if file != "test.yml" {
					t.Error(file)
				}
				b = []byte(test.file)
				return
			}

			if err := loadFile(v3, "test", "config-file", []string{"-config-file", "test.yml"}, readFile); err != nil {
				t.Error(err)
			}

			setValue(v2, v3)
			x1 := v1.Interface()
			x2 := v2.Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Errorf("\n<<< %#v\n>>> %#v", x1, x2)
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
		X int `conf:"x"`
		Y int `conf:"y"`
	}

	type config struct {
		Points []point `conf:"points"`
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

func parseURL(s string) url.URL {
	u, _ := url.Parse(s)
	return *u
}

func parseEmail(s string) mail.Address {
	a, _ := mail.ParseAddress(s)
	return *a
}
