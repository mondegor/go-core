package errors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mrerrors "github.com/mondegor/go-core/mrtype/errors"
)

func TestParamMaxValueError_Error(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		maxValue any
		want     string
	}

	tests := []testCase{
		{
			name:     "int",
			maxValue: 10,
			want:     "param contains value greater than max (type=Param, maxValue=10)",
		},
		{
			name:     "float",
			maxValue: 1.5,
			want:     "param contains value greater than max (type=Param, maxValue=1.5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.EqualError(t, mrerrors.NewParamMaxValueError("Param", tt.maxValue), tt.want)
		})
	}
}
