package mrlog

import (
	stdlog "log"
)

// Fatal - замена log.Fatal: если первые аргументы - Logger и строка-сообщение,
// логирует через FatalError (остальные аргументы - атрибуты), иначе вызывает log.Fatal(v...).
func Fatal(v ...any) {
	if l, msg, ok := extractLogger(v); ok {
		FatalError(l, msg, v[2:]...) // stop if success
	}

	stdlog.Fatal(v...)
}

// extractLogger - извлекает Logger и сообщение из слайса аргументов.
// Ожидает, что первый аргумент - Logger, а второй - строка.
func extractLogger(v []any) (l Logger, msg string, ok bool) {
	if len(v) < 2 {
		return nil, "", false
	}

	if l, ok = v[0].(Logger); !ok {
		return nil, "", false
	}

	if msg, ok = v[1].(string); !ok {
		return nil, "", false
	}

	return l, msg, true
}
