package slog

import (
	"context"
	"fmt"
	stdlog "log/slog"
	"os"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/color"
	"github.com/mondegor/go-core/mrlog/level"
)

type (
	// LoggerAdapter - реализация mrlog.Logger на основе slog.Logger.
	// Поддерживает цветной, текстовый и JSON-вывод, а также middleware.
	// Значения args передаются в slog.Logger без изменений, поэтому для отложенных
	// вычислений, помимо fmt.Stringer и encoding.TextMarshaler, поддерживается slog.LogValuer
	// (раскрывается до ReplaceAttr). Цветной вывод форматирует значения через fmt.
	LoggerAdapter struct {
		sl *stdlog.Logger
	}

	// attrColor - цветовая схема для атрибута (ключ/значение).
	attrColor struct {
		key   string
		value string
	}
)

// NewLoggerAdapter - создаёт LoggerAdapter с заданными параметрами.
// По умолчанию: уровень DEBUG, цветной вывод в os.Stdout, время в RFC3339 и UTC.
// При включённых colorMode и jsonFormat приоритет у JSON.
func NewLoggerAdapter(opts ...Option) (logger *LoggerAdapter, err error) {
	o := options{
		levelString: level.Debug.String(),
		jsonFormat:  false,
		timeFormat:  "RFC3339",
		timeZone:    "UTC",
		colorMode:   true,
		attrKey2color: map[string]attrColor{
			// сообщение записи заметнее значений атрибутов, но остаётся рядовым
			// ключом: перекрывается через WithColorizeAttr как любой другой и,
			// как любой другой, действует по имени без учёта группы - вложенный
			// msg получает тот же цвет (см. colorizeBuiltinAttr)
			stdlog.MessageKey: {key: color.Cyan, value: color.White},
		},
		attrColorByDefault: attrColor{
			key:   color.Cyan,
			value: color.LightGray,
		},
	}

	for _, opt := range opts {
		opt(&o)
	}

	if o.stdout == nil {
		o.stdout = os.Stdout
	}

	o.level, err = level.Parse(o.levelString)
	if err != nil {
		return nil, fmt.Errorf("error parsing level: %w", err)
	}

	o.timeFormat, err = mrlog.ParseDateTimeFormat(o.timeFormat)
	if err != nil {
		return nil, fmt.Errorf("error parsing time: %w", err)
	}

	// часовой пояс процесса задаётся явно через WithTimeZone("Local")
	o.timeLocation, err = mrlog.ParseTimeZone(o.timeZone)
	if err != nil {
		return nil, fmt.Errorf("error parsing timezone: %w", err)
	}

	if o.replaceAttr == nil {
		o.replaceAttr = func(attr stdlog.Attr) (newAttr stdlog.Attr) {
			return attr
		}
	}

	var handler stdlog.Handler

	switch {
	case o.jsonFormat:
		handler = stdlog.NewJSONHandler(o.stdout, handlerOptions(o))
	case o.colorMode:
		handler = newColoredHandler(o.stdout, o)
	default:
		handler = stdlog.NewTextHandler(o.stdout, handlerOptions(o))
	}

	for i := len(o.middlewareHandlers) - 1; i >= 0; i-- {
		handler = o.middlewareHandlers[i](handler)
	}

	return New(handler), nil
}

// New - создаёт LoggerAdapter из готового slog.Handler.
// Позволяет использовать кастомные обработчики slog напрямую.
func New(handler stdlog.Handler) *LoggerAdapter {
	return &LoggerAdapter{
		sl: stdlog.New(handler),
	}
}

// WithAttributes - возвращает новый логгер с прикреплёнными атрибутами.
// Один атрибут состоит из двух аргументов: ключ:string/значение:any.
func (l *LoggerAdapter) WithAttributes(attrs ...any) *LoggerAdapter {
	if len(attrs) == 0 {
		return l
	}

	c := *l
	c.sl = l.sl.With(attrs...)

	return &c
}

// WithAttrs - см. WithAttributes.
func (l *LoggerAdapter) WithAttrs(attrs ...any) mrlog.Logger {
	return l.WithAttributes(attrs...)
}

// Enabled - сообщает, включён ли указанный уровень логирования для ctx (как в slog.Logger).
func (l *LoggerAdapter) Enabled(ctx context.Context, lvl level.Enum) bool {
	return l.sl.Enabled(ctx, stdlog.Level(lvl))
}

// Log - логирует сообщения на указанном уровне.
func (l *LoggerAdapter) Log(ctx context.Context, lvl level.Enum, msg string, args ...any) {
	l.sl.Log(ctx, stdlog.Level(lvl), msg, args...)
}

// Debug - логирует сообщения на уровне level.Debug.
func (l *LoggerAdapter) Debug(ctx context.Context, msg string, args ...any) {
	l.sl.DebugContext(ctx, msg, args...)
}

// DebugFunc - логирует сообщения на уровне level.Debug с их отложенным созданием:
// createMsg вызывается, только если обработчик включает level.Debug для ctx
// (уровень проверяется с контекстом вызова, как в slog.Logger).
func (l *LoggerAdapter) DebugFunc(ctx context.Context, createMsg func() string, args ...any) {
	if !l.Enabled(ctx, level.Debug) {
		return
	}

	l.sl.DebugContext(ctx, createMsg(), args...)
}

// Info - логирует сообщения на уровне level.Info.
func (l *LoggerAdapter) Info(ctx context.Context, msg string, args ...any) {
	l.sl.InfoContext(ctx, msg, args...)
}

// Warn - логирует сообщения на уровне level.Warn.
func (l *LoggerAdapter) Warn(ctx context.Context, msg string, args ...any) {
	l.sl.WarnContext(ctx, msg, args...)
}

// Error - логирует сообщения на уровне level.Error.
func (l *LoggerAdapter) Error(ctx context.Context, msg string, args ...any) {
	l.sl.ErrorContext(ctx, msg, args...)
}

// Trace - логирует запись с сообщением "trace" и атрибутами args на уровне level.Trace.
func (l *LoggerAdapter) Trace(ctx context.Context, args ...any) {
	l.sl.Log(ctx, stdlog.Level(level.Trace), "trace", args...)
}
