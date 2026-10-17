package level_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog/level"
)

func TestEnum_String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		lvl  level.Enum
		want string
	}

	tests := []testCase{
		{lvl: level.Debug, want: "DEBUG"},
		{lvl: level.Info, want: "INFO"},
		{lvl: level.Warn, want: "WARN"},
		{lvl: level.Error, want: "ERROR"},
		{lvl: level.Fatal, want: "FATAL"},
		{lvl: level.Trace, want: "TRACE"},
		{lvl: level.Enum(100), want: "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.lvl.String())
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	for _, lvl := range []level.Enum{level.Debug, level.Info, level.Warn, level.Error} {
		got, err := level.Parse(lvl.String())
		require.NoError(t, err)
		assert.Equal(t, lvl, got)
	}

	// служебные уровни в конфиге не назначаются
	for _, value := range []string{"FATAL", "TRACE", "debug", ""} {
		got, err := level.Parse(value)
		require.Error(t, err)
		assert.Equal(t, level.Error, got)
	}
}
