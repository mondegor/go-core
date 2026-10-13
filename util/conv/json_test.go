package conv_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/conv"
)

func TestJSONValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value any
		want  string
	}

	tests := []testCase{
		{name: "string", value: `a "b"`, want: `"a \"b\""`},
		{name: "string like nil", value: "<NIL>", want: `"<NIL>"`},
		{name: "number", value: 12, want: "12"},
		{name: "bool", value: false, want: "false"},
		{name: "nil", value: nil, want: "null"},
		{name: "error", value: errors.New("err"), want: `"err"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, conv.JSONValue(tt.value))
		})
	}
}
