package xio_test

import (
	"context"
	"sync"
)

type (
	logRecord struct {
		level string
		msg   string
		args  []any
	}

	// recordLogger - фейковый логгер, запоминающий все записи.
	recordLogger struct {
		mu      sync.Mutex
		records []logRecord
	}
)

func (l *recordLogger) Debug(_ context.Context, msg string, args ...any) {
	l.add("debug", msg, args)
}

func (l *recordLogger) Error(_ context.Context, msg string, args ...any) {
	l.add("error", msg, args)
}

func (l *recordLogger) add(level, msg string, args []any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.records = append(l.records, logRecord{level: level, msg: msg, args: args})
}

func (l *recordLogger) levels() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	levels := make([]string, 0, len(l.records))

	for _, r := range l.records {
		levels = append(levels, r.level)
	}

	return levels
}
