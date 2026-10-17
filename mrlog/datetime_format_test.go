package mrlog_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog"
)

func TestParseDateTimeFormat(t *testing.T) {
	t.Parallel()

	type testCase struct {
		value string
		want  string
	}

	tests := []testCase{
		{value: "RFC3339", want: time.RFC3339},
		{value: "RFC3339Nano", want: time.RFC3339Nano},
		{value: "DateTime", want: time.DateTime},
		{value: "TimeOnly", want: time.TimeOnly},
		{value: "Kitchen", want: time.Kitchen},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			got, err := mrlog.ParseDateTimeFormat(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateTimeFormatError(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "rfc3339", time.RFC3339} {
		got, err := mrlog.ParseDateTimeFormat(value)
		require.Error(t, err)
		assert.Equal(t, time.RFC3339, got)
	}
}
