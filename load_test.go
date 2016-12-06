package conf

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	testTime = time.Date(2016, 12, 6, 1, 1, 42, 123456789, time.UTC)
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
			val: struct{ S string }{"Hello World!"},
			env: []string{"TEST_S=Hello World!"},
		},
		{
			val: struct{ S []byte }{[]byte("Hello World!")},
			env: []string{"TEST_S=SGVsbG8gV29ybGQh"},
		},
		{
			val: struct{ L []int }{[]int{1, 2, 3}},
			env: []string{"TEST_L=[1, 2, 3]"},
		},
		{
			val: struct{ L [3]int }{[3]int{1, 2, 3}},
			env: []string{"TEST_L=[1, 2, 3]"},
		},
		{
			val: struct{ P *point }{&point{1, 2}},
			env: []string{"TEST_P_X=1", "TEST_P_Y=2"},
		},
		{
			val: struct{ P *point }{&point{1, 2}},
			env: []string{"TEST_P={ 'x': 1, 'y': 2 }"},
		},
		{
			val: struct{ D time.Duration }{10 * time.Second},
			env: []string{"TEST_D=10s"},
		},
		{
			val: struct{ T time.Time }{testTime},
			env: []string{"TEST_T=2016-12-06T01:01:42.123456789Z"},
		},
		{
			val: struct{ T *time.Time }{&testTime},
			env: []string{"TEST_T=2016-12-06T01:01:42.123456789Z"},
		},
		{
			val: struct{ M map[string]int }{map[string]int{"answer": 42}},
			env: []string{"TEST_M={ answer: 42 }"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
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
			val:  struct{ S string }{"Hello World!"},
			args: []string{"-S", "Hello World!"},
		},
		{
			val:  struct{ S []byte }{[]byte("Hello World!")},
			args: []string{"-S", "SGVsbG8gV29ybGQh\n"},
		},
		{
			val:  struct{ L []int }{[]int{1, 2, 3}},
			args: []string{"-L", "[1,2,3]"},
		},
		{
			val:  struct{ L [3]int }{[3]int{1, 2, 3}},
			args: []string{"-L", "[1,2,3]"},
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			args: []string{"-P.x", "1", "-P.y", "2"},
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			args: []string{"-P", "{ 'x': 1, 'y': 2 }"},
		},
		{
			val:  struct{ D time.Duration }{10 * time.Second},
			args: []string{"-D=10s"},
		},
		{
			val:  struct{ T time.Time }{testTime},
			args: []string{"-T=2016-12-06T01:01:42.123456789Z"},
		},
		{
			val:  struct{ T *time.Time }{&testTime},
			args: []string{"-T=2016-12-06T01:01:42.123456789Z"},
		},
		{
			val:  struct{ M map[string]int }{map[string]int{"answer": 42}},
			args: []string{"-M={ answer: 42 }"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
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
			val:  struct{ S string }{"Hello World!"},
			file: `S: Hello World!`,
		},
		{
			val:  struct{ S []byte }{[]byte("Hello World!")},
			file: `S: SGVsbG8gV29ybGQh`,
		},
		{
			val:  struct{ L []int }{[]int{1, 2, 3}},
			file: `L: [1, 2, 3]`,
		},
		{
			val:  struct{ L [3]int }{[3]int{1, 2, 3}},
			file: `L: [1, 2, 3]`,
		},
		{
			val:  struct{ P *point }{&point{1, 2}},
			file: `P: { 'x': 1, 'y': 2 }`,
		},
		{
			val:  struct{ D time.Duration }{10 * time.Second},
			file: `D: 10s`,
		},
		{
			val:  struct{ T time.Time }{testTime},
			file: `T: 2016-12-06T01:01:42.123456789Z`,
		},
		{
			val:  struct{ T *time.Time }{&testTime},
			file: `T: 2016-12-06T01:01:42.123456789Z`,
		},
		{
			val:  struct{ M map[string]int }{map[string]int{"answer": 42}},
			file: `M: { answer: 42 }`,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
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
