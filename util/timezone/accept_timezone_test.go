package timezone_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/timezone"
)

func TestParseAcceptTimeZone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    timezone.AcceptTimeZone
		wantErr bool
	}{
		{
			name:  "name only",
			value: "Asia/Tokyo",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			name:  "name with parameters",
			value: "Asia/Tokyo;offset=+09:00;dst=0",
			want: timezone.AcceptTimeZone{
				Name:      "Asia/Tokyo",
				Offset:    9 * time.Hour,
				HasOffset: true,
			},
		},
		{
			name:  "offset without a name",
			value: "offset=+09:00;dst=0",
			want: timezone.AcceptTimeZone{
				Offset:    9 * time.Hour,
				HasOffset: true,
			},
		},
		{
			name:  "negative offset with dst flag",
			value: "offset=-07:30;dst=1",
			want: timezone.AcceptTimeZone{
				Offset:    -(7*time.Hour + 30*time.Minute),
				IsDST:     true,
				HasOffset: true,
			},
		},
		{
			// подбор по одному лишь смещению давал бы пояс наугад,
			// поэтому без признака летнего времени смещение не принимается
			name:  "offset without the dst flag is not accepted",
			value: "Asia/Tokyo;offset=+09:00",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			name:  "dst flag without an offset is ignored",
			value: "Asia/Tokyo;dst=0",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			// знак обязателен: "09:00" за смещение с угаданным знаком не принимается
			name:  "offset without a sign is not accepted",
			value: "Asia/Tokyo;offset=09:00;dst=0",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			// разбор строгий: пробел перед ключом делает негодным весь параметр
			name:  "spaces around the segments are not tolerated",
			value: "Asia/Tokyo; offset=+09:00; dst=0",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			// пробел после "=" делает негодным значение параметра
			name:  "spaces around the parameter value are not tolerated",
			value: "Asia/Tokyo;offset= +09:00;dst= 0",
			want:  timezone.AcceptTimeZone{Name: "Asia/Tokyo"},
		},
		{
			// имя не подрезается: пробелы - часть значения первого сегмента
			name:  "name is not trimmed",
			value: " Asia/Tokyo ;offset=+09:00;dst=0",
			want: timezone.AcceptTimeZone{
				Name:      " Asia/Tokyo ",
				Offset:    9 * time.Hour,
				HasOffset: true,
			},
		},
		{
			// потолок - 4 сегмента, поэтому dst остаётся за ним и отбрасывается
			// вместе со смещением, а первый сегмент именем пояса не является
			name:  "segments beyond the items limit are dropped",
			value: "a;b;c;offset=+09:00;dst=0",
			want:  timezone.AcceptTimeZone{Name: "a"},
		},
		{
			name:    "no name and no offset returns an error",
			value:   ";;;",
			wantErr: true,
		},
		{
			name:    "parameters only, none of them usable, returns an error",
			value:   "=;offset=nonsense",
			wantErr: true,
		},
		{
			name:    "empty value returns an error",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := timezone.ParseAcceptTimeZone(tt.value)

			if tt.wantErr {
				require.ErrorIs(t, err, timezone.ErrInvalidAcceptTimeZone)
				assert.Equal(t, timezone.AcceptTimeZone{}, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
