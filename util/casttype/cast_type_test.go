package casttype_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/casttype"
)

func TestBoolToNumber(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 1, casttype.BoolToNumber[int](true))
	assert.Equal(t, 0, casttype.BoolToNumber[int](false))
	assert.InDelta(t, 1.0, casttype.BoolToNumber[float64](true), 0)
	assert.Equal(t, uint8(0), casttype.BoolToNumber[uint8](false))
}

func TestBoolToPointer(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		value    bool
		required []bool
		wantNil  bool
	}

	tests := []testCase{
		{name: "true", value: true},
		{name: "false is nil", value: false, wantNil: true},
		{name: "false not required is nil", value: false, required: []bool{false}, wantNil: true},
		{name: "false required", value: false, required: []bool{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := casttype.BoolToPointer(tt.value, tt.required...)
			if tt.wantNil {
				assert.Nil(t, got)

				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.value, *got)
		})
	}
}

func TestNumberToPointer(t *testing.T) {
	t.Parallel()

	got := casttype.NumberToPointer(5)
	require.NotNil(t, got)
	assert.Equal(t, 5, *got)

	assert.Nil(t, casttype.NumberToPointer(0))
	assert.Nil(t, casttype.NumberToPointer(0.0, false))

	got = casttype.NumberToPointer(0, true)
	require.NotNil(t, got)
	assert.Equal(t, 0, *got)
}

func TestStringToPointer(t *testing.T) {
	t.Parallel()

	got := casttype.StringToPointer("abc")
	require.NotNil(t, got)
	assert.Equal(t, "abc", *got)

	assert.Nil(t, casttype.StringToPointer(""))

	got = casttype.StringToPointer("", true)
	require.NotNil(t, got)
	assert.Empty(t, *got)
}

func TestTimeToPointer(t *testing.T) {
	t.Parallel()

	now := time.Now()

	got := casttype.TimeToPointer(now)
	require.NotNil(t, got)
	assert.Equal(t, now, *got)

	assert.Nil(t, casttype.TimeToPointer(time.Time{}))

	got = casttype.TimeToPointer(time.Time{}, true)
	require.NotNil(t, got)
	assert.True(t, got.IsZero())
}
