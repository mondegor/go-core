package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/mime"
	"github.com/mondegor/go-core/util/mime/config"
)

func TestValidateMimeTypes(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		mimeTypes []mime.Type
		wantErr   bool
	}

	tests := []testCase{
		{name: "empty list", mimeTypes: nil},
		{
			name: "valid",
			mimeTypes: []mime.Type{
				{ContentType: "image/jpeg", Extension: "jpg"},
				{ContentType: "image/jpeg", Extension: "jpeg"},
			},
		},
		{
			name:      "empty extension",
			mimeTypes: []mime.Type{{ContentType: "image/jpeg"}},
			wantErr:   true,
		},
		{
			name: "duplicate extension",
			mimeTypes: []mime.Type{
				{ContentType: "image/jpeg", Extension: "jpg"},
				{ContentType: "image/pjpeg", Extension: "jpg"},
			},
			wantErr: true,
		},
		{
			name: "duplicate extension with dot",
			mimeTypes: []mime.Type{
				{ContentType: "image/jpeg", Extension: "jpg"},
				{ContentType: "image/pjpeg", Extension: ".jpg"},
			},
			wantErr: true,
		},
		{
			name:      "dot only extension",
			mimeTypes: []mime.Type{{ContentType: "image/jpeg", Extension: "."}},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateMimeTypes(tt.mimeTypes)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
