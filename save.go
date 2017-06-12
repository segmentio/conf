package conf

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/segmentio/objconv/json"
)

// SaveTo writes a config struct into the file name in YAML format.
// name is created if it doesn't exist.
func SaveTo(name string, cfg interface{}) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return Save(f, cfg)
}

// Save writes a config struct into w in YAML format.
func Save(w io.Writer, cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprint("cfg should be a struct"))
	}

	sw := &saveWriter{Writer: w}
	saveStruct(sw, v, 0)
	return sw.err
}

type saveWriter struct {
	io.Writer
	err error
}

func (w *saveWriter) Write(b []byte) (n int, err error) {
	if err = w.err; err == nil {
		if n, err = w.Writer.Write(b); err != nil {
			w.err = err
		}
	}
	return
}

func saveStruct(w *saveWriter, v reflect.Value, indent int) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		conf := f.Tag.Get("conf")
		if conf == "-" {
			continue
		}

		if help := f.Tag.Get("help"); len(help) != 0 {
			fmt.Fprintln(w)
			saveIndent(w, indent)
			fmt.Fprintln(w, "#", help)
		}

		name := f.Name
		if len(conf) != 0 {
			name = conf
		}

		saveIndent(w, indent)
		fmt.Fprintf(w, "%v: ", name)
		save(w, fv, indent)
	}
}

func save(w *saveWriter, v reflect.Value, indent int) {
	switch v.Kind() {
	case reflect.Struct:
		switch s := v.Interface().(type) {
		case time.Time:
			fmt.Fprintln(w, s.Format(time.RFC3339Nano))
			return
		}

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

	case reflect.Bool:
		fmt.Fprintln(w, v.Interface())

	case reflect.String:
		saveString(w, v, indent)

	default:
		fmt.Fprintln(w, v.Interface())
	}
}

func saveMap(w *saveWriter, v reflect.Value, indent int) {
	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)

		saveIndent(w, indent)
		fmt.Fprintf(w, "%v: ", mk.Interface())
		save(w, mv, indent)
	}
}

func saveSlice(w *saveWriter, v reflect.Value, indent int) {
	for i := 0; i < v.Len(); i++ {
		sv := v.Index(i)

		saveIndent(w, indent)
		fmt.Fprint(w, "- ")
		save(w, sv, indent)
	}
}

func saveString(w *saveWriter, v reflect.Value, indent int) {
	str := fmt.Sprint(v.Interface())

	if len(str) == 0 {
		fmt.Fprintln(w)
	}

	trimed := strings.TrimSpace(str)
	if len(trimed) != 0 {
		marshal := false

		switch trimed[0] {
		case '\'', '"', '`', '>', '|', '?', '!', '&', '@', '%', '*', '-', '[', ']', '{', '}', ':':
			marshal = true
		}

		switch trimed {
		case "true", "True", "TRUE", "false", "False", "FALSE", "null", "Null", "NULL", ".nan", ".NaN", ".NAN", ".inf", ".Inf", ".INF":
			marshal = true
		}

		if marshal {
			d, err := json.Marshal(str)
			if err != nil {
				w.err = err
				return
			}
			fmt.Fprintf(w, "%s\n", d)
			return
		}
	}

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

func saveIndent(w *saveWriter, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "  ")
	}
}