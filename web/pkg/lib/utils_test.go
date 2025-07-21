package lib_test

import (
	"testing"

	"go.trulyao.dev/hubble/web/pkg/lib"
)

func Test_Chunk(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o"}
	tests := []struct {
		name     string
		size     int
		expected [][]string
	}{
		{
			name: "chunk size 1",
			size: 1,
			expected: [][]string{
				{
					"a",
				}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}, {"g"}, {"h"}, {"i"}, {"j"}, {"k"}, {"l"}, {"m"}, {"n"}, {"o"},
			},
		},
		{
			name: "chunk size 4",
			size: 4,
			expected: [][]string{
				{
					"a", "b", "c", "d",
				}, {"e", "f", "g", "h"}, {"i", "j", "k", "l"}, {"m", "n", "o"},
			},
		},
		{
			name: "chunk size 7",
			size: 7,
			expected: [][]string{
				{
					"a", "b", "c", "d", "e", "f", "g",
				}, {"h", "i", "j", "k", "l", "m", "n"}, {"o"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := lib.Chunk(input, tt.size)
			if len(actual) != len(tt.expected) {
				t.Fatalf("expected %d chunks, got %d", len(tt.expected), len(actual))
			}

			for i := range actual {
				if len(actual[i]) != len(tt.expected[i]) {
					t.Fatalf(
						"expected chunk %d to have %d elements, got %d",
						i,
						len(tt.expected[i]),
						len(actual[i]),
					)
				}

				for j := range actual[i] {
					if actual[i][j] != tt.expected[i][j] {
						t.Fatalf(
							"expected chunk %d to have element %d to be %s, got %s",
							i,
							j,
							tt.expected[i][j],
							actual[i][j],
						)
					}
				}
			}
		})
	}
}

func Test_Slugify(t *testing.T) {
	type test struct {
		input    string
		expected string
	}

	tests := []test{
		{"Hello, World!", "hello-world"},
		{"Hello, World!123", "hello-world-123"},
		{"Hello, World! 123", "hello-world-123"},
		{"Hello, World! 123 456", "hello-world-123-456"},
		{" The quickbrown   fox", "the-quickbrown-fox"},
		{"tHe quick brown fox", "the-quick-brown-fox"},
		{"tHe quick? brown fox£", "the-quick-brown-fox"},
		{"tHe quick? brown fox£", "the-quick-brown-fox"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := lib.Slugify(tt.input)
			if actual != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func Test_Diff(t *testing.T) {
	type test struct {
		description string
		input       []string
		compare     []string
		output      []string
	}

	tests := []test{
		{
			description: "empty input and empty compare",
			input:       []string{"a", "b", "c"},
			compare:     []string{"a", "b", "c"},
			output:      []string{},
		},
		{
			description: "empty input and non-empty compare",
			input:       []string{"a", "b", "c"},
			compare:     []string{"a", "b", "d"},
			output:      []string{"c", "d"},
		},
		{
			description: "non-empty input and empty compare (flipped)",
			input:       []string{"a", "b", "d"},
			compare:     []string{"a", "b", "c"},
			output:      []string{"d", "c"},
		},
		{
			description: "non-empty input and non-empty compare",
			input:       []string{"a", "b", "c"},
			compare:     []string{"a", "b", "c", "d"},
			output:      []string{"d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual := lib.Diff(tt.input, tt.compare)
			if len(actual) != len(tt.output) {
				t.Fatalf("expected %d elements, got %d", len(tt.output), len(actual))
			}

			for i := range actual {
				if actual[i] != tt.output[i] {
					t.Fatalf("expected %s, got %s", tt.output[i], actual[i])
				}
			}
		})
	}
}
