package level

import (
	"fmt"
	"log/slog"
)

type (
	// Enum - уровень логирования.
	Enum int
)

// Перечисление уровней логирования.
const (
	Debug = Enum(slog.LevelDebug)     // Error + Warn + Info + отладочные сообщения
	Info  = Enum(slog.LevelInfo)      // Error + Warn + информационные сообщения
	Warn  = Enum(slog.LevelWarn)      // Error + предупреждения
	Error = Enum(slog.LevelError)     // ошибки (максимальный уровень, который можно назначить в конфиге)
	Fatal = Enum(slog.LevelError + 4) // дополнительный уровень для фатальных ошибок (служебный)
	Trace = Enum(slog.LevelError + 8) // режим трассировки (служебный, управляется отдельно)
)

// String - возвращает значение уровня логирования в виде строки.
func (e Enum) String() string {
	switch e {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	case Trace:
		return "TRACE"
	}

	return "UNKNOWN"
}

// Parse - преобразует строку в уровень логирования.
// Возвращает найденный уровень при успехе, или ошибку при нераспознанном значении.
func Parse(value string) (Enum, error) {
	switch value {
	case "DEBUG":
		return Debug, nil
	case "INFO":
		return Info, nil
	case "WARN":
		return Warn, nil
	case "ERROR":
		return Error, nil
	}

	return Error, fmt.Errorf("the value is not a logging level (value='%s')", value)
}
