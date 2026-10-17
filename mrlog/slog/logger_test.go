package slog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	stdlog "log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/color"
	"github.com/mondegor/go-core/mrlog/level"
	"github.com/mondegor/go-core/mrlog/slog"
)

func TestNewLoggerAdapterErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		opt  slog.Option
	}

	tests := []testCase{
		{name: "invalid level", opt: slog.WithLevel("FATAL")},
		{name: "invalid time format", opt: slog.WithTimeFormat("ISO")},
		{name: "empty time zone", opt: slog.WithTimeZone("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := slog.NewLoggerAdapter(slog.WithWriter(&bytes.Buffer{}), tt.opt)
			require.Error(t, err)
			assert.Nil(t, logger)
		})
	}
}

func TestLoggerAdapter_Levels(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newJSONLogger(t, &buf, slog.WithLevel("DEBUG"))

	logger.Debug(t.Context(), "debug msg", "k", 1)
	logger.DebugFunc(t.Context(), func() string { return "debug func msg" })
	logger.Info(t.Context(), "info msg")
	logger.Warn(t.Context(), "warn msg")
	logger.Error(t.Context(), "error msg")
	logger.Log(t.Context(), level.Fatal, "fatal msg")
	logger.Trace(t.Context(), "k", 2)

	records := decodeRecords(t, &buf)
	require.Len(t, records, 7)

	type want struct{ level, msg string }

	got := make([]want, 0, len(records))
	for _, r := range records {
		got = append(got, want{level: r["level"].(string), msg: r["msg"].(string)})
	}

	assert.Equal(
		t,
		[]want{
			{level: "DEBUG", msg: "debug msg"},
			{level: "DEBUG", msg: "debug func msg"},
			{level: "INFO", msg: "info msg"},
			{level: "WARN", msg: "warn msg"},
			{level: "ERROR", msg: "error msg"},
			{level: "FATAL", msg: "fatal msg"},
			{level: "TRACE", msg: "trace"},
		},
		got,
	)
	assert.InDelta(t, 1.0, records[0]["k"], 0)
	assert.InDelta(t, 2.0, records[6]["k"], 0)
}

func TestLoggerAdapter_LevelFilter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newJSONLogger(t, &buf, slog.WithLevel("WARN"))

	assert.False(t, logger.Enabled(t.Context(), level.Debug))
	assert.False(t, logger.Enabled(t.Context(), level.Info))
	assert.True(t, logger.Enabled(t.Context(), level.Warn))
	assert.True(t, mrlog.Enabled(logger, level.Error))

	logger.Debug(t.Context(), "debug msg")
	logger.DebugFunc(t.Context(), func() string {
		t.Error("createMsg must not be called for disabled level")

		return ""
	})
	logger.Info(t.Context(), "info msg")
	logger.Warn(t.Context(), "warn msg")

	records := decodeRecords(t, &buf)
	require.Len(t, records, 1)
	assert.Equal(t, "warn msg", records[0]["msg"])
}

func TestLoggerAdapter_EnabledUsesContext(t *testing.T) {
	t.Parallel()

	type debugKey struct{}

	// обработчик включает level.Debug только для контекста с debugKey
	logger := slog.New(
		handlerEnabledFunc{
			Handler: stdlog.DiscardHandler,
			enabled: func(ctx context.Context, lvl stdlog.Level) bool {
				return lvl >= stdlog.LevelInfo || ctx.Value(debugKey{}) != nil
			},
		},
	)

	assert.True(t, logger.Enabled(context.WithValue(t.Context(), debugKey{}, true), level.Debug))
	assert.False(t, logger.Enabled(t.Context(), level.Debug))
	assert.False(t, mrlog.DebugEnabled(logger))
	assert.True(t, mrlog.InfoEnabled(logger))
}

func TestLoggerAdapter_DebugFuncUsesCallContext(t *testing.T) {
	t.Parallel()

	type debugKey struct{}

	var buf bytes.Buffer

	// обработчик включает level.Debug только для контекста с debugKey
	handler := stdlog.NewJSONHandler(&buf, &stdlog.HandlerOptions{Level: stdlog.LevelDebug})
	logger := slog.New(
		handlerEnabledFunc{
			Handler: handler,
			enabled: func(ctx context.Context, lvl stdlog.Level) bool {
				return lvl >= stdlog.LevelInfo || ctx.Value(debugKey{}) != nil
			},
		},
	)

	logger.DebugFunc(context.WithValue(t.Context(), debugKey{}, true), func() string { return "debug func msg" })
	logger.DebugFunc(t.Context(), func() string {
		t.Error("createMsg must not be called for disabled level")

		return ""
	})

	records := decodeRecords(t, &buf)
	require.Len(t, records, 1)
	assert.Equal(t, "debug func msg", records[0]["msg"])
}

func TestLoggerAdapter_WithAttributes(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newJSONLogger(t, &buf)

	assert.Same(t, logger, logger.WithAttributes())

	child := logger.WithAttributes("k1", "v1")
	assert.NotSame(t, logger, child)

	childLogger := mrlog.WithAttrs(child, "k2", "v2")
	childLogger.Info(t.Context(), "child msg")
	logger.Info(t.Context(), "parent msg")

	records := decodeRecords(t, &buf)
	require.Len(t, records, 2)
	assert.Equal(t, "v1", records[0]["k1"])
	assert.Equal(t, "v2", records[0]["k2"])
	assert.NotContains(t, records[1], "k1")
}

func TestLoggerAdapter_TextFormatLevels(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(slog.WithWriter(&buf), slog.WithColorMode(false))
	require.NoError(t, err)

	logger.Debug(t.Context(), "msg")
	logger.Info(t.Context(), "msg")
	logger.Log(t.Context(), level.Fatal, "msg")
	logger.Trace(t.Context())

	out := buf.String()
	assert.Contains(t, out, "level=DEBUG")
	assert.Contains(t, out, "level=INFO")
	assert.Contains(t, out, "level=FATAL")
	assert.Contains(t, out, "level=TRACE")
}

func TestLoggerAdapter_ColoredLevels(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(slog.WithWriter(&buf))
	require.NoError(t, err)

	logger.Debug(t.Context(), "msg")
	logger.Log(t.Context(), level.Fatal, "msg")
	logger.Trace(t.Context())

	out := buf.String()
	assert.Contains(t, out, color.ColorizeText(color.Yellow, "DBG"))
	assert.Contains(t, out, color.ColorizeText(color.Red, "FAT"))
	assert.Contains(t, out, color.ColorizeText(color.Blue, "TRC"))
}

func TestLoggerAdapter_ColoredAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(
		slog.WithWriter(&buf),
		slog.WithColorizeAttr(slog.AttrColorByDefaultKey, color.Green, color.Magenta),
		slog.WithColorizeAttr("keyOnly", color.Red, ""),
	)
	require.NoError(t, err)

	logger.Info(t.Context(), "msg", "plain", "v1", "keyOnly", "v2", "err", errors.New("boom"))

	out := buf.String()
	assert.Contains(t, out, color.ColorizeText(color.Green, "plain"))
	assert.Contains(t, out, color.ColorizeText(color.Magenta, "v1"))
	assert.Contains(t, out, color.ColorizeText(color.Red, "keyOnly"))
	assert.Contains(t, out, color.End+"v2") // значение без цвета идёт сразу за сбросом цвета ключа
	assert.Contains(t, out, "boom")
	assert.NotContains(t, out, color.ColorizeText(color.Green, "err"))
}

func TestLoggerAdapter_ReplaceAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newJSONLogger(
		t,
		&buf,
		slog.WithReplaceAttrs(func(attr stdlog.Attr) stdlog.Attr {
			if attr.Key == "secret" {
				return stdlog.String(attr.Key, "***")
			}

			return attr
		}),
	)

	logger.Info(t.Context(), "msg", "secret", "password")

	records := decodeRecords(t, &buf)
	require.Len(t, records, 1)
	assert.Equal(t, "***", records[0]["secret"])
}

func TestLoggerAdapter_MiddlewareOrder(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	// middleware дописывает своё имя к сообщению перед передачей записи дальше
	appendName := func(name string) func(next stdlog.Handler) stdlog.Handler {
		return func(next stdlog.Handler) stdlog.Handler {
			return handlerFunc{
				Handler: next,
				handle: func(ctx context.Context, r stdlog.Record) error {
					r.Message += "|" + name

					return next.Handle(ctx, r)
				},
			}
		}
	}

	logger := newJSONLogger(
		t,
		&buf,
		slog.WithMiddlewareHandler(appendName("first"), appendName("second")),
		slog.WithMiddlewareHandler(appendName("third")),
	)

	logger.Info(t.Context(), "msg")

	records := decodeRecords(t, &buf)
	require.Len(t, records, 1)
	assert.Equal(t, "msg|first|second|third", records[0]["msg"])
}

type (
	// handlerFunc - slog.Handler с подменённым методом Handle.
	handlerFunc struct {
		stdlog.Handler
		handle func(ctx context.Context, r stdlog.Record) error
	}

	// handlerEnabledFunc - slog.Handler с подменённым методом Enabled.
	handlerEnabledFunc struct {
		stdlog.Handler
		enabled func(ctx context.Context, lvl stdlog.Level) bool
	}
)

func (h handlerFunc) Handle(ctx context.Context, r stdlog.Record) error {
	return h.handle(ctx, r)
}

func (h handlerEnabledFunc) Enabled(ctx context.Context, lvl stdlog.Level) bool {
	return h.enabled(ctx, lvl)
}

func newJSONLogger(t *testing.T, buf *bytes.Buffer, opts ...slog.Option) *slog.LoggerAdapter {
	t.Helper()

	logger, err := slog.NewLoggerAdapter(append([]slog.Option{slog.WithWriter(buf), slog.WithJsonFormat(true)}, opts...)...)
	require.NoError(t, err)

	return logger
}

func decodeRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var records []map[string]any

	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}

		var record map[string]any

		require.NoError(t, json.Unmarshal([]byte(line), &record))

		records = append(records, record)
	}

	return records
}
