package mrpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrpath"
)

func TestNewTail(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		basePath string
		path     string
		want     string
	}

	tests := []testCase{
		{name: "base with slash", basePath: "/base/dir/", path: "file.txt", want: "/base/dir/file.txt"},
		{name: "base without slash is glued", basePath: "/base/dir-", path: "file.txt", want: "/base/dir-file.txt"},
		{name: "base with many slashes", basePath: "/base/dir///", path: "file.txt", want: "/base/dir/file.txt"},
		{name: "slashes of url scheme are kept", basePath: "https://", path: "host/x", want: "https://host/x"},
		{name: "slashes of file scheme are kept", basePath: "file:///", path: "x", want: "file:///x"},
		{name: "base of slashes only", basePath: "///", path: "file.txt", want: "/file.txt"},
		{name: "leading slashes of path are trimmed", basePath: "/base/", path: "//sub/file/", want: "/base/sub/file/"},
		{name: "inner double slashes are kept", basePath: "/a//b/", path: "x//y", want: "/a//b/x//y"},
		{name: "empty base", basePath: "", path: "/file.txt/", want: "file.txt/"},
		{name: "empty path", basePath: "/base", path: "", want: ""},
		{name: "single slash", basePath: "/base", path: "/", want: ""},
		{name: "path of slashes only", basePath: "/base", path: "//", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, mrpath.NewTail(tt.basePath).BuildPath(tt.path))
		})
	}
}
