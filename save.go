package conf

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// SaveTo writes a config struct into the file name in YAML format.
// name is created if not exists.
func SaveTo(name string, cfg interface{}) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	Save(f, cfg)
	return nil
}

// Save writes a config struct into w in YAML format.
func Save(w io.Writer, cfg interface{}) {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprint("cfg should be a struct"))
	}

	saveStruct(w, v, 0)
}

func saveStruct(w io.Writer, v reflect.Value, indent int) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		conf := f.Tag.Get("conf")
		if conf == "-" {
			continue
		}

		confArgs := strings.Split(conf, ",")
		if len(confArgs) >= 2 && confArgs[1] == "omitempty" {
			fi := fv.Interface()
			zi := reflect.Zero(f.Type).Interface()
			if reflect.DeepEqual(fi, zi) {
				continue
			}
		}

		if help := f.Tag.Get("help"); len(help) != 0 {
			fmt.Fprintln(w)
			saveIndent(w, indent)
			fmt.Fprintln(w, "#", help)
		}

		name := f.Name
		if len(confArgs) > 0 && len(confArgs[0]) != 0 {
			name = confArgs[0]
		}

		saveIndent(w, indent)
		fmt.Fprintf(w, "%v: ", name)
		save(w, fv, indent)
	}
}

func save(w io.Writer, v reflect.Value, indent int) {
	if _, ok := v.Type().MethodByName("String"); ok {
		saveString(w, v, indent)
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		fmt.Fprintln(w)
		saveStruct(w, v, indent+1)

	case reflect.Map:
		fmt.Fprintln(w)
		saveMap(w, v, indent+1)

	case reflect.Slice:
		fmt.Fprintln(w)
		saveSlice(w, v, indent+1)

	case reflect.Ptr:
		save(w, v.Elem(), indent)

	default:
		saveString(w, v, indent)
	}
}

func saveMap(w io.Writer, v reflect.Value, indent int) {
	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)
		saveIndent(w, indent)
		fmt.Fprintf(w, "%v: ", mk.Interface())
		save(w, mv, indent)
	}
}

func saveSlice(w io.Writer, v reflect.Value, indent int) {
	for i := 0; i < v.Len(); i++ {
		sv := v.Index(i)
		saveIndent(w, indent)
		fmt.Fprint(w, "- ")
		save(w, sv, indent)
	}
}

func saveString(w io.Writer, v reflect.Value, indent int) {
	str := fmt.Sprint(v.Interface())

	s := strings.Split(str, "\n")
	if len(s) == 1 {
		fmt.Fprintln(w, str)
		return
	}

	fmt.Fprintln(w, "|")
	indent++

	for _, line := range s {
		saveIndent(w, indent)
		fmt.Fprintln(w, line)
	}
}

func saveIndent(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "  ")
	}
}
