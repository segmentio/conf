package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// Load the program's configuration into dst.
//
// The dst argument is expected to be a pointer to a struct type where exported
// fields or fields with a "conf" tag will be used to load the program
// configuration.
// The function panics if dst is not a pointer to struct, or if it's a nil
// pointer.
//
// The configuration is loaded from the command line, environment and optional
// configuration file if the -config-file option is present in the program
// arguments.
//
// Values found in the progrma arguments take precendence over those found in
// the environment, which takes precendence over the configuration file.
//
// If an error is detected with the configurable the function print the usage
// message to stdout and exit with status code 1.
func Load(dst interface{}) {
	if err := (Loader{}).Load(dst); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// A Loader can be used to provide a costomized configurable for loading a
// configuration.
//
// The zero-value is a valid loader that uses os.Args[1:], os.Environ(), the
// program name, and the "-config-file" argument to load the configuration.
type Loader struct {
	Args     []string // the list of arguments
	Env      []string // the program's environment variables
	Program  string   // the name of the program
	FileFlag string   // command line option for the configuration file
}

// Load uses the loader ld to load the program configuration into dst.
//
// The dst argument is expected to be a pointer to a struct type where exported
// fields or fields with a "conf" tag will be used to load the program
// configuration.
// The function panics if dst is not a pointer to struct, or if it's a nil
// pointer.
func (ld Loader) Load(dst interface{}) error {
	v := reflect.ValueOf(dst)

	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("cannot load configuration into %T", dst))
	}

	if v.IsNil() {
		panic(fmt.Sprintf("cannot load configuration into nil %T", dst))
	}

	if v = v.Elem(); v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cannot load configuration into %T", dst))
	}

	if ld.Args == nil {
		ld.Args = os.Args[1:]
	}

	if ld.Env == nil {
		ld.Env = os.Environ()
	}

	if len(ld.Program) == 0 {
		ld.Program = filepath.Base(os.Args[0])
	}

	if len(ld.FileFlag) == 0 {
		ld.FileFlag = "config-file"
	}

	return ld.load(v)
}

func (ld Loader) load(dst reflect.Value) (err error) {

	return
}
