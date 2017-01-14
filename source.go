package conf

import (
	"bytes"
	"flag"
	"html/template"
	"reflect"
	"strings"
)

type Source interface {
	Load(dst interface{}) error
}

type FlagSource interface {
	Source

	Flag() string

	Help() string

	flag.Value
}

type SourceFunc func(dst interface{}) error

func (f SourceFunc) Load(dst interface{}) error {
	return f(dst)
}

func NewEnvSource(prefix string, env ...string) Source {
	return SourceFunc(func(dst interface{}) (err error) {
		if len(env) != 0 {
			type entry struct {
				key string
				val flagValue
			}
			var entries []entry

			scanFields(reflect.ValueOf(dst).Elem(), prefix, "_", func(key string, help string, val reflect.Value) {
				entries = append(entries, entry{
					key: snakecaseUpper(key) + "=",
					val: makeFlagValue(val),
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
	})
}

func NewFileSource(flag string, vars interface{}, readFile func(string) ([]byte, error), unmarshal func([]byte, interface{}) error) Source {
	return &fileSource{
		flag:      flag,
		vars:      vars,
		readFile:  readFile,
		unmarshal: unmarshal,
	}
}

type fileSource struct {
	flag      string
	path      string
	vars      interface{}
	readFile  func(string) ([]byte, error)
	unmarshal func([]byte, interface{}) error
}

func (f *fileSource) Load(dst interface{}) (err error) {
	var b []byte

	if len(f.path) == 0 {
		return
	}

	if b, err = f.readFile(f.path); err != nil {
		return
	}

	tpl := template.New(f.flag)
	buf := &bytes.Buffer{}
	buf.Grow(len(b))

	if _, err = tpl.Parse(string(b)); err != nil {
		return
	}

	if err = tpl.Execute(buf, f.vars); err != nil {
		return
	}

	err = f.unmarshal(buf.Bytes(), dst)
	return
}

func (f *fileSource) Flag() string {
	return f.flag
}

func (f *fileSource) Help() string {
	return "Location to load the configuration file from."
}

func (f *fileSource) Set(s string) error {
	f.path = s
	return nil
}

func (f *fileSource) String() string {
	return f.path
}
