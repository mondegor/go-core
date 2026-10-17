package middleware

import (
	"context"
	"log/slog"
)

// BeforeHandle - создаёт middleware, передающий в следующий slog.Handler запись,
// возвращённую fn. Производные обработчики (WithAttrs, WithGroup) сохраняют middleware.
// Атрибуты, добавленные fn, попадают в текущую группу обработчика (после WithGroup - внутрь неё),
// как и собственные атрибуты записи в slog.
func BeforeHandle(fn func(ctx context.Context, record slog.Record) slog.Record) func(next slog.Handler) slog.Handler {
	return func(next slog.Handler) slog.Handler {
		return &decorator{
			handler:      next,
			beforeHandle: fn,
		}
	}
}
