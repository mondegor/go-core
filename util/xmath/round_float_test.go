package xmath_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xmath"
)

func TestRoundFloat(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		x         float64
		precision int
		want      float64
	}

	tests := []testCase{
		{
			name:      "test1",
			x:         123.12345678,
			precision: 0,
			want:      123,
		},
		{
			name:      "test2",
			x:         123.12345678,
			precision: 4,
			want:      123.1235,
		},
		{
			name:      "test3",
			x:         123.666666,
			precision: 3,
			want:      123.667,
		},
		{
			name:      "test4",
			x:         123.333333333,
			precision: 3,
			want:      123.333,
		},
		{
			name:      "test5",
			x:         123.12000078,
			precision: 4,
			want:      123.12,
		},
		{
			name:      "test6",
			x:         123.00000078,
			precision: 5,
			want:      123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			delta := 1.0 / math.Pow(10, float64(tt.precision)+2)

			got := xmath.RoundFloat(tt.x, tt.precision)
			assert.InDelta(t, tt.want, got, delta)
		})
	}
}

func TestRoundFloatNaN(t *testing.T) {
	t.Parallel()

	got := xmath.RoundFloat(math.NaN(), 2)
	assert.True(t, math.IsNaN(got))
}

func TestRoundFloatInf(t *testing.T) {
	t.Parallel()

	got := xmath.RoundFloat(math.Inf(1), 2)
	assert.True(t, math.IsInf(got, 1))

	got = xmath.RoundFloat(math.Inf(-1), 2)
	assert.True(t, math.IsInf(got, -1))
}

func TestRoundFloatN(t *testing.T) {
	t.Parallel()

	const x = 1.123456789

	assert.InDelta(t, 1.12, xmath.RoundFloat2(x), 1e-12)
	assert.InDelta(t, 1.1235, xmath.RoundFloat4(x), 1e-12)
	assert.InDelta(t, 1.12345679, xmath.RoundFloat8(x), 1e-12)
}
