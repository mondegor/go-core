package color_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrlog/color"
)

func TestColorizeText(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "\033[91mtext\033[0m", color.ColorizeText(color.Red, "text"))
	assert.Equal(t, color.End, color.ColorizeText("", ""))
}
