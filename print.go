package conf

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

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

	set := newFlagSet(makeValue(v), ld.Name, ld.Sources...)

	fmt.Fprintf(w, "%s\n", col.titles("Usage:"))
	fmt.Fprintf(w, "  %s [-h] [-help] [options...]\n\n", ld.Name)

	fmt.Fprintf(w, "%s\n", col.titles("Options:"))

	// Outputs the flags following the same format than the standard flag
	// package. The main difference is in the type names which are set to
	// values returned by prettyType.
	set.VisitAll(func(f *flag.Flag) {
		var t string
		var h []string
		var empty bool
		var boolean bool

		switch v := f.Value.(type) {
		case flagValue:
			t = prettyValueType(v.v)
			empty = isEmptyValue(v.v)
			boolean = v.IsBoolFlag()
		case FlagSource:
			t = "source"
		default:
			t = "value"
		}

		fmt.Fprintf(w, "  %s", col.keys("-"+f.Name))

		switch {
		case !boolean:
			fmt.Fprintf(w, " %s\n", col.types(t))
		case len(f.Name) >= 4: // put help message inline for boolean flags
			fmt.Fprint(w, "\n")
		}

		if s := f.Usage; len(s) != 0 {
			h = append(h, s)
		}

		if s := f.DefValue; len(s) != 0 && !empty && !boolean {
			h = append(h, col.defvals("(default "+s+")"))
		}

		if len(h) != 0 {
			if !boolean || len(f.Name) >= 4 {
				fmt.Fprint(w, "    ")
			}
			fmt.Fprintf(w, "\t%s\n", strings.Join(h, " "))
		}

		fmt.Fprint(w, "\n")
	})
}

func prettyValueType(v reflect.Value) string {
	if x, ok := v.Interface().(specialValue); ok {
		return prettyValueType(x.v)
	}
	return prettyType(v.Type())
}

func prettyType(t reflect.Type) string {
	if t == nil {
		return "unknown"
	}

	switch {
	case t.Implements(objconvValueDecoderInterface):
		return "value"
	case t.Implements(textUnmarshalerInterface):
		return "string"
	}

	switch t {
	case timeDurationType:
		return "duration"
	case timeTimeType:
		return "time"
	case netTCPAddrType, netUDPAddrType:
		return "address"
	case urlURLType:
		return "url"
	case mailAddressType:
		return "email"
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
	}
	return monochrome()
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
