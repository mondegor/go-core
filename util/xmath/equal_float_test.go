package xmath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xmath"
)

func TestEqualFloat(t *testing.T) {
	t.Parallel()

	assert.True(t, xmath.EqualFloat(1.0, 1.5, 0.5))
	assert.False(t, xmath.EqualFloat(1.0, 1.6, 0.5))

	assert.True(t, xmath.EqualFloatE3(1.0, 1.0009))
	assert.False(t, xmath.EqualFloatE3(1.0, 1.002))

	assert.True(t, xmath.EqualFloatE6(1.0, 1.0000009))
	assert.False(t, xmath.EqualFloatE6(1.0, 1.000002))

	assert.True(t, xmath.EqualFloatE9(-1.0, -1.0000000009))
	assert.False(t, xmath.EqualFloatE9(-1.0, -1.000000002))
}
