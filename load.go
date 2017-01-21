package conf

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	validator "gopkg.in/validator.v2"

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
	_, args = LoadWith(cfg, defaultLoader(os.Args, os.Environ()))
	return
}

// LoadWith behaves like Load but uses ld as a loader to parse the program
// configuration.
//
// The function panics if cfg is not a pointer to struct, or if it's a nil
// pointer and no commands were set.
func LoadWith(cfg interface{}, ld Loader) (cmd string, args []string) {
	var err error
	switch cmd, args, err = ld.Load(cfg); err {
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

// A Command represents a command supported by a configuration loader.
type Command struct {
	Name string // name of the command
	Help string // help message describing what the command does
}

// A Loader exposes an API for customizing how a configuration is loaded and
// where it's loaded from.
type Loader struct {
	Name     string    // program name
	Usage    string    // program usage
	Args     []string  // list of arguments
	Commands []Command // list of commands
	Sources  []Source  // list of sources to load configuration from.
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
// pointer and no commands were set.
func (ld Loader) Load(cfg interface{}) (cmd string, args []string, err error) {
	var v1 reflect.Value

	if cfg == nil {
		v1 = reflect.ValueOf(&struct{}{})
	} else {
		v1 = reflect.ValueOf(cfg)
	}

	if v1.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	if v1.IsNil() {
		panic(fmt.Sprintf("cannot load configuration into nil %T", cfg))
	}

	if v1 = v1.Elem(); v1.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	if len(ld.Commands) != 0 {
		if len(ld.Args) == 0 {
			err = errors.New("missing command")
			return
		}

		found := false
		for _, c := range ld.Commands {
			if c.Name == ld.Args[0] {
				found, cmd, ld.Args = true, ld.Args[0], ld.Args[1:]
				break
			}
		}

		if !found {
			err = errors.New("unknown command: " + ld.Args[0])
			return
		}

		if cfg == nil {
			args = ld.Args
			return
		}
	}

	v2 := makeValue(v1)

	if args, err = ld.load(v2); err != nil {
		return
	}

	setZero(v1)
	setValue(v1, v2)

	if err = validator.Validate(v1.Interface()); err != nil {
		if errmap, ok := err.(validator.ErrorMap); ok {
			errkeys := make([]string, 0, len(errmap))
			errlist := make(errorList, 0, len(errmap))

			for errkey := range errmap {
				errkeys = append(errkeys, errkey)
			}

			sort.Strings(errkeys)

			for _, errkey := range errkeys {
				path := fieldPath(v1.Type(), errkey)

				if len(errmap[errkey]) == 1 {
					errlist = append(errlist, fmt.Errorf("invalid value passed to %s: %s", path, errmap[errkey][0]))
				} else {
					buf := &bytes.Buffer{}
					fmt.Fprintf(buf, "invalid value passed to %s: ", path)

					for i, errval := range errmap[errkey] {
						if i != 0 {
							buf.WriteString("; ")
						}
						buf.WriteString(errval.Error())
					}

					errlist = append(errlist, errors.New(buf.String()))
				}
			}

			err = errlist
		}
	}

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

type errorList []error

func (err errorList) Error() string {
	if len(err) > 0 {
		return err[0].Error()
	}
	return ""
}
