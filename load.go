package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/segmentio/objconv/yaml"
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
	return LoadWith(cfg, defaultLoader(os.Args, os.Environ()))
}

// LoadWith behaves like Load but uses ld as a loader to parse the program
// configuration.
func LoadWith(cfg interface{}, ld Loader) (args []string) {
	var err error
	switch args, err = ld.Load(cfg); err {
	case nil:
	case flag.ErrHelp:
		ld.PrintHelp(cfg)
		os.Exit(0)
	default:
		ld.PrintHelp(cfg)
		ld.PrintError(err)
		os.Exit(1)
	}
	return
}

// A Loader can be used to provide a costomized configurable for loading a
// configuration.
type Loader struct {
	Name    string   // program name
	Usage   string   // program usage
	Args    []string // list of arguments
	Sources []Source // list of sources to load configuration from.
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

	v2 := makeValue(v1)

	if args, err = ld.load(v2); err != nil {
		return
	}

	setZero(v1)
	setValue(v1, v2)
	return
}

func (ld Loader) load(cfg reflect.Value) (args []string, err error) {
	set := newFlagSet(cfg, ld.Name, ld.Sources...)

	// Parse the arguments a first time so the sources that implement the
	// FlagSource interface get their values loaded.
	if err = set.Parse(ld.Args); err != nil {
		return
	}

	// Load the configuration from the sources that have been configured on the
	// loader.
	// Order is important here because the values will get overwritten by each
	// source that loads the configuration.
	for _, source := range ld.Sources {
		if err = source.Load(cfg.Addr().Interface()); err != nil {
			return
		}
	}

	// Parse the arguments a second time to overwrite values loaded by sources
	// which were also passed to the program arguments.
	if err = set.Parse(ld.Args); err != nil {
		return
	}

	args = set.Args()
	return
}

func defaultLoader(args []string, env []string) Loader {
	var name = filepath.Base(args[0])
	return Loader{
		Name: name,
		Args: args[1:],
		Sources: []Source{
			NewFileSource("config-file", makeEnvVars(env), ioutil.ReadFile, yaml.Unmarshal),
			NewEnvSource(name, env...),
		},
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
