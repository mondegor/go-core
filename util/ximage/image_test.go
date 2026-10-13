package ximage_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/ximage"
)

type (
	// failSeeker - читает данные, но не умеет перемещаться по ним.
	failSeeker struct {
		io.Reader
	}
)

func (failSeeker) Seek(int64, int) (int64, error) {
	return 0, errors.New("seek failed")
}

func encodeImage(t *testing.T, format string) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	img.Set(0, 0, color.White)

	var (
		buf bytes.Buffer
		err error
	)

	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	}

	require.NoError(t, err)

	return buf.Bytes()
}

func TestDecodeImageConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		format      string
		contentType string
	}

	tests := []testCase{
		{name: "jpeg", format: "jpeg", contentType: "image/jpeg"},
		{name: "jpg without prefix", format: "jpeg", contentType: "jpg"},
		{name: "jpe", format: "jpeg", contentType: "image/jpe"},
		{name: "png", format: "png", contentType: "image/png"},
		{name: "gif", format: "gif", contentType: "image/gif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file := bytes.NewReader(encodeImage(t, tt.format))

			cfg, err := ximage.DecodeImageConfig(file, tt.contentType)
			require.NoError(t, err)
			assert.Equal(t, 3, cfg.Width)
			assert.Equal(t, 2, cfg.Height)

			// позиция чтения возвращается в начало файла
			pos, err := file.Seek(0, io.SeekCurrent)
			require.NoError(t, err)
			assert.Zero(t, pos)
		})
	}
}

func TestDecodeImageConfigErrors(t *testing.T) {
	t.Parallel()

	_, err := ximage.DecodeImageConfig(bytes.NewReader(encodeImage(t, "png")), "image/webp")
	require.Error(t, err)

	_, err = ximage.DecodeImageConfig(bytes.NewReader([]byte("not an image")), "image/png")
	require.Error(t, err)

	_, err = ximage.DecodeImageConfig(failSeeker{Reader: bytes.NewReader(encodeImage(t, "png"))}, "image/png")
	require.Error(t, err)
}

func TestDecodeImage(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"jpeg", "png", "gif"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			file := bytes.NewReader(encodeImage(t, format))

			img, err := ximage.DecodeImage(file, "image/"+format)
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, 3, 2), img.Bounds())

			pos, err := file.Seek(0, io.SeekCurrent)
			require.NoError(t, err)
			assert.Zero(t, pos)
		})
	}
}

func TestDecodeImageErrors(t *testing.T) {
	t.Parallel()

	_, err := ximage.DecodeImage(bytes.NewReader(encodeImage(t, "png")), "image/webp")
	require.Error(t, err)

	_, err = ximage.DecodeImage(bytes.NewReader([]byte("not an image")), "image/gif")
	require.Error(t, err)

	_, err = ximage.DecodeImage(failSeeker{Reader: bytes.NewReader(encodeImage(t, "png"))}, "image/png")
	require.Error(t, err)
}

func TestCheckImage(t *testing.T) {
	t.Parallel()

	require.NoError(t, ximage.CheckImage(bytes.NewReader(encodeImage(t, "png")), "image/png"))
	require.Error(t, ximage.CheckImage(bytes.NewReader([]byte("bad")), "image/png"))
}
