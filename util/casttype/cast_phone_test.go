package casttype_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/casttype"
)

func TestUintToPhone(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "+79001234567", casttype.UintToPhone(79001234567))
	assert.Empty(t, casttype.UintToPhone(0))
}
