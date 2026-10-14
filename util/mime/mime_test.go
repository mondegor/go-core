package mime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/mime"
)

func newTypeList() *mime.TypeList {
	return mime.NewTypeList(
		[]mime.Type{
			{ContentType: "image/jpeg", Extension: ".jpg"},
			{ContentType: "image/jpeg", Extension: "jpeg"},
			{ContentType: "image/png", Extension: "png"},
		},
	)
}

func TestTypeList_ContentTypeByExt(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		ext     string
		want    string
		wantErr bool
	}

	tests := []testCase{
		{name: "registered with dot", ext: "jpg", want: "image/jpeg"},
		{name: "second ext of type", ext: "jpeg", want: "image/jpeg"},
		{name: "arg with dot", ext: ".png", want: "image/png"},
		{name: "unknown", ext: "gif", wantErr: true},
		{name: "empty", ext: "", wantErr: true},
		{name: "dot only", ext: ".", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newTypeList().ContentTypeByExt(tt.ext)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeList_ExtByContentType(t *testing.T) {
	t.Parallel()

	list := newTypeList()

	got, err := list.ExtByContentType("image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, "jpg", got) // первое зарегистрированное расширение типа

	_, err = list.ExtByContentType("image/gif")
	require.Error(t, err)

	_, err = list.ExtByContentType("")
	require.Error(t, err)
}

func TestTypeList_TypesByExts(t *testing.T) {
	t.Parallel()

	list := newTypeList()

	got, err := list.TypesByExts([]string{"png", ".jpeg"})
	require.NoError(t, err)
	assert.Equal(
		t,
		[]mime.Type{
			{ContentType: "image/png", Extension: "png"},
			{ContentType: "image/jpeg", Extension: "jpeg"},
		},
		got,
	)

	got, err = list.TypesByExts([]string{"png", "gif"})
	require.Error(t, err)
	assert.Nil(t, got)
}
