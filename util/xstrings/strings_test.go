package xstrings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xstrings"
)

func TestTrimBeforeSep(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		s    string
		want string
	}

	tests := []testCase{
		{name: "found", s: "key.value", want: "value"},
		{name: "first sep only", s: "a.b.c", want: "b.c"},
		{name: "sep at end", s: "abc.", want: ""},
		{name: "not found", s: "abc", want: "abc"},
		{name: "empty", s: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, xstrings.TrimBeforeSep(tt.s, '.'))
		})
	}
}

func TestSortedUnique(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"a", "b", "c"}, xstrings.SortedUnique([]string{"c", "a", "b", "a", "c"}))
	assert.Empty(t, xstrings.SortedUnique(nil))
}
