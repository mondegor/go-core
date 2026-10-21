package gotext_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrlocale/provider/gotext"
)

func TestMessageConverter(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		message string
		args    []any
		wantMsg string
		want    string
	}

	tests := []testCase{
		{
			name:    "placeholders converted",
			message: "Hello {Name}, you are {Age}",
			args:    []any{"Ivan", "30"},
			wantMsg: "Hello %[1]s, you are %[2]s",
			want:    "Hello Ivan, you are 30",
		},
		{
			name:    "percent escaped",
			message: "100% done {Name}",
			args:    []any{"Ivan"},
			wantMsg: "100%% done %[1]s",
			want:    "100% done Ivan",
		},
		{
			name:    "percent escaped without placeholders",
			message: "100% done",
			wantMsg: "100%% done",
			want:    "100% done",
		},
	}

	converter := gotext.MessageConverter("{", "}")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMsg, gotArgs := converter(tt.message, tt.args)
			assert.Equal(t, tt.wantMsg, gotMsg)
			assert.Equal(t, tt.args, gotArgs)
			assert.Equal(t, tt.want, fmt.Sprintf(gotMsg, gotArgs...))
		})
	}
}

// TestNewMessageFormatter - плейсхолдеры заменяются позиционными глаголами, повтор получает тот же номер.
func TestNewMessageFormatter(t *testing.T) {
	t.Parallel()

	message, placeholders := gotext.NewMessageFormatter("{{", "}}").Format("{{Name}} is {{Age}}, {{Name}}")
	assert.Equal(t, "%[1]s is %[2]s, %[1]s", message)
	assert.Equal(t, []string{"%[1]s", "%[2]s"}, placeholders)
}
