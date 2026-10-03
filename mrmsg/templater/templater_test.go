package templater_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrmsg/templater"
)

func TestTemplater_Render(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		leftDelim  string
		rightDelim string
		opts       []templater.Option
		message    string
		data       map[string]string
		want       string
		wantErr    bool
	}

	tests := []testCase{
		{
			name:    "empty message",
			message: "",
			want:    "",
		},
		{
			name:    "message without delimiters",
			message: "plain <b>message</b> 'quoted'",
			want:    "plain <b>message</b> 'quoted'",
		},
		{
			name:    "text mode by default: value is not escaped",
			message: "email: {{.Email}}",
			data:    map[string]string{"Email": "user+tag@x.com"},
			want:    "email: user+tag@x.com",
		},
		{
			name:    "text mode by default: special chars are not escaped",
			message: "ua: {{.UA}}",
			data:    map[string]string{"UA": `'"<>&`},
			want:    `ua: '"<>&`,
		},
		{
			name:    "text mode explicitly",
			opts:    []templater.Option{templater.WithMode(templater.ModeText)},
			message: "email: {{.Email}}",
			data:    map[string]string{"Email": "user+tag@x.com"},
			want:    "email: user+tag@x.com",
		},
		{
			name:    "html mode: value is escaped",
			opts:    []templater.Option{templater.WithMode(templater.ModeHTML)},
			message: "email: {{.Email}}",
			data:    map[string]string{"Email": "user+tag@x.com"},
			want:    "email: user&#43;tag@x.com",
		},
		{
			name:    "html mode: special chars are escaped",
			opts:    []templater.Option{templater.WithMode(templater.ModeHTML)},
			message: "ua: {{.UA}}",
			data:    map[string]string{"UA": `'"<>&`},
			want:    "ua: &#39;&#34;&lt;&gt;&amp;",
		},
		{
			name:    "html mode: template text is not escaped",
			opts:    []templater.Option{templater.WithMode(templater.ModeHTML)},
			message: "<b>{{.X}}</b>",
			data:    map[string]string{"X": "<i>"},
			want:    "<b>&lt;i&gt;</b>",
		},
		{
			name:    "text mode: missing key is empty",
			message: "[{{.Missing}}]",
			data:    map[string]string{},
			want:    "[]",
		},
		{
			name:    "html mode: missing key is empty",
			opts:    []templater.Option{templater.WithMode(templater.ModeHTML)},
			message: "[{{.Missing}}]",
			data:    map[string]string{},
			want:    "[]",
		},
		{
			name:       "custom delimiters",
			leftDelim:  "[[",
			rightDelim: "]]",
			message:    "hello [[.Name]] {{.Name}}",
			data:       map[string]string{"Name": "Bob"},
			want:       "hello Bob {{.Name}}",
		},
		{
			name:       "html mode: custom delimiters",
			leftDelim:  "[[",
			rightDelim: "]]",
			opts:       []templater.Option{templater.WithMode(templater.ModeHTML)},
			message:    "<b>[[.X]]</b> {{.X}}",
			data:       map[string]string{"X": "<i>"},
			want:       "<b>&lt;i&gt;</b> {{.X}}",
		},
		{
			name:    "unknown mode falls back to html",
			opts:    []templater.Option{templater.WithMode(templater.Mode(255))},
			message: "{{.X}}",
			data:    map[string]string{"X": "<i>"},
			want:    "&lt;i&gt;",
		},
		{
			name:    "text mode: invalid template",
			message: "hello {{.Name",
			wantErr: true,
		},
		{
			name:    "html mode: invalid template",
			opts:    []templater.Option{templater.WithMode(templater.ModeHTML)},
			message: "hello {{.Name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tpl := templater.NewTemplater(tt.leftDelim, tt.rightDelim, tt.opts...)

			var buf strings.Builder

			got, err := tpl.Render(tt.message, tt.data)
			errTo := tpl.RenderTo(&buf, tt.message, tt.data)

			if tt.wantErr {
				require.Error(t, err)
				require.Error(t, errTo)

				return
			}

			require.NoError(t, err)
			require.NoError(t, errTo)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
