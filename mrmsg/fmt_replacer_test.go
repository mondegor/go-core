package mrmsg_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrmsg"
)

func TestFmtReplacer_Replace(t *testing.T) {
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
			name:    "empty message",
			message: "",
			args:    []any{"value1"},
			want:    "",
		},
		{
			name:    "without placeholders: message processed by fmt",
			message: "100%% done",
			args:    []any{"value1"},
			want:    "100% done",
		},
		{
			name:         "args equal placeholders",
			message:      "Hello %[1]s, you are %[2]d",
			placeholders: []string{"{name}", "{age}"},
			args:         []any{"Ivan", 30},
			want:         "Hello Ivan, you are 30",
		},
		{
			name:         "extra args are ignored",
			message:      "Hello %[1]s",
			placeholders: []string{"{name}"},
			args:         []any{"Ivan", 30},
			want:         "Hello Ivan",
		},
		{
			name:         "missing args",
			message:      "Hello %[1]s, you are %[2]s",
			placeholders: []string{"{name}", "{age}"},
			args:         []any{"Ivan"},
			want:         "Hello Ivan, you are !MISSINGARG",
		},
		{
			name:         "without args",
			message:      "Hello %[1]s",
			placeholders: []string{"{name}"},
			args:         nil,
			want:         "Hello !MISSINGARG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			replacer := mrmsg.NewFmtReplacer(tt.message, tt.placeholders)

			got, err := replacer.Replace(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			var buf strings.Builder

			require.NoError(t, replacer.ReplaceTo(&buf, tt.args))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestFmtReplacer_ReplaceKeepsCallerArgs(t *testing.T) {
	t.Parallel()

	replacer := mrmsg.NewFmtReplacer("%[1]s %[2]s", []string{"{a}", "{b}"})

	backing := []any{"value1", "keep"}
	args := backing[:1] // свободная ёмкость слайса аргументов не должна перезаписываться

	got, err := replacer.Replace(args)
	require.NoError(t, err)
	assert.Equal(t, "value1 !MISSINGARG", got)
	assert.Equal(t, "keep", backing[1])
}

func TestFmtReplacer_ReplaceToWriteError(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, mrmsg.NewFmtReplacer("message", nil).ReplaceTo(failWriter{}, nil), errWrite)
	require.ErrorIs(t, mrmsg.NewFmtReplacer("%[1]s", []string{"{a}"}).ReplaceTo(failWriter{}, nil), errWrite)
	require.NoError(t, mrmsg.NewFmtReplacer("", nil).ReplaceTo(failWriter{}, nil))
}

func TestFmtReplacer_CountArgs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, mrmsg.NewFmtReplacer("message", nil).CountArgs())
	assert.Equal(t, 2, mrmsg.NewFmtReplacer("%[1]s %[2]s", []string{"{a}", "{b}"}).CountArgs())
}
