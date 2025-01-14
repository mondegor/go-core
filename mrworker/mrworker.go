package mrworker

import (
	"context"
	"time"
)

type (
	// Job - некоторая работа, которая должна быть выполнена.
	Job interface {
		Do(ctx context.Context) error
	}

	// Task - задача исполняемая планировщиком задач.
	Task interface {
		Caption() string
		Startup() bool
		Period() time.Duration
		Timeout() time.Duration
		Job
	}

	// MessageConsumer - получатель сообщений с возможностью подтверждения их получения.
	MessageConsumer interface {
		ReadMessages(ctx context.Context, limit uint32) (messages []any, err error)
		CancelMessages(ctx context.Context, messages []any) error
		CommitMessage(ctx context.Context, message any, preCommit func(ctx context.Context) error) error
		RejectMessage(ctx context.Context, message any, causeErr error) error
	}

	// MessageHandler - обработчик сообщений с закреплением результата.
	MessageHandler interface {
		Execute(ctx context.Context, message any) (commit func(ctx context.Context) error, err error)
	}

	// JobFunc - обёртка для возможности представления анонимной функции в виде Job интерфейса.
	JobFunc func(ctx context.Context) error
)

// Do - запускает анонимную функцию в Job интерфейсе.
func (f JobFunc) Do(ctx context.Context) error {
	return f(ctx)
}
