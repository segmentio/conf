package conf

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// PrintError outputs the error message for err to stderr.
func (ld Loader) PrintError(err error) {
	w := bufio.NewWriter(os.Stderr)
	ld.fprintError(w, err, stderr())
	w.Flush()
}

// FprintError outputs the error message for err to w.
func (ld Loader) FprintError(w io.Writer, err error) {
	ld.fprintError(w, err, monochrome())
}

// PrintHelp outputs the help message for cfg to stderr.
func (ld Loader) PrintHelp(cfg interface{}) {
	w := bufio.NewWriter(os.Stderr)
	ld.fprintHelp(w, cfg, stderr())
	w.Flush()
}

// FprintHelp outputs the help message for cfg to w.
func (ld Loader) FprintHelp(w io.Writer, cfg interface{}) {
	ld.fprintHelp(w, cfg, monochrome())
}

func (ld Loader) fprintError(w io.Writer, err error, col colors) {
	fmt.Fprintf(w, "%s\n  %s\n\n", col.titles("Error:"), col.errors(err.Error()))
}

func (ld Loader) fprintHelp(w io.Writer, cfg interface{}, col colors) {
	v := reflect.ValueOf(cfg)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	set := newFlagSet(makeConfValue(v), ld.Program)

	if len(ld.FileFlag) != 0 {
		addFileFlag(set, nil, ld.FileFlag)
	}

	fmt.Fprintf(w, "%s\n", col.titles(fmt.Sprintf("Usage of %s:", ld.Program)))

	// Outputs the flags following the same format than the standard flag
	// package. The main difference is in the type names which are set to
	// values returned by prettyType.
	set.VisitAll(func(f *flag.Flag) {
		v := f.Value.(value)
		h := []string{}

		fmt.Fprintf(w, "  %s", col.keys("-"+f.Name))

		switch {
		case !v.IsBoolFlag():
			fmt.Fprintf(w, " %s\n", col.types(prettyType(v.v.Type())))
		case len(f.Name) > 4: // put help message inline for boolean flags
			fmt.Fprint(w, "\n")
		}

		if s := f.Usage; len(s) != 0 {
			h = append(h, s)
		}

		if s := f.DefValue; len(s) != 0 && !v.IsBoolFlag() && !isZeroValue(v.v) && v.v.Kind() != reflect.Struct {
			h = append(h, col.defvals("(default "+s+")"))
		}

		if len(h) != 0 {
			if !v.IsBoolFlag() || len(f.Name) > 4 {
				fmt.Fprint(w, "    ")
			}
			fmt.Fprintf(w, "\t%s\n", strings.Join(h, " "))
		}

		fmt.Fprint(w, "\n")
	})
}

func prettyType(t reflect.Type) string {
	if t == nil {
		return "unknown"
	}

	switch {
	case t == reflect.TypeOf(time.Duration(0)):
		return "duration"

	case t == reflect.TypeOf(duration(0)):
		return "duration"

	case t == reflect.TypeOf(time.Time{}):
		return "time"
	}

	switch t.Kind() {
	case reflect.Struct, reflect.Map:
		return "object"

	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return "base64"
		}
		return "list"

	case reflect.Ptr:
		return prettyType(t.Elem())

	default:
		return t.String()
	}
}

type colors struct {
	titles  func(string) string
	keys    func(string) string
	types   func(string) string
	defvals func(string) string
	errors  func(string) string
}

func stderr() colors {
	if terminal.IsTerminal(2) {
		return colorized()
	} else {
		return monochrome()
	}
}

func colorized() colors {
	return colors{
		titles:  bold,
		keys:    blue,
		types:   green,
		defvals: grey,
		errors:  red,
	}
}

func monochrome() colors {
	return colors{
		titles:  normal,
		keys:    normal,
		types:   normal,
		defvals: normal,
		errors:  normal,
	}
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[1;34m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[1;32m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[1;31m" + s + "\033[0m"
}

func grey(s string) string {
	return "\033[1;30m" + s + "\033[0m"
}

func normal(s string) string {
	return s
}
