package conv_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/conv"
)

func TestGroup_String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		group conv.Group
		want  string
	}

	tests := []testCase{
		{
			name:  "data with string value",
			group: map[string]any{"key1": "stringValue"},
			want:  `{"key1": "stringValue"}`,
		},
		{
			name:  "data with int value",
			group: map[string]any{"key1": 1234},
			want:  `{"key1": 1234}`,
		},
		{
			name:  "data with boolean value",
			group: map[string]any{"key1": true},
			want:  `{"key1": true}`,
		},
		{
			name:  "empty data",
			group: nil,
			want:  "{}",
		},
		{
			name:  "data with nil value",
			group: map[string]any{"key1": nil},
			want:  `{"key1": null}`,
		},
		{
			name: "data with structure value",
			group: map[string]any{"key1": struct {
				key2 string
				key3 int
			}{`value and "quotes"`, 1234}},
			want: `{"key1": "{value and \"quotes\" 1234}"}`,
		},
		{
			name:  "data with a few values",
			group: map[string]any{"key1": `stringValue and "quotes"`, "key3": false, "key2": 1234.56, "key4": nil, "key5": 1234.0},
			want:  `{"key1": "stringValue and \"quotes\"", "key2": 1234.56, "key3": false, "key4": null, "key5": 1234}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.group.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_ValueString(t *testing.T) {
	t.Parallel()

	group := conv.Group{"key1": 12, "key2": "str"}

	assert.Equal(t, "12", group.ValueString("key1"))
	assert.Equal(t, "str", group.ValueString("key2"))
	assert.Empty(t, group.ValueString("unknown"))
}

func TestGroup_StringMap(t *testing.T) {
	t.Parallel()

	assert.Nil(t, conv.Group(nil).StringMap())
	assert.Equal(t, map[string]string{}, conv.Group{}.StringMap())
	assert.Equal(
		t,
		map[string]string{"key1": "12", "key2": "<NIL>"},
		conv.Group{"key1": 12, "key2": nil}.StringMap(),
	)
}
