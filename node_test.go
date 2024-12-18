package conf

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/objconv/json"
)

func TestEqualNode(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		node1 Node
		node2 Node
		equal bool
	}{
		{
			name:  "nil nodes",
			node1: nil,
			node2: nil,
			equal: true,
		},

		{
			name:  "scalar and nil",
			node1: Scalar{},
			node2: nil,
			equal: false,
		},

		{
			name:  "two empty scalars",
			node1: Scalar{},
			node2: Scalar{},
			equal: true,
		},

		{
			name:  "42 and empty scalar",
			node1: Scalar{reflect.ValueOf(42)},
			node2: Scalar{},
			equal: false,
		},

		{
			name:  "empty scalar and 42",
			node1: Scalar{},
			node2: Scalar{reflect.ValueOf(42)},
			equal: false,
		},

		{
			name:  "42 and 42",
			node1: Scalar{reflect.ValueOf(42)},
			node2: Scalar{reflect.ValueOf(42)},
			equal: true,
		},

		{
			name:  "42 and empty array",
			node1: Scalar{reflect.ValueOf(42)},
			node2: Array{},
			equal: false,
		},

		{
			name:  "non-equal scalars (type mismatch)",
			node1: Scalar{reflect.ValueOf(42)},
			node2: Scalar{reflect.ValueOf("Hello World!")},
			equal: false,
		},

		{
			name:  "equal scalars (time values)",
			node1: Scalar{reflect.ValueOf(now)},
			node2: Scalar{reflect.ValueOf(now.In(time.UTC))},
			equal: true,
		},

		{
			name:  "two empty arrays",
			node1: Array{},
			node2: Array{},
			equal: true,
		},

		{
			name: "equal non-empty arrays",
			node1: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
			node2: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
			equal: true,
		},

		{
			name: "non-equal arrays (value mismatch)",
			node1: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
			node2: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(1)},
			)},
			equal: false,
		},

		{
			name: "non-equal arrays (length mismatch)",
			node1: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
			node2: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
			)},
			equal: false,
		},

		{
			name:  "two empty maps",
			node1: Map{},
			node2: Map{},
			equal: true,
		},

		{
			name: "equal non-empty maps",
			node1: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
			node2: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
			equal: true,
		},

		{
			name: "non-equal maps (value mismatch)",
			node1: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
			node2: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(1)}},
			)},
			equal: false,
		},

		{
			name: "non-equal maps (value not found)",
			node1: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
			node2: Map{items: newMapItems(
				MapItem{Name: "D", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "E", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "F", Value: Scalar{reflect.ValueOf(3)}},
			)},
			equal: false,
		},

		{
			name: "non-equal maps (length mismatch)",
			node1: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
			node2: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
			)},
			equal: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if equal := EqualNode(test.node1, test.node2); equal != test.equal {
				t.Errorf("EqualNode: expected %t but found %t", test.equal, equal)
			}
		})
	}
}

func TestMakeNode(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		value interface{}
		node  Node
	}{
		{
			name:  "nil",
			value: nil,
			node:  Scalar{reflect.ValueOf(nil)},
		},

		{
			name:  "scalar (integer)",
			value: 42,
			node:  Scalar{reflect.ValueOf(42)},
		},

		{
			name:  "scalar (time)",
			value: now,
			node:  Scalar{reflect.ValueOf(now)},
		},

		{
			name:  "slice",
			value: []int{1, 2, 3},
			node: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
		},

		{
			name:  "map",
			value: map[string]int{"A": 1, "B": 2, "C": 3},
			node: Map{items: newMapItems(
				MapItem{Name: "A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
			)},
		},

		{
			name: "struct",
			value: struct {
				A int  `conf:"a" help:"value of A"` // override name
				B int  `conf:"-" help:"value of B"` // skip
				C int  // default name
				D *int // allocate the pointer
				e int  // unexported
			}{1, 2, 3, nil, 42},
			node: Map{items: newMapItems(
				MapItem{Name: "a", Help: "value of A", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "C", Value: Scalar{reflect.ValueOf(3)}},
				MapItem{Name: "D", Value: Scalar{reflect.ValueOf(0)}},
			)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if node := MakeNode(test.value); !EqualNode(node, test.node) {
				t.Errorf("\n<<< %s\n>>> %s", test.node, node)
			}
		})
	}
}

func TestNodeValue(t *testing.T) {
	tests := []struct {
		node  Node
		value interface{}
	}{
		{
			node:  Scalar{},
			value: nil,
		},
		{
			node:  Scalar{reflect.ValueOf(42)},
			value: 42,
		},
		{
			node:  Array{},
			value: nil,
		},
		{
			node:  Array{value: reflect.ValueOf([]int{1, 2, 3})},
			value: []int{1, 2, 3},
		},
		{
			node:  Map{},
			value: nil,
		},
		{
			node:  Map{value: reflect.ValueOf(map[string]int{"A": 1, "B": 2, "C": 3})},
			value: map[string]int{"A": 1, "B": 2, "C": 3},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.value), func(t *testing.T) {
			if value := test.node.Value(); !reflect.DeepEqual(value, test.value) {
				t.Error(value)
			}
		})
	}
}

func TestNodeString(t *testing.T) {
	date := time.Date(2016, 12, 31, 23, 42, 59, 0, time.UTC)

	tests := []struct {
		repr string
		node Node
	}{
		{
			repr: `null`,
			node: Scalar{},
		},
		{
			repr: `42`,
			node: Scalar{reflect.ValueOf(42)},
		},
		{
			repr: `Hello World!`,
			node: Scalar{reflect.ValueOf("Hello World!")},
		},
		{
			repr: `"2016-12-31T23:42:59Z"`,
			node: Scalar{reflect.ValueOf(date)},
		},
		{
			repr: `[ ]`,
			node: Array{},
		},
		{
			repr: `[1, 2, 3]`,
			node: Array{items: newArrayItems(
				Scalar{reflect.ValueOf(1)},
				Scalar{reflect.ValueOf(2)},
				Scalar{reflect.ValueOf(3)},
			)},
		},
		{
			repr: `{ }`,
			node: Map{},
		},
		{
			repr: `{ A: 1 (first), B: 2, C: 3 (last) }`,
			node: Map{items: newMapItems(
				MapItem{Name: "A", Help: "first", Value: Scalar{reflect.ValueOf(1)}},
				MapItem{Name: "B", Value: Scalar{reflect.ValueOf(2)}},
				MapItem{Name: "C", Help: "last", Value: Scalar{reflect.ValueOf(3)}},
			)},
		},
	}

	for _, test := range tests {
		t.Run(test.repr, func(t *testing.T) {
			if repr := test.node.String(); repr != test.repr {
				t.Error("representation mismatch")
				t.Logf("want: %v, %#v", reflect.TypeOf(test.repr), test.repr)
				t.Logf("got: %v, %#v", reflect.TypeOf(repr), repr)
			}
		})
	}
}

func TestNodeJSON(t *testing.T) {
	tests := []struct {
		node Node
		json string
	}{
		{
			node: MakeNode(nil),
			json: `null`,
		},
		{
			node: MakeNode(42),
			json: `42`,
		},
		{
			node: MakeNode("Hello World!"),
			json: `"Hello World!"`,
		},
		{
			node: MakeNode([]int{}),
			json: `[]`,
		},
		{
			node: MakeNode([]int{1, 2, 3}),
			json: `[1,2,3]`,
		},
		{
			node: MakeNode(map[string]int{}),
			json: `{}`,
		},
		{
			node: MakeNode(map[string]int{"A": 1, "B": 2, "C": 3}),
			json: `{"A":1,"B":2,"C":3}`,
		},
		{
			node: MakeNode(struct{ A, B, C int }{1, 2, 3}),
			json: `{"A":1,"B":2,"C":3}`,
		},
		{
			node: MakeNode(struct{ A []int }{[]int{1, 2, 3}}),
			json: `{"A":[1,2,3]}`,
		},
	}

	t.Run("Encode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.json, func(t *testing.T) {
				b, err := json.Marshal(test.node)

				if err != nil {
					t.Error(err)
				}

				if s := string(b); s != test.json {
					t.Error(s)
				}
			})
		}
	})

	t.Run("Decode", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.json, func(t *testing.T) {
				if test.node.Value() == nil {
					return // skip
				}

				value := reflect.New(reflect.TypeOf(test.node.Value()))
				node := MakeNode(value.Interface())

				if err := json.Unmarshal([]byte(test.json), &node); err != nil {
					t.Error(err)
				}
				if !EqualNode(node, test.node) {
					t.Errorf("%+v", node)
				}
			})
		}
	})
}

func Test_FlattenedEmbeddedStructs(t *testing.T) {

	type Smallest struct {
		SmallestOne string
	}

	type Small struct {
		Smallest `conf:"_"`
		SmallOne string
	}

	type Medium struct {
		Small     `conf:"_"`
		MediumOne string
	}

	type Matroska struct {
		Medium   `conf:"_"`
		LargeOne string
	}

	m := Matroska{}
	node := makeNodeStruct(reflect.ValueOf(m), reflect.TypeOf(m))
	if len(node.Items()) != 4 {
		t.Errorf("expected to find four flattened fields...got %d", len(node.Items()))
	}

	for _, name := range []string{"SmallestOne", "SmallOne", "MediumOne", "LargeOne"} {
		f := node.Item(name)
		if f == nil {
			t.Errorf("flattened field %s is missing", name)
		}
		if f.Kind() != ScalarNode {
			t.Errorf("flattened field %s should have been scalar but was %d", name, f.Kind())
		}
	}
}

func Test_InvalidFlattenedEmbeddedStructs(t *testing.T) {

	type Thing1 struct {
		Stuff string
	}

	type Thing2 struct {
		Stuff string
	}

	type ConflictingName struct {
		Thing1 `conf:"_"`
		Thing2 `conf:"_"`
	}

	type EmbedPrimitive struct {
		Str string `conf:"_"`
	}

	type EmbedNamedStruct struct {
		Thing Thing1 `conf:"_"`
	}

	tests := []struct {
		val          interface{}
		errFragments []string
	}{
		{
			val:          ConflictingName{},
			errFragments: []string{"'Stuff'", "duplicate"},
		},
		{
			val:          EmbedPrimitive{},
			errFragments: []string{"\"_\"", "at path EmbedPrimitive.Str"},
		},
		{
			val:          EmbedNamedStruct{},
			errFragments: []string{"\"_\"", "at path EmbedNamedStruct.Thing"},
		},
	}

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.val).Name(), func(t *testing.T) {
			defer func() {
				recovered := recover()
				msg, ok := recovered.(string)
				if !ok {
					t.Errorf("expected a string to be recovered...got %v", recovered)
				}

				// NOTE : ensure that the type name is included in the message!
				for _, frag := range append(tt.errFragments, reflect.TypeOf(tt.val).Name()) {
					if !strings.Contains(msg, frag) {
						t.Errorf("message should have contained fragment \"%s\": %s", frag, msg)
					}
				}
			}()

			makeNodeStruct(reflect.ValueOf(tt.val), reflect.TypeOf(tt.val))
			t.Error("test should have panicked")
		})
	}
}
