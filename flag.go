package conf

import (
	"flag"
	"io/ioutil"
	"reflect"

	"github.com/segmentio/objconv"
)

func newFlagSet(cfg reflect.Value, name string, sources ...Source) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(ioutil.Discard)

	scanFields(cfg, "", ".", func(key string, help string, val reflect.Value) {
		set.Var(makeFlagValue(val), key, help)
	})

	for _, source := range sources {
		if f, ok := source.(FlagSource); ok {
			set.Var(f, f.Flag(), f.Help())
		}
	}

	return set
}

func scanFields(v reflect.Value, base string, sep string, do func(string, string, reflect.Value)) {
	t := v.Type()

	for i, n := 0, v.NumField(); i != n; i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		name := ft.Name
		help := ft.Tag.Get("help")
		tag, _, _ := objconv.ParseTag(ft.Tag.Get("objconv"))

		if tag == "-" {
			continue
		}

		if len(tag) != 0 {
			name = tag
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
		if fv.Kind() == reflect.Struct && !specialType(ft.Type) {
			scanFields(fv, name, sep, do)
		}
	}
}
