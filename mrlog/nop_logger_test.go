package mrlog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrlog"
)

func TestNopLogger(t *testing.T) {
	t.Parallel()

	logger := mrlog.NopLogger()

	assert.NotPanics(t, func() {
		logger.Debug(t.Context(), "msg")
		logger.DebugFunc(t.Context(), func() string {
			t.Error("createMsg must not be called")

			return ""
		})
		logger.Info(t.Context(), "msg")
		logger.Warn(t.Context(), "msg")
		logger.Error(t.Context(), "msg")
	})
}
