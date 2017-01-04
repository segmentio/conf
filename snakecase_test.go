package conf

import "testing"

var (
	snakecaseTests = []struct {
		in  string
		out string
	}{
		{"", ""},
		{"A", "a"},
		{"HelloWorld", "hello_world"},
		{"HELLOWorld", "hello_world"},
		{"Hello1World2", "hello1_world2"},
		{"123_", "123_"},
		{"_", "_"},
		{"___", "___"},
		{"HELLO_WORLD", "hello_world"},
		{"HelloWORLD", "hello_world"},
		{"test_P_x", "test_p_x"},
		{"__hello_world__", "__hello_world__"},
		{"__Hello_World__", "__hello_world__"},
		{"__Hello__World__", "__hello__world__"},
		{"hello-world", "hello_world"},
	}
)

func TestSnakecaseLower(t *testing.T) {
	for _, test := range snakecaseTests {
		t.Run(test.in, func(t *testing.T) {
			if s := snakecaseLower(test.in); s != test.out {
				t.Error(s)
			}
		})
	}
}

func BenchmarkSnakecase(b *testing.B) {
	for _, test := range snakecaseTests {
		b.Run(test.in, func(b *testing.B) {
			for i := 0; i != b.N; i++ {
				snakecase(test.in)
			}
		})
	}
}
