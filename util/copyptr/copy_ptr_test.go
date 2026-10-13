package copyptr_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/copyptr"
)

func TestTime(t *testing.T) {
	t.Parallel()

	assert.Nil(t, copyptr.Time(nil))
	assert.Nil(t, copyptr.Time(&time.Time{}))

	value := time.Now()

	got := copyptr.Time(&value)
	require.NotNil(t, got)
	assert.Equal(t, value, *got)
	assert.NotSame(t, &value, got)
}
