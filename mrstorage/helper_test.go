package mrstorage_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrstorage"
)

func TestNonZeroLimit(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value int
		want  string
	}

	tests := []testCase{
		{
			name:  "negative",
			value: -10,
			want:  " LIMIT 1",
		},
		{
			name:  "zero",
			value: 0,
			want:  " LIMIT 1",
		},
		{
			name:  "one",
			value: 1,
			want:  " LIMIT 1",
		},
		{
			name:  "positive",
			value: 25,
			want:  " LIMIT 25",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mrstorage.NonZeroLimit(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageLimit(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value int
		want  int
	}

	tests := []testCase{
		{name: "negative", value: -1, want: 1},
		{name: "zero", value: 0, want: 1},
		{name: "one", value: 1, want: 1},
		{name: "regular", value: 50, want: 50},
		{name: "max limit", value: math.MaxInt32 - 1, want: math.MaxInt32 - 1},
		{name: "max int32 is capped", value: math.MaxInt32, want: math.MaxInt32 - 1},
		{name: "max int is capped", value: math.MaxInt, want: math.MaxInt32 - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, mrstorage.PageLimit(tt.value))
		})
	}
}
