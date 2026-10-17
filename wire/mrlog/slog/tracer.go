package slog

import (
	"context"

	"github.com/mondegor/go-core/mrlog/level"
	"github.com/mondegor/go-core/mrlog/slog"
	"github.com/mondegor/go-core/mrtrace"
)

// InitTracer - создаёт mrtrace.Tracer на основе slog.LoggerAdapter.
// Если уровень логирования включает Debug, возвращает сам логгер как Tracer.
// В противном случае возвращает заглушку (NopTracer), чтобы избежать накладных расходов.
func InitTracer(logger *slog.LoggerAdapter) mrtrace.Tracer {
	if logger.Enabled(context.Background(), level.Debug) {
		return logger
	}

	return mrtrace.NopTracer()
}
