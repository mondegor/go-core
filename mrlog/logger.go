package mrlog

import (
	"context"
	"os"

	"github.com/mondegor/go-core/mrlog/level"
)

type (
	// Logger - интерфейс логирования сообщений и ошибок с использованием контекста.
	// Функции пакета не проверяют логгер на nil (в том числе typed-nil), как и log/slog:
	// nil - ошибка вызывающего. Если логирование не нужно, передаётся NopLogger()
	// (аналог slog.DiscardHandler).
	// Дополнительные методы (WithAttrs, Enabled(ctx, lvl), Log) определяются через type assertion,
	// как опциональные интерфейсы стандартной библиотеки (io.WriterTo, http.Flusher); если логгер их не реализует,
	// функции пакета используют описанное у них запасное поведение.
	// Значения args методов логирования форматируются реализацией только при фактической записи
	// сообщения (после проверки уровня), поэтому дорогие вычисления можно отложить, передав значение,
	// которое вычисляется при форматировании: fmt.Stringer, а для структурных форматов
	// также encoding.TextMarshaler. Это не относится к атрибутам WithAttrs: реализация может
	// форматировать их сразу (так делают TextHandler и JSONHandler из log/slog).
	Logger interface {
		Debug(ctx context.Context, msg string, args ...any)
		DebugFunc(ctx context.Context, createMsg func() string, args ...any)
		Info(ctx context.Context, msg string, args ...any)
		Warn(ctx context.Context, msg string, args ...any)
		Error(ctx context.Context, msg string, args ...any)
	}
)

// WithAttrs - возвращает логгер с прикреплёнными к нему атрибутами.
// Если логгер не поддерживает WithAttrs, пишет предупреждение и возвращает исходный логгер.
func WithAttrs(l Logger, attrs ...any) Logger {
	if l, ok := l.(interface{ WithAttrs(attrs ...any) Logger }); ok {
		return l.WithAttrs(attrs...)
	}

	Warn(l, "Logger not supported method 'WithAttrs'")

	return l
}

// DebugEnabled - сообщает, логирует ли указанный логгер сообщения уровня level.Debug.
func DebugEnabled(l Logger) bool {
	return Enabled(l, level.Debug)
}

// InfoEnabled - сообщает, логирует ли указанный логгер сообщения уровня level.Info.
func InfoEnabled(l Logger) bool {
	return Enabled(l, level.Info)
}

// Enabled - сообщает, включён ли указанный уровень логирования.
// Проверяет без контекста, предназначена для этапа инициализации;
// в рабочем коде уровень учитывают методы логгера с ctx (например, DebugFunc).
// Если логгер не поддерживает Enabled, пишет предупреждение и возвращает false.
func Enabled(l Logger, lvl level.Enum) bool {
	if ll, ok := l.(interface {
		Enabled(ctx context.Context, lvl level.Enum) bool
	}); ok {
		return ll.Enabled(context.Background(), lvl)
	}

	Warn(l, "Logger not supported method 'Enabled'")

	return false
}

// Debug - логирует сообщения на уровне level.Debug без использования контекста.
func Debug(l Logger, msg string, args ...any) {
	l.Debug(context.Background(), msg, args...)
}

// DebugFunc - логирует сообщения на уровне level.Debug без использования контекста с отложенным созданием сообщения.
func DebugFunc(l Logger, createMsg func() string, args ...any) {
	l.DebugFunc(context.Background(), createMsg, args...)
}

// Info - логирует сообщения на уровне level.Info без использования контекста.
func Info(l Logger, msg string, args ...any) {
	l.Info(context.Background(), msg, args...)
}

// Warn - логирует сообщения на уровне level.Warn без использования контекста.
func Warn(l Logger, msg string, args ...any) {
	l.Warn(context.Background(), msg, args...)
}

// Error - логирует сообщения на уровне level.Error без использования контекста.
func Error(l Logger, msg string, args ...any) {
	l.Error(context.Background(), msg, args...)
}

// FatalError - логирует сообщение на уровне level.Fatal без использования контекста и завершает процесс с кодом 1.
// Если логгер не поддерживает Log, сообщение пишется через Error с префиксом "Fatal error: ".
func FatalError(l Logger, msg string, args ...any) {
	if ll, ok := l.(interface {
		Log(ctx context.Context, lvl level.Enum, message string, args ...any)
	}); ok {
		ll.Log(context.Background(), level.Fatal, msg, args...)
	} else {
		l.Error(context.Background(), "Fatal error: "+msg, args...)
	}

	os.Exit(1) //nolint:revive
}
