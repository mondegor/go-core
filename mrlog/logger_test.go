package mrlog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/level"
)

func TestWithAttrs(t *testing.T) {
	t.Parallel()

	logger := &fullLogger{}

	got := mrlog.WithAttrs(logger, "key", "value")
	require.IsType(t, &fullLogger{}, got)
	assert.Equal(t, []any{"key", "value"}, got.(*fullLogger).attrs)
	assert.Empty(t, logger.attrs)
}

func TestWithAttrsNotSupported(t *testing.T) {
	t.Parallel()

	logger := &basicLogger{}

	got := mrlog.WithAttrs(logger, "key", "value")
	assert.Same(t, logger, got)
	assert.Equal(t, []logRecord{{level: "warn", msg: "Logger not supported method 'WithAttrs'"}}, logger.all())
}

func TestEnabled(t *testing.T) {
	t.Parallel()

	logger := &fullLogger{minLevel: level.Info}

	assert.False(t, mrlog.DebugEnabled(logger))
	assert.True(t, mrlog.InfoEnabled(logger))
	assert.True(t, mrlog.Enabled(logger, level.Error))
	assert.Empty(t, logger.all())
}

func TestEnabledNotSupported(t *testing.T) {
	t.Parallel()

	logger := &basicLogger{}

	assert.False(t, mrlog.DebugEnabled(logger))
	assert.False(t, mrlog.InfoEnabled(logger))
	assert.Equal(
		t,
		[]logRecord{
			{level: "warn", msg: "Logger not supported method 'Enabled'"},
			{level: "warn", msg: "Logger not supported method 'Enabled'"},
		},
		logger.all(),
	)
}

func TestLevelHelpers(t *testing.T) {
	t.Parallel()

	logger := &basicLogger{}

	mrlog.Debug(logger, "debug msg", "k", 1)
	mrlog.DebugFunc(logger, func() string { return "debug func msg" })
	mrlog.Info(logger, "info msg")
	mrlog.Warn(logger, "warn msg")
	mrlog.Error(logger, "error msg")

	assert.Equal(
		t,
		[]logRecord{
			{level: "debug", msg: "debug msg", args: []any{"k", 1}},
			{level: "debug", msg: "debug func msg"},
			{level: "info", msg: "info msg"},
			{level: "warn", msg: "warn msg"},
			{level: "error", msg: "error msg"},
		},
		logger.all(),
	)
}
