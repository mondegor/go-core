package itemstatus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrworkflow/itemstatus"
)

func TestEnum_Scan(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		value   any
		want    itemstatus.Enum
		wantErr string
	}

	tests := []testCase{
		{
			name:  "first",
			value: int64(1),
			want:  itemstatus.Draft,
		},
		{
			name:  "last",
			value: int64(3),
			want:  itemstatus.Disabled,
		},
		{
			name:    "zero",
			value:   int64(0),
			wantErr: "value is not found in enum set (value=0, enum=ItemStatus)",
		},
		{
			name:    "after_last",
			value:   int64(4),
			wantErr: "value is not found in enum set (value=4, enum=ItemStatus)",
		},
		{
			name:    "max_uint8",
			value:   int64(255),
			wantErr: "value is not found in enum set (value=255, enum=ItemStatus)",
		},
		{
			name:    "negative",
			value:   int64(-1),
			wantErr: "value is out of range (type=ItemStatus, value=-1)",
		},
		{
			// без проверки диапазона uint8(256) превратился бы в 0
			name:    "greater_than_uint8",
			value:   int64(256),
			wantErr: "value is out of range (type=ItemStatus, value=256)",
		},
		{
			name:    "string",
			value:   "1",
			wantErr: `invalid type assertion (type=ItemStatus, value="1")`,
		},
		{
			name:    "bytes",
			value:   []byte("1"),
			wantErr: "invalid type assertion (type=ItemStatus, value=[]byte{0x31})",
		},
		{
			name:    "nil",
			value:   nil,
			wantErr: "invalid type assertion (type=ItemStatus, value=<nil>)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := itemstatus.Enabled
			err := got.Scan(tt.value)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Equal(t, itemstatus.Enabled, got) // при ошибке значение не меняется

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
