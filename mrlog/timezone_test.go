package mrlog_test

// база часовых поясов встраивается в тестовый бинарник, чтобы тесты
// проходили в минимальных образах, где она отсутствует в системе.
import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog"
)

func TestParseTimeZone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    *time.Location
		wantErr bool
	}{
		{
			name:  "utc",
			value: "UTC",
			want:  time.UTC,
		},
		{
			name:  "local",
			value: "Local",
			want:  time.Local,
		},
		{
			name:  "iana name",
			value: "Europe/Moscow",
			want:  mustLoadLocation(t, "Europe/Moscow"),
		},
		{
			name:    "empty value returns utc and error",
			value:   "",
			want:    time.UTC,
			wantErr: true,
		},
		{
			name:    "unknown value returns utc and error",
			value:   "Nowhere/Nowhere",
			want:    time.UTC,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := mrlog.ParseTimeZone(tt.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

// mustLoadLocation - загружает часовой пояс или прерывает тест.
func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()

	loc, err := time.LoadLocation(name)
	require.NoError(t, err)

	return loc
}
