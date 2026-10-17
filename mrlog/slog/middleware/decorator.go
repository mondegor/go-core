package middleware

import (
	"context"
	"fmt"
	"log/slog"
)

type (
	// decorator - предназначен для модификации объекта slog.Record
	// непосредственно перед вызовом метода Handle в интерфейсе slog.Handler.
	decorator struct {
		handler      slog.Handler
		beforeHandle func(ctx context.Context, record slog.Record) slog.Record
	}
)

// Enabled - реализация метода для интерфейса slog.Handler.
func (d *decorator) Enabled(ctx context.Context, level slog.Level) bool {
	return d.handler.Enabled(ctx, level)
}

// Handle - реализация метода для интерфейса slog.Handler:
// передаёт в обработчик запись, изменённую beforeHandle.
func (d *decorator) Handle(ctx context.Context, record slog.Record) error {
	if err := d.handler.Handle(ctx, d.beforeHandle(ctx, record)); err != nil {
		return fmt.Errorf("middleware.decorator.Handle error: %w", err)
	}

	return nil
}

// WithAttrs - реализация метода для интерфейса slog.Handler.
func (d *decorator) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := *d
	c.handler = d.handler.WithAttrs(attrs)

	return &c
}

// WithGroup - реализация метода для интерфейса slog.Handler.
func (d *decorator) WithGroup(name string) slog.Handler {
	c := *d
	c.handler = d.handler.WithGroup(name)

	return &c
}
