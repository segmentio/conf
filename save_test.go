package conf

import (
	"bytes"
	"testing"
	"time"

	"github.com/segmentio/objconv/yaml"
)

type CfgSave struct {
	String         string                       `conf:"string"                 help:"A string"`
	Int            int                          `conf:"int"                    help:"An int"`
	OmitString     string                       `conf:"omit-string,omitempty"  help:"Omit a string"`
	OmitInt        int                          `conf:"omit-int,omitempty"     help:"Omit an int"`
	Ignored        int                          `conf:"-"                      help:"Ignored field"`
	Date           time.Time                    `conf:"date"                   help:"A date"`
	SubStruct      CfgSub                       `conf:"sub-struct"             help:"A sub struct"`
	Map            map[string]string            `conf:"map"                    help:"A map[string]string"`
	MapStruct      map[int]CfgSub               `conf:"map-struct"             help:"A map[int]CfgSub"`
	MapMap         map[string]map[string]string `conf:"map-map"                help:"A map[string]map[string]string"`
	MapSlice       map[string][]int             `conf:"map-slice"              help:"A map[string][]int"`
	SliceString    []string                     `conf:"slice-string"           help:"A slice of string"`
	SliceStruct    []CfgSub                     `conf:"slice-struct"           help:"A slice of CfgSub"`
	SliceMap       []map[string]string          `conf:"slice-map"              help:"A slice of map"`
	SliceSlice     [][]string                   `conf:"slice-slice"            help:"A slice of slice"`
	StructPtr      *CfgSub                      `conf:"struct-ptr"             help:"A ptr to CfgSub"`
	Bool           bool                         `conf:"bool"                   help:"A boolean"`
	MutilineString string                       `conf:"multi-line-string"      help:"A string with multiple lines"`
	SpecialString  string                       `conf:"special-string"         help:"A string with special char"`
}

type CfgSub struct {
	Name string
	Age  int
}

type BufferSource struct {
	buffer *bytes.Buffer
}

func (src BufferSource) Load(dst interface{}) error {
	dec := yaml.NewDecoder(src.buffer)
	return dec.Decode(dst)
}

func TestSave(t *testing.T) {
	w := &bytes.Buffer{}

	mline := `Hello world!
I'm a gopher and this text is
for multi line test...
    `

	cfg := CfgSave{
		String:  "Hello world",
		Int:     42,
		Ignored: 777,
		Date:    time.Now(),
		SubStruct: CfgSub{
			Name: "Jonhy",
			Age:  42,
		},
		Map: map[string]string{
			"Hello": "world",
			"Foo":   "bar",
		},
		MapStruct: map[int]CfgSub{
			1: CfgSub{
				Name: "Max",
				Age:  30,
			},
			2: CfgSub{
				Name: "Ach",
				Age:  29,
			},
		},
		MapMap: map[string]map[string]string{
			"magic": map[string]string{
				"foo": "bar",
				"ala": "kazam",
			},
			"food": map[string]string{
				"apple": "pie",
				"moji":  "to",
			},
		},
		MapSlice: map[string][]int{
			"geek": []int{42, 4242},
			"luck": []int{777, 13},
		},
		SliceString: []string{
			"Go",
			"C++",
			"C#",
			"Java",
		},
		SliceStruct: []CfgSub{
			CfgSub{
				Name: "Vader",
				Age:  42,
			},
			CfgSub{
				Name: "Sidious",
				Age:  84,
			},
		},
		SliceMap: []map[string]string{
			map[string]string{
				"foo": "bar",
				"ala": "kazam",
			},
			map[string]string{
				"apple": "pie",
				"moji":  "to",
			},
		},
		SliceSlice: [][]string{
			[]string{"tonkatsu", "yakiniku"},
			[]string{"panang", "masaman"},
		},
		StructPtr: &CfgSub{
			Name: "Hamtaro",
			Age:  2,
		},
		MutilineString: mline,
		SpecialString:  "> hello ' ', > world",
	}

	Save(w, cfg)
	t.Log(w.String())

	ld := Loader{
		Sources: []Source{
			BufferSource{
				buffer: w,
			},
		},
	}

	var newCfg CfgSave
	_, _, err := ld.Load(&newCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", newCfg)
}
