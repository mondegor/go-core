package mrmsg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrmsg"
)

func TestPlaceholderExtractor_Extract(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		leftDelim  string
		rightDelim string
		message    string
		want       []string
	}

	tests := []testCase{
		{name: "empty message", message: "", want: nil},
		{name: "default delims", message: "a {Name} b {Age1}", want: []string{"{Name}", "{Age1}"}},
		{name: "duplicates are skipped", message: "{A} {B} {A}", want: []string{"{A}", "{B}"}},
		{name: "empty placeholder", message: "{} {A}", want: []string{"{A}"}},
		{name: "invalid body", message: "{1a} {a-b} {a b}", want: nil},
		{name: "body inside invalid", message: "{x {Name}", want: []string{"{Name}"}},
		{name: "without right delim", message: "{Name", want: nil},
		{
			name:       "custom delims",
			leftDelim:  "{{",
			rightDelim: "}}",
			message:    "{{Name}} {Age} {{Name}}",
			want:       []string{"{{Name}}"},
		},
		{
			name:       "only left delim custom",
			leftDelim:  "[",
			rightDelim: "",
			message:    "[Name} {Age}",
			want:       []string{"[Name}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mrmsg.NewPlaceholderExtractor(tt.leftDelim, tt.rightDelim).Extract(tt.message)
			assert.Equal(t, tt.want, got)
		})
	}
}
