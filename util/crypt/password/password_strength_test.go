package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/crypt/password"
)

// TestCalcStrength проверяет границы уровней надёжности по числу уникальных символов
// и числу используемых групп символов.
func TestCalcStrength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  password.PassStrength
	}{
		{name: "empty", value: "", want: password.PassStrengthNotRated},
		{name: "one set", value: "abcdefghijklmnop", want: password.PassStrengthNotRated},
		{name: "4 sets, 6 uniq", value: "aB1!cD", want: password.PassStrengthNotRated},
		{name: "4 sets, 7 uniq", value: "aB1!cD2", want: password.PassStrengthWeak},
		{name: "4 sets, 9 uniq", value: "aB1!cD2#e", want: password.PassStrengthMedium},
		{name: "4 sets, 10 uniq", value: "aB1!cD2#eF", want: password.PassStrengthStrong},
		{name: "4 sets, 11 uniq", value: "aB1!cD2#eF3", want: password.PassStrengthStrong},
		{name: "4 sets, 12 uniq", value: "aB1!cD2#eF3$", want: password.PassStrengthBest},
		{name: "4 sets, repeated chars", value: "aB1!aB1!aB1!aB1!", want: password.PassStrengthNotRated},
		{name: "lower, upper, signs, 8 uniq", value: "aB!cD#eF", want: password.PassStrengthWeak},
		{name: "lower, upper, signs, 9 uniq", value: "aB!cD#eF$", want: password.PassStrengthMedium},
		{name: "lower, upper, signs, 10 uniq", value: "aB!cD#eF$g", want: password.PassStrengthMedium},
		{name: "lower, upper, signs, 11 uniq", value: "aB!cD#eF$gH", want: password.PassStrengthStrong},
		{name: "numerals, lower, upper, 10 uniq", value: "aB1cD2eF3g", want: password.PassStrengthMedium},
		{name: "numerals, lower, upper, 11 uniq", value: "aB1cD2eF3gH", want: password.PassStrengthStrong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, password.CalcStrength(tt.value))
		})
	}
}
