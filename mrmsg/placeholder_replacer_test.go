package mrmsg_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrmsg"
)

func Test_NewMessageReplacer(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		message string
		args    []any
		want    string
	}

	tests := []testCase{
		{
			name:    "empty message",
			message: "",
			args:    nil,
			want:    "",
		},
		{
			name:    "message with arg",
			message: "test message {Arg1}",
			args:    []any{"value1"},
			want:    "test message value1",
		},
		{
			name:    "message with arg without value",
			message: "test message {Arg1}",
			args:    nil,
			want:    "test message !MISSINGARG",
		},
		{
			name:    "message with a few placeholders",
			message: "{Arg1} test {Arg2} message {Arg3}",
			args:    []any{"value1", "value2", "value3"},
			want:    "value1 test value2 message value3",
		},
		{
			name:    "message with missed }",
			message: "message {Arg1",
			args:    []any{"value1"},
			want:    "message {Arg1",
		},
		{
			name:    "message with missed } 2",
			message: "message {Arg1 {Arg2}",
			args:    []any{"value1", "value2"},
			want:    "message {Arg1 value1",
		},
		{
			name:    "message with missed {",
			message: "message {Arg1} Arg2}",
			args:    []any{"value1", "value2"},
			want:    "message value1 Arg2}",
		},
		{
			name:    "message is missing arg1",
			message: "test message {Arg2}",
			args:    []any{"value1", "value2"},
			want:    "test message value1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			replacer := mrmsg.NewMessageReplacer("{", "}", tt.message)

			got, err := replacer.Replace(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlaceholderReplacer_Replace(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		message      string
		placeholders []string
		args         []any
		want         string
	}

	tests := []testCase{
		{
			name:    "without placeholders: message as is",
			message: "message {Arg1}",
			args:    []any{"value1"},
			want:    "message {Arg1}",
		},
		{
			name:         "extra args are ignored",
			message:      "message {Arg1}",
			placeholders: []string{"{Arg1}"},
			args:         []any{1, 2},
			want:         "message 1",
		},
		{
			name:         "repeated placeholder",
			message:      "{Arg1} and {Arg1}",
			placeholders: []string{"{Arg1}"},
			args:         []any{nil},
			want:         "<NIL> and <NIL>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			replacer := mrmsg.NewPlaceholderReplacer(tt.message, tt.placeholders)

			got, err := replacer.Replace(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			var buf strings.Builder

			require.NoError(t, replacer.ReplaceTo(&buf, tt.args))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestPlaceholderReplacer_ReplaceTo(t *testing.T) {
	t.Parallel()

	var buf strings.Builder

	replacer := mrmsg.NewMessageReplacer("", "", "{A} {B}")

	require.NoError(t, replacer.ReplaceTo(&buf, []any{1}))
	assert.Equal(t, "1 !MISSINGARG", buf.String())

	require.ErrorIs(t, replacer.ReplaceTo(failWriter{}, nil), errWrite)
	require.ErrorIs(t, mrmsg.NewPlaceholderReplacer("message", nil).ReplaceTo(failWriter{}, nil), errWrite)
	require.NoError(t, mrmsg.NewPlaceholderReplacer("", nil).ReplaceTo(failWriter{}, nil))
}

func TestPlaceholderReplacer_CountArgs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, mrmsg.NewMessageReplacer("{", "}", "message").CountArgs())
	assert.Equal(t, 2, mrmsg.NewMessageReplacer("{", "}", "{A} {B} {A}").CountArgs())
}
