package xstrings_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xstrings"
)

const testMaxLen = 64

func TestSanitizePrintable(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		value  string
		maxLen int
		want   string
	}

	tests := []testCase{
		{name: "empty", value: "", maxLen: testMaxLen, want: ""},
		{name: "regular", value: "Mozilla/5.0 (X11; Linux x86_64) Firefox/131.0", maxLen: testMaxLen, want: "Mozilla/5.0 (X11; Linux x86_64) Firefox/131.0"},
		{name: "multibyte", value: "Браузер/1.0", maxLen: testMaxLen, want: "Браузер/1.0"},
		{name: "invalid utf8", value: "Mozilla\xff\xfe/5.0", maxLen: testMaxLen, want: "Mozilla/5.0"},
		{name: "control chars", value: "Mozilla\x00/5.0\r\n\tX", maxLen: testMaxLen, want: "Mozilla/5.0 X"},
		{name: "c1 control", value: "Mozilla\u0080/5.0", maxLen: testMaxLen, want: "Mozilla/5.0"},
		{name: "bidi and zero width", value: "Mozilla\u202e/5.0\u200b", maxLen: testMaxLen, want: "Mozilla/5.0"},
		{name: "outer spaces", value: "  Mozilla/5.0  ", maxLen: testMaxLen, want: "Mozilla/5.0"},
		{name: "spaces exposed by removal", value: "\x01 Mozilla/5.0 \x01", maxLen: testMaxLen, want: "Mozilla/5.0"},
		{name: "only garbage", value: "\x00\x01\xff\u200b", maxLen: testMaxLen, want: ""},
		{name: "exact max", value: strings.Repeat("a", testMaxLen), maxLen: testMaxLen, want: strings.Repeat("a", testMaxLen)},
		{name: "long ascii", value: strings.Repeat("a", testMaxLen+100), maxLen: testMaxLen, want: strings.Repeat("a", testMaxLen)},
		{name: "long multibyte", value: strings.Repeat("я", testMaxLen+100), maxLen: testMaxLen, want: strings.Repeat("я", testMaxLen)},
		{
			name:   "removed chars do not count",
			value:  strings.Repeat("\x01", 100) + strings.Repeat("a", testMaxLen+1),
			maxLen: testMaxLen,
			want:   strings.Repeat("a", testMaxLen),
		},
		{name: "exposed leading spaces do not count", value: "\x01   abc", maxLen: 3, want: "abc"},
		{name: "cut on space", value: "ab cd", maxLen: 3, want: "ab"},
		{name: "max one", value: "\x00 я!", maxLen: 1, want: "я"},
		{name: "whitespace to space", value: "a\tb\u00a0c\u3000d\u0085e", maxLen: testMaxLen, want: "a b c d e"},
		{name: "collapse repeats", value: "a  \t\r\n  b", maxLen: testMaxLen, want: "a b"},
		{name: "collapse across removed", value: "a \x01\u200b b", maxLen: testMaxLen, want: "a b"},
		{name: "encoded replacement char kept", value: "a\uFFFDb\xffc", maxLen: testMaxLen, want: "a\uFFFDbc"},
		{name: "repeats do not count", value: "a" + strings.Repeat(" ", 100) + "bc", maxLen: 3, want: "a b"},
		{name: "leading combining marks", value: "\u0301\u20ddabc", maxLen: 3, want: "abc"},
		{name: "inner combining mark kept", value: "e\u0301", maxLen: testMaxLen, want: "e\u0301"},
		{name: "combining mark after space kept", value: "a \u0301b", maxLen: testMaxLen, want: "a \u0301b"},
		{name: "leading spacing mark kept", value: "\u0903abc", maxLen: testMaxLen, want: "\u0903abc"},
		{name: "trailing space input", value: "ab ", maxLen: 3, want: "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := xstrings.SanitizePrintable(tt.value, tt.maxLen)

			assert.Equal(t, tt.want, got)
			assert.True(t, utf8.ValidString(got))
			assert.LessOrEqual(t, utf8.RuneCountInString(got), tt.maxLen)
		})
	}
}

// TestSanitizePrintable_NoLimit - при maxLen < 1 длина не ограничивается, очистка выполняется.
func TestSanitizePrintable_NoLimit(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 10000)

	assert.Equal(t, long, xstrings.SanitizePrintable(long+"\x00", 0))
	assert.Equal(t, long, xstrings.SanitizePrintable("\xff"+long, -1))
	assert.Equal(t, "a b", xstrings.SanitizePrintable(" \t a \x01\r\n b \n", 0))
}
