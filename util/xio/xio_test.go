package xio_test

import (
	"context"
)

type (
	// recordLogger - фейковый логгер, запоминающий уровни всех записей.
	recordLogger struct {
		levels []string
	}
)

func (l *recordLogger) Debug(context.Context, string, ...any) {
	l.levels = append(l.levels, "debug")
}

func (l *recordLogger) Error(context.Context, string, ...any) {
	l.levels = append(l.levels, "error")
}
