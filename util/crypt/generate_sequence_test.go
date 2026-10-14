package crypt_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/crypt"
)

func TestGenerateSequence(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		charset string
		length  int
	}

	tests := []testCase{
		{name: "one char", charset: "a", length: 10},
		{name: "power of two", charset: "01234567", length: 100},
		{name: "not power of two", charset: "abc", length: 100},
		{name: "max charset", charset: maxCharset(), length: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := crypt.GenerateSequence([]byte(tt.charset), tt.length)
			require.NoError(t, err)
			assert.Len(t, got, tt.length)
			assertCharset(t, tt.charset, got)
		})
	}
}

func TestGenerateSequenceErrors(t *testing.T) {
	t.Parallel()

	_, err := crypt.GenerateSequence(nil, 1)
	require.Error(t, err)

	_, err = crypt.GenerateSequence([]byte(strings.Repeat("x", 129)), 1)
	require.Error(t, err)

	_, err = crypt.GenerateSequence([]byte("abc"), 0)
	require.Error(t, err)
}

func TestGenerateBytes(t *testing.T) {
	t.Parallel()

	got, err := crypt.GenerateBytes([]byte("ab"), 5)
	require.NoError(t, err)
	assert.Len(t, got, 5)
	assert.Equal(t, 5, cap(got))
	assertCharset(t, "ab", string(got))
}

func TestGenerateDigitsHexToken(t *testing.T) {
	t.Parallel()

	digits, err := crypt.GenerateDigits(32)
	require.NoError(t, err)
	assert.Len(t, digits, 32)
	assertCharset(t, "0123456789", digits)

	hex, err := crypt.GenerateHex(32)
	require.NoError(t, err)
	assert.Len(t, hex, 32)
	assertCharset(t, "0123456789abcdef", hex)

	token, err := crypt.GenerateToken(32)
	require.NoError(t, err)
	assert.Len(t, token, 32)
	assertCharset(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", token)
}

// maxCharset - возвращает набор из 128 различных символов (все ASCII).
func maxCharset() string {
	charset := make([]byte, 128)

	for i := range charset {
		charset[i] = byte(i)
	}

	return string(charset)
}

func assertCharset(t *testing.T, charset, got string) {
	t.Helper()

	for _, c := range got {
		assert.Contains(t, charset, string(c))
	}
}
