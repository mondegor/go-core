package process_test

import (
	"context"
	"sync"

	"github.com/mondegor/go-core/mrtrace"
	mrcontext "github.com/mondegor/go-core/mrtrace/context"
	"github.com/mondegor/go-core/mrtrace/process"
)

type (
	// warnLogger - фейковый логгер, запоминающий предупреждения.
	warnLogger struct {
		mu       sync.Mutex
		messages []string
	}
)

func (l *warnLogger) Warn(_ context.Context, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// в тестах предупреждения пишутся с одним атрибутом "key"
	if len(args) == 2 && args[0] == "key" {
		msg = msg + ": " + args[1].(string)
	}

	l.messages = append(l.messages, msg)
}

func (l *warnLogger) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]string(nil), l.messages...)
}

func keyGetIDWithID() []process.KeyGetIDWithID {
	return []process.KeyGetIDWithID{
		{Key: mrtrace.KeyCorrelationID, GetID: mrcontext.CorrelationID, WithID: mrcontext.WithCorrelationID},
		{Key: mrtrace.KeyRequestID, GetID: mrcontext.RequestID, WithID: mrcontext.WithRequestID},
		{Key: mrtrace.KeyProcessID, GetID: mrcontext.ProcessID, WithID: mrcontext.WithProcessID},
	}
}
