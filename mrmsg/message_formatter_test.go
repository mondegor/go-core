package mrmsg_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrmsg"
)

func Test_NewMessageFormatter(t *testing.T) {
	t.Parallel()

	formatter := mrmsg.NewMessageFormatter(
		"{",
		"}",
		func(placeholder string, index int) (newPlaceholder string) {
			return "[" + placeholder + "-" + strconv.Itoa(index) + "]"
		},
	)

	type testCase struct {
		name             string
		message          string
		wantMessage      string
		wantPlaceholders []string
	}

	tests := []testCase{
		{
			name:             "1",
			message:          "",
			wantMessage:      "",
			wantPlaceholders: nil,
		},
		{
			name:             "2",
			message:          "test message",
			wantMessage:      "test message",
			wantPlaceholders: nil,
		},
		{
			name:             "3",
			message:          "test message {Arg1}",
			wantMessage:      "test message [{Arg1}-0]",
			wantPlaceholders: []string{"[{Arg1}-0]"},
		},
		{
			name:             "4",
			message:          "test {Arg1 message",
			wantMessage:      "test {Arg1 message",
			wantPlaceholders: nil,
		},
		{
			name:             "5",
			message:          "{Arg1} test {Arg2} message {Arg3}",
			wantMessage:      "[{Arg1}-0] test [{Arg2}-1] message [{Arg3}-2]",
			wantPlaceholders: []string{"[{Arg1}-0]", "[{Arg2}-1]", "[{Arg3}-2]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMessage, gotPlaceholders := formatter.Format(tt.message)
			require.Equal(t, tt.wantMessage, gotMessage)

			assert.Equal(t, tt.wantPlaceholders, gotPlaceholders)
		})
	}
}

func TestMessageFormatter_FormatWithoutFormatter(t *testing.T) {
	t.Parallel()

	message, placeholders := mrmsg.NewMessageFormatter("{", "}", nil).Format("message {A}")
	assert.Equal(t, "message {A}", message)
	assert.Equal(t, []string{"{A}"}, placeholders)
}
