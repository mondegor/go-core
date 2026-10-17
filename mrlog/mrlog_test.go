package mrlog_test

import (
	"context"
	"sync"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/level"
)

type (
	logRecord struct {
		level string
		msg   string
		args  []any
	}

	// basicLogger - фейковый логгер, реализующий только интерфейс mrlog.Logger.
	basicLogger struct {
		mu      sync.Mutex
		records []logRecord
	}

	// fullLogger - фейковый логгер с необязательными методами WithAttrs, Enabled и Log.
	fullLogger struct {
		basicLogger
		minLevel level.Enum
		attrs    []any
	}
)

func (l *basicLogger) Debug(_ context.Context, msg string, args ...any) {
	l.add("debug", msg, args)
}

func (l *basicLogger) DebugFunc(_ context.Context, createMsg func() string, args ...any) {
	l.add("debug", createMsg(), args)
}

func (l *basicLogger) Info(_ context.Context, msg string, args ...any) {
	l.add("info", msg, args)
}

func (l *basicLogger) Warn(_ context.Context, msg string, args ...any) {
	l.add("warn", msg, args)
}

func (l *basicLogger) Error(_ context.Context, msg string, args ...any) {
	l.add("error", msg, args)
}

func (l *basicLogger) add(lvl, msg string, args []any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.records = append(l.records, logRecord{level: lvl, msg: msg, args: args})
}

func (l *basicLogger) all() []logRecord {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]logRecord(nil), l.records...)
}

func (l *fullLogger) WithAttrs(attrs ...any) mrlog.Logger {
	return &fullLogger{minLevel: l.minLevel, attrs: append(append([]any(nil), l.attrs...), attrs...)}
}

func (l *fullLogger) Enabled(_ context.Context, lvl level.Enum) bool {
	return lvl >= l.minLevel
}

func (l *fullLogger) Log(_ context.Context, lvl level.Enum, msg string, args ...any) {
	l.add(lvl.String(), msg, args)
}
