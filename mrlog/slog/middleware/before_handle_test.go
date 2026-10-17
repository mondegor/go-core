package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog/slog/middleware"
)

type (
	// failHandler - обработчик, всегда возвращающий ошибку.
	failHandler struct {
		slog.Handler
	}
)

var errHandle = errors.New("handle failed")

func (failHandler) Handle(context.Context, slog.Record) error {
	return errHandle
}

func newHandler(buf *bytes.Buffer) slog.Handler {
	next := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})

	return middleware.BeforeHandle(
		func(_ context.Context, record slog.Record) slog.Record {
			record.Message = "modified: " + record.Message
			record.AddAttrs(slog.String("added", "yes"))

			return record
		},
	)(next)
}

func TestBeforeHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := newHandler(&buf)

	assert.False(t, handler.Enabled(t.Context(), slog.LevelDebug))
	assert.True(t, handler.Enabled(t.Context(), slog.LevelInfo))

	slog.New(handler).Info("msg")

	var record map[string]any

	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	assert.Equal(t, "modified: msg", record["msg"])
	assert.Equal(t, "yes", record["added"])
}

func TestBeforeHandle_WithAttrsAndGroup(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	// производные обработчики сохраняют модификацию записи
	slog.New(newHandler(&buf)).With("attr", 1).WithGroup("group").Info("msg", "k", "v")

	var record map[string]any

	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	assert.Equal(t, "modified: msg", record["msg"])
	assert.InDelta(t, 1.0, record["attr"], 0)
	assert.Equal(t, map[string]any{"k": "v", "added": "yes"}, record["group"])
}

func TestBeforeHandle_HandleError(t *testing.T) {
	t.Parallel()

	handler := middleware.BeforeHandle(
		func(_ context.Context, record slog.Record) slog.Record {
			return record
		},
	)(failHandler{Handler: slog.DiscardHandler})

	err := handler.Handle(t.Context(), slog.Record{})
	require.ErrorIs(t, err, errHandle)
}
