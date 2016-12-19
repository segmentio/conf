package conf

import (
	"bytes"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/segmentio/jutil"
)

// Load the program's configuration into cfg, and returns the list of leftover
// arguments.
//
// The cfg argument is expected to be a pointer to a struct type where exported
// fields or fields with a "conf" tag will be used to load the program
// configuration.
// The function panics if cfg is not a pointer to struct, or if it's a nil
// pointer.
//
// The configuration is loaded from the command line, environment and optional
// configuration file if the -config-file option is present in the program
// arguments.
//
// Values found in the progrma arguments take precedence over those found in
// the environment, which takes precedence over the configuration file.
//
// If an error is detected with the configurable the function print the usage
// message to stderr and exit with status code 1.
func Load(cfg interface{}) (args []string) {
	var err error
	var name = filepath.Base(os.Args[0])
	var env = os.Environ()
	var ld = Loader{
		Args:     os.Args[1:],
		Env:      env,
		Vars:     makeEnvVars(env),
		Program:  name,
		FileFlag: "config-file",
	}

	switch args, err = ld.Load(cfg); err {
	case nil:
	case flag.ErrHelp:
		ld.PrintHelp(cfg)
		os.Exit(0)
	default:
		ld.PrintError(err)
		ld.PrintHelp(cfg)
		os.Exit(1)
	}

	return
}

// A Loader can be used to provide a costomized configurable for loading a
// configuration.
type Loader struct {
	Args     []string    // list of arguments
	Env      []string    // list of environment variables ["KEY=VALUE", ...]
	Vars     interface{} // template variables, may be a struct, map, etc..
	Program  string      // name of the program
	FileFlag string      // command line option for the configuration file
}

// Load uses the loader ld to load the program configuration into cfg, and
// returns the list of program arguments that were not used.
//
// The function returns flag.ErrHelp when the list of arguments contained -h,
// -help, or --help.
//
// The cfg argument is expected to be a pointer to a struct type where exported
// fields or fields with a "conf" tag will be used to load the program
// configuration.
// The function panics if cfg is not a pointer to struct, or if it's a nil
// pointer.
func (ld Loader) Load(cfg interface{}) (args []string, err error) {
	v1 := reflect.ValueOf(cfg)

	if v1.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	if v1.IsNil() {
		panic(fmt.Sprintf("cannot load configuration into nil %T", cfg))
	}

	if v1 = v1.Elem(); v1.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	v2 := makeConfValue(v1)

	if args, err = ld.load(v2); err != nil {
		return
	}

	setZero(v1)
	setValue(v1, v2)
	return
}

func (ld Loader) load(cfg reflect.Value) (args []string, err error) {
	if err = loadFile(cfg, ld.Program, ld.FileFlag, ld.Args, ld.Vars, ioutil.ReadFile); err != nil {
		args = nil
		return
	}

	if err = loadEnv(cfg, ld.Program, ld.Env); err != nil {
		args = nil
		return
	}

	return loadArgs(cfg, ld.Program, ld.FileFlag, ld.Args)
}

func loadFile(cfg reflect.Value, name string, fileFlag string, args []string, vars interface{}, readFile func(string) ([]byte, error)) (err error) {
	if len(fileFlag) != 0 {
		var a = append([]string{}, args...)
		var b []byte
		var f string
		var v = reflect.New(cfg.Type()).Elem() // discard values from the arguments

		set := newFlagSet(v, name)
		addFileFlag(set, &f, fileFlag)

		if err = set.Parse(a); err != nil {
			return
		}

		if len(f) == 0 {
			return
		}

		if b, err = readFile(f); err != nil {
			return
		}

		tpl := template.New("config")
		buf := &bytes.Buffer{}
		buf.Grow(65536)

		if _, err = tpl.Parse(string(b)); err != nil {
			return
		}

		if err = tpl.Execute(buf, vars); err != nil {
			return
		}

		if err = yaml.Unmarshal(buf.Bytes(), cfg.Addr().Interface()); err != nil {
			return
		}
	}
	return
}

func loadEnv(cfg reflect.Value, name string, env []string) (err error) {
	if len(env) != 0 {
		type entry struct {
			key string
			val flagValue
		}
		var entries []entry

		scanFields(cfg, name, "_", func(key string, help string, val reflect.Value) {
			entries = append(entries, entry{
				key: snakecaseUpper(key) + "=",
				val: flagValue{val},
			})
		})

		for _, e := range entries {
			for _, kv := range env {
				if strings.HasPrefix(kv, e.key) {
					if err = e.val.Set(kv[len(e.key):]); err != nil {
						return
					}
					break
				}
			}
		}
	}
	return
}

func loadArgs(cfg reflect.Value, name string, fileFlag string, args []string) (leftover []string, err error) {
	if len(args) != 0 {
		args = append([]string{}, args...)
		set := newFlagSet(cfg, name)

		if len(fileFlag) != 0 {
			addFileFlag(set, nil, fileFlag)
		}

		if err = set.Parse(args); err != nil {
			return
		}

		leftover = set.Args()
	}
	return
}

type flagValue struct {
	v reflect.Value
}

func (f flagValue) String() string {
	var b []byte

	if !f.v.IsValid() {
		return ""
	}

	switch v := f.v.Interface().(type) {
	case encoding.TextMarshaler:
		b, _ = v.MarshalText()
	default:
		b, _ = json.Marshal(v)
	}

	return string(b)
}

func (f flagValue) Get() interface{} {
	if f.v.IsValid() {
		return nil
	}
	return f.v.Interface()
}

func (f flagValue) Set(s string) error {
	return yaml.Unmarshal([]byte(s), f.v.Addr().Interface())
}

func (f flagValue) IsBoolFlag() bool {
	return f.v.IsValid() && f.v.Kind() == reflect.Bool
}

func newFlagSet(cfg reflect.Value, name string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(ioutil.Discard)

	scanFields(cfg, "", ".", func(key string, help string, val reflect.Value) {
		set.Var(flagValue{val}, key, help)
	})

	return set
}

func addFileFlag(set *flag.FlagSet, f *string, arg string) {
	if f == nil {
		f = new(string)
	}
	set.Var(flagValue{reflect.ValueOf(f).Elem()}, arg, "Path to the configuration file")
}

func scanFields(v reflect.Value, base string, sep string, do func(string, string, reflect.Value)) {
	t := v.Type()

	for i, n := 0, v.NumField(); i != n; i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if len(ft.PkgPath) != 0 {
			continue // unexported field
		}

		name := ft.Name
		help := ft.Tag.Get("help")
		jtag := jutil.ParseTag(ft.Tag.Get("json"))

		if jtag.Skip {
			continue
		}

		if len(jtag.Name) != 0 {
			name = jtag.Name
		}

		if len(base) != 0 {
			name = base + sep + name
		}

		// Dereference all pointers and create objects on the ones that are nil.
		for fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				fv.Set(reflect.New(ft.Type.Elem()))
			}
			fv = fv.Elem()
		}

		// For all fields the delegate is called.
		do(name, help, fv)

		// Inner structs are flattened to allow composition of configuration
		// objects.
		if fv.Kind() == reflect.Struct {
			switch ft.Type {
			case timeTimeType:
			case netTCPAddrType:
			case netUDPAddrType:
			case confNetAddrType:
			case urlURLType:
			case confURLType:
			case mailAddressType:
			case confEmailType:
			default:
				scanFields(fv, name, sep, do)
			}
		}
	}
}

func makeEnvVars(env []string) (vars map[string]string) {
	vars = make(map[string]string)

	for _, e := range env {
		var k string
		var v string

		if off := strings.IndexByte(e, '='); off >= 0 {
			k, v = e[:off], e[off+1:]
		} else {
			k = e
		}

		vars[k] = v
	}

	return vars
}
