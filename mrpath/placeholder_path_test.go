package mrpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrpath"
)

func TestNewPlaceholder(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		basePath    string
		placeholder string
		path        string
		want        string
	}

	tests := []testCase{
		{
			name:        "placeholder in middle",
			basePath:    "/dir/{{path}}/postfix",
			placeholder: "{{path}}",
			path:        "real-value",
			want:        "/dir/real-value/postfix",
		},
		{
			name:        "slashes of path are trimmed",
			basePath:    "/dir/{{path}}/postfix",
			placeholder: "{{path}}",
			path:        "//sub/value//",
			want:        "/dir/sub/value/postfix",
		},
		{
			name:        "placeholder at end",
			basePath:    "https://host/{id}",
			placeholder: "{id}",
			path:        "123",
			want:        "https://host/123",
		},
		{
			name:        "slashes of url scheme are kept",
			basePath:    "https://{id}/img",
			placeholder: "{id}",
			path:        "cdn.ru",
			want:        "https://cdn.ru/img",
		},
		{
			name:        "slashes of file scheme are kept",
			basePath:    "file:///{{path}}",
			placeholder: "{{path}}",
			path:        "x",
			want:        "file:///x",
		},
		{
			name:        "placeholder at end keeps trailing slashes of path",
			basePath:    "/dir/{{path}}",
			placeholder: "{{path}}",
			path:        "//sub/value/",
			want:        "/dir/sub/value/",
		},
		{
			name:        "placeholder glued to prefix",
			basePath:    "/dir-{{path}}",
			placeholder: "{{path}}",
			path:        "/x",
			want:        "/dir-x",
		},
		{
			name:        "slashes of prefix are collapsed",
			basePath:    "/dir//{{path}}",
			placeholder: "{{path}}",
			path:        "x",
			want:        "/dir/x",
		},
		{
			name:        "slashes of prefix are collapsed before postfix",
			basePath:    "/dir//{{path}}/postfix",
			placeholder: "{{path}}",
			path:        "x",
			want:        "/dir/x/postfix",
		},
		{
			name:        "slashes of postfix are collapsed",
			basePath:    "/dir/{{path}}//postfix",
			placeholder: "{{path}}",
			path:        "x",
			want:        "/dir/x/postfix",
		},
		{
			name:        "postfix of slashes only",
			basePath:    "/dir/{{path}}//",
			placeholder: "{{path}}",
			path:        "x/",
			want:        "/dir/x/",
		},
		{
			name:        "slashes of path are trimmed before postfix slash",
			basePath:    "/dir/{{path}}/",
			placeholder: "{{path}}",
			path:        "//sub/value/",
			want:        "/dir/sub/value/",
		},
		{
			name:        "slashes of path are trimmed before postfix extension",
			basePath:    "/dir/{{path}}.png",
			placeholder: "{{path}}",
			path:        "value/",
			want:        "/dir/value.png",
		},
		{
			name:        "placeholder at start",
			basePath:    "{{path}}/postfix",
			placeholder: "{{path}}",
			path:        "value",
			want:        "value/postfix",
		},
		{
			name:        "placeholder is whole base path",
			basePath:    "{{path}}",
			placeholder: "{{path}}",
			path:        "/value/",
			want:        "value/",
		},
		{
			name:        "inner double slashes are kept",
			basePath:    "/a//b/{{path}}/c",
			placeholder: "{{path}}",
			path:        "x//y",
			want:        "/a//b/x//y/c",
		},
		{
			name:        "only first placeholder is replaced",
			basePath:    "/a/{{path}}/b/{{path}}",
			placeholder: "{{path}}",
			path:        "x",
			want:        "/a/x/b/{{path}}",
		},
		{
			name:        "empty path",
			basePath:    "/dir/{{path}}",
			placeholder: "{{path}}",
			path:        "",
			want:        "",
		},
		{
			name:        "single slash",
			basePath:    "/dir/{{path}}",
			placeholder: "{{path}}",
			path:        "/",
			want:        "",
		},
		{
			name:        "path of slashes only with postfix",
			basePath:    "/dir/{{path}}/postfix",
			placeholder: "{{path}}",
			path:        "//",
			want:        "",
		},
		{
			name:        "path of slashes only",
			basePath:    "/dir/{{path}}",
			placeholder: "{{path}}",
			path:        "//",
			want:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder, err := mrpath.NewPlaceholder(tt.basePath, tt.placeholder)
			require.NoError(t, err)
			assert.Equal(t, tt.want, builder.BuildPath(tt.path))
		})
	}
}

func TestNewPlaceholderPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		basePath string
		path     string
		want     string
	}

	tests := []testCase{
		{name: "placeholder in middle", basePath: "/dir/{{path}}/postfix", path: "x", want: "/dir/x/postfix"},
		{name: "placeholder at end", basePath: "/dir/{{path}}", path: "/x/", want: "/dir/x/"},
		{name: "empty path", basePath: "/dir/{{path}}", path: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder, err := mrpath.NewPlaceholderPath(tt.basePath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, builder.BuildPath(tt.path))
		})
	}
}

func TestNewPlaceholderErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		basePath    string
		placeholder string
	}

	tests := []testCase{
		{name: "empty placeholder", basePath: "/dir/{{path}}", placeholder: ""},
		{name: "placeholder not found", basePath: "/dir/path", placeholder: "{{path}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder, err := mrpath.NewPlaceholder(tt.basePath, tt.placeholder)
			require.Error(t, err)
			assert.Nil(t, builder)
		})
	}
}

func TestNewPlaceholderPathError(t *testing.T) {
	t.Parallel()

	builder, err := mrpath.NewPlaceholderPath("/dir/{id}")
	require.Error(t, err)
	assert.Nil(t, builder)
}

func TestNewPlaceholderAtEndEqualsTail(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		template string
		basePath string
	}

	tests := []testCase{
		{name: "prefix with many slashes", template: "/dir//{{path}}", basePath: "/dir//"},
		{name: "prefix without slash", template: "/dir-{{path}}", basePath: "/dir-"},
		{name: "empty prefix", template: "{{path}}", basePath: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder, err := mrpath.NewPlaceholder(tt.template, "{{path}}")
			require.NoError(t, err)

			tail := mrpath.NewTail(tt.basePath)

			for _, path := range []string{"x", "//a/b/", "", "/"} {
				assert.Equal(t, tail.BuildPath(path), builder.BuildPath(path), "path=%q", path)
			}
		})
	}
}
