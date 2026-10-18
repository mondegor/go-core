package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEncodeBaseN проверяет, что число записывается с выравниванием по правому краю:
// недостающие старшие разряды заполняются '0', лишние старшие разряды отбрасываются.
func TestEncodeBaseN(t *testing.T) {
	t.Parallel()

	const maxValue = 4738381338321616895 // 36^12 - 1

	tests := []struct {
		name string
		n    uint64
		want string
	}{
		{name: "zero", n: 0, want: "000000000000"},
		{name: "one", n: 1, want: "000000000001"},
		{name: "last single digit", n: 35, want: "00000000000Z"},
		{name: "first two digits", n: 36, want: "000000000010"},
		{name: "max", n: maxValue, want: "ZZZZZZZZZZZZ"},
		{name: "overflow drops high digit", n: maxValue + 1, want: "000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf [bufferLen]byte

			encodeBaseN(buf[:], tt.n)
			assert.Equal(t, tt.want, string(buf[:]))
		})
	}
}
