package mrstorage_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrtype"
)

func TestNewIDCursor(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		params mrtype.CursorParams
		want   mrstorage.IDCursor
	}

	tests := []testCase{
		{
			name:   "first page",
			params: mrtype.CursorParams{Value: "", Limit: 20},
			want:   mrstorage.IDCursor{Limit: 20},
		},
		{
			name:   "record id",
			params: mrtype.CursorParams{Value: "12345", Limit: 20},
			want:   mrstorage.IDCursor{ID: 12345, Limit: 20},
		},
		{
			name:   "unparsable value is the first page",
			params: mrtype.CursorParams{Value: "abc", Limit: 20},
			want:   mrstorage.IDCursor{Limit: 20},
		},
		{
			name:   "negative value is the first page",
			params: mrtype.CursorParams{Value: "-1", Limit: 20},
			want:   mrstorage.IDCursor{Limit: 20},
		},
		{
			name:   "max bigint",
			params: mrtype.CursorParams{Value: "9223372036854775807", Limit: 20},
			want:   mrstorage.IDCursor{ID: math.MaxInt64, Limit: 20},
		},
		{
			name:   "value above bigint range is the first page",
			params: mrtype.CursorParams{Value: "9223372036854775808", Limit: 20},
			want:   mrstorage.IDCursor{Limit: 20},
		},
		{
			name:   "zero limit",
			params: mrtype.CursorParams{Value: "7", Limit: 0},
			want:   mrstorage.IDCursor{ID: 7, Limit: 1},
		},
		{
			name:   "negative limit",
			params: mrtype.CursorParams{Value: "7", Limit: -5},
			want:   mrstorage.IDCursor{ID: 7, Limit: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, mrstorage.NewIDCursor(tt.params))
		})
	}
}

func TestIDCursor_Bounds(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		cursor     mrstorage.IDCursor
		wantAfter  int64
		wantBefore int64
	}

	tests := []testCase{
		{
			// нулевое значение - первая страница в обоих направлениях
			name:       "first page",
			cursor:     mrstorage.IDCursor{},
			wantAfter:  0,
			wantBefore: math.MaxInt64,
		},
		{
			name:       "record id",
			cursor:     mrstorage.IDCursor{ID: 100, Limit: 10},
			wantAfter:  100,
			wantBefore: 100,
		},
		{
			// для DESC граница совпадает с первой страницей, но id < MaxInt64 корректно
			name:       "max bigint",
			cursor:     mrstorage.IDCursor{ID: math.MaxInt64, Limit: 10},
			wantAfter:  math.MaxInt64,
			wantBefore: math.MaxInt64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantAfter, tt.cursor.AfterID())
			assert.Equal(t, tt.wantBefore, tt.cursor.BeforeID())
		})
	}
}
