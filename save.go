package conf

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"regexp"

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

	return saveStruct(w, v, 0)
}

func saveStruct(w io.Writer, v reflect.Value, indent int) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		conf := f.Tag.Get("conf")
		if conf == "-" {
			continue
		}

		if help := f.Tag.Get("help"); len(help) != 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			if err := saveIndent(w, indent); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, "#", help); err != nil {
				return err
			}
		}

		name := f.Name
		if len(conf) != 0 {
			name = conf
		}

		if err := saveIndent(w, indent); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%v: ", name); err != nil {
			return err
		}
		if err := save(w, fv, indent); err != nil {
			return err
		}
	}
	return nil
}

func save(w io.Writer, v reflect.Value, indent int) error {
	switch v.Kind() {
	case reflect.Struct:
		switch s := v.Interface().(type) {
		case time.Time:
			_, err := fmt.Fprintln(w, s.Format(time.RFC3339Nano))
			return err
		}

		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		return saveStruct(w, v, indent+1)

	case reflect.Map:
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		return saveMap(w, v, indent+1)

	case reflect.Slice:
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		return saveSlice(w, v, indent+1)

	case reflect.Ptr:
		return save(w, v.Elem(), indent)

	case reflect.Bool:
		_, err := fmt.Fprintln(w, v.Interface())
		return err

	default:
		return saveString(w, v, indent)
	}
}

func saveMap(w io.Writer, v reflect.Value, indent int) error {
	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)

		if err := saveIndent(w, indent); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%v: ", mk.Interface()); err != nil {
			return err
		}
		if err := save(w, mv, indent); err != nil {
			return err
		}
	}
	return nil
}

func saveSlice(w io.Writer, v reflect.Value, indent int) error {
	for i := 0; i < v.Len(); i++ {
		sv := v.Index(i)

		if err := saveIndent(w, indent); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "- "); err != nil {
			return err
		}
		if err := save(w, sv, indent); err != nil {
			return err
		}
	}
	return nil
}

func saveString(w io.Writer, v reflect.Value, indent int) error {
	str := fmt.Sprint(v.Interface())

	if len(str) == 0 {
		_, err := fmt.Fprintln(w)
		return err
	}

	trimed := strings.TrimSpace(str)
	if len(trimed) != 0 {
		marshal := false

		switch trimed[0] {
		case '\'', '"', '`', '>', '|', '?', '!', '&', '@', '%', '*', '-', '[', ']', '{', '}', ':':
			marshal = true
		}

		switch trimed {
		case "true", "True", "TRUE", "false", "False", "FALSE", "null", "Null", "NULL":
			marshal = true
		}

		rxpNan := regexp.MustCompile(`^\.(nan|NaN|NAN)`)
		if rxpNan.MatchString(trimed) {
			marshal = true
		}

		rxpInf := regexp.MustCompile(`^\.(inf|Inf|INF)`)
		if rxpInf.MatchString(trimed) {
			marshal = true
		}

		if marshal {
			d, err := json.Marshal(str)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(w, "%s\n", d)
			return err
		}
	}

	s := strings.Split(str, "\n")
	if len(s) == 1 {
		_, err := fmt.Fprintln(w, str)
		return err
	}

	if _, err := fmt.Fprintln(w, "|"); err != nil {
		return err
	}
	indent++

	for _, line := range s {
		if err := saveIndent(w, indent); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func saveIndent(w io.Writer, n int) error {
	for i := 0; i < n; i++ {
		if _, err := fmt.Fprint(w, "  "); err != nil {
			return err
		}
	}
	return nil
}
