package conv_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/conv"
)

func TestString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value any
		want  string
	}

	tests := []testCase{
		{name: "string", value: "abc", want: "abc"},
		{name: "int", value: -1, want: "-1"},
		{name: "uint", value: uint(1), want: "1"},
		{name: "int64", value: int64(-64), want: "-64"},
		{name: "uint64", value: uint64(64), want: "64"},
		{name: "bool", value: true, want: "true"},
		{name: "int8", value: int8(-8), want: "-8"},
		{name: "uint8", value: uint8(8), want: "8"},
		{name: "int16", value: int16(-16), want: "-16"},
		{name: "uint16", value: uint16(16), want: "16"},
		{name: "int32", value: int32(-32), want: "-32"},
		{name: "uint32", value: uint32(32), want: "32"},
		{name: "float64", value: 1.5, want: "1.5"},
		{name: "float32", value: float32(0.1), want: "0.1"},
		{name: "nil", value: nil, want: "<NIL>"},
		{name: "error", value: errors.New("err text"), want: "err text"},
		{name: "stringer", value: time.Second, want: "1s"},
		{name: "other", value: []int{1, 2}, want: "[1 2]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, conv.String(tt.value))
		})
	}
}
