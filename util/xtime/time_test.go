package xtime_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xtime"
)

func TestTimeLeftInSec(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		offset time.Duration
		want   int64
	}

	// смещения выбраны вдали от целых секунд, чтобы время выполнения теста не влияло на результат
	tests := []testCase{
		{name: "past", offset: -time.Hour, want: 0},
		{name: "recently passed", offset: -200 * time.Millisecond, want: 0},
		{name: "less than second left", offset: 200 * time.Millisecond, want: 1},
		{name: "small fraction rounded up", offset: 10*time.Second + 200*time.Millisecond, want: 11},
		{name: "large fraction rounded up", offset: 10*time.Second + 800*time.Millisecond, want: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, xtime.TimeLeftInSec(time.Now().Add(tt.offset)))
		})
	}
}
