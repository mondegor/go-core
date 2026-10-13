package xio

import (
	"context"
	"io"
)

type (
	logger interface {
		Debug(ctx context.Context, msg string, args ...any)
		Error(ctx context.Context, msg string, args ...any)
	}
)

// Write - записывает bytes в w, ошибка записи логируется. Пустые данные не пишутся.
func Write(ctx context.Context, logger logger, w io.Writer, bytes []byte) {
	if len(bytes) == 0 {
		return
	}

	if _, err := w.Write(bytes); err != nil {
		logger.Error(ctx, "write failed", "error", err)
	}
}
