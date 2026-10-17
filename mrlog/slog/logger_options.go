package slog

import (
	"io"
	stdlog "log/slog"
	"time"

	"github.com/mondegor/go-core/mrlog/level"
)

const (
	// AttrColorByDefaultKey - ключ используемый в WithColorizeAttr
	// для установки цветов ключей и значений по умолчанию.
	AttrColorByDefaultKey = "attrColorByDefault"
)

type (
	// Option - функция-опция для настройки LoggerAdapter.
	Option func(o *options)

	options struct {
		stdout             io.Writer
		levelString        string
		level              level.Enum
		jsonFormat         bool
		timeFormat         string
		timeZone           string
		timeLocation       *time.Location
		middlewareHandlers []func(next stdlog.Handler) stdlog.Handler
		replaceAttr        func(attr stdlog.Attr) (newAttr stdlog.Attr)
		colorMode          bool
		attrKey2color      map[string]attrColor
		attrColorByDefault attrColor
	}
)

// WithWriter - задаёт поток вывода для логирования (по умолчанию os.Stdout).
func WithWriter(value io.Writer) Option {
	return func(o *options) {
		o.stdout = value
	}
}

// WithLevel - задаёт минимальный уровень логирования (DEBUG, INFO, WARN, ERROR).
// Сообщения с более низким уровнем будут отфильтрованы.
func WithLevel(value string) Option {
	return func(o *options) {
		o.levelString = value
	}
}

// WithJsonFormat - включает вывод логов в формате JSON (имеет приоритет над WithColorMode).
func WithJsonFormat(value bool) Option {
	return func(o *options) {
		o.jsonFormat = value
	}
}

// WithTimeFormat - задаёт формат вывода времени в логах.
// Поддерживаемые именованные форматы: "RFC3339", "RFC3339Nano", "DateTime", "TimeOnly", "Kitchen".
func WithTimeFormat(value string) Option {
	return func(o *options) {
		o.timeFormat = value
	}
}

// WithTimeZone - задаёт часовой пояс вывода времени в логах.
// Поддерживаются IANA-имена (например: "Europe/Moscow"), а также "UTC" и "Local".
// Пустое значение считается ошибкой; если опция не указана, то используется UTC
// независимо от режима вывода. Чтобы выводить время в часовом поясе процесса,
// укажите "Local".
//
// ВНИМАНИЕ: для IANA-имён требуется база часовых поясов в системе;
// в минимальных образах (scratch, distroless) её нужно встроить
// в бинарник импортом `_ "time/tzdata"`.
// Значения "UTC" и "Local" доступны всегда.
func WithTimeZone(value string) Option {
	return func(o *options) {
		o.timeZone = value
	}
}

// WithMiddlewareHandler - добавляет один или несколько middleware-обработчиков.
// Запись проходит через middleware в порядке их добавления, затем попадает в основной обработчик.
func WithMiddlewareHandler(value ...func(next stdlog.Handler) stdlog.Handler) Option {
	return func(o *options) {
		o.middlewareHandlers = append(o.middlewareHandlers, value...)
	}
}

// WithReplaceAttrs - задаёт функцию обработки каждого атрибута перед выводом
// (вызывается до форматирования времени и уровня). Нулевой атрибут (slog.Attr{}) удаляется из записи.
func WithReplaceAttrs(value func(attr stdlog.Attr) (newAttr stdlog.Attr)) Option {
	return func(o *options) {
		o.replaceAttr = value
	}
}

// WithColorMode - включает/отключает цветной вывод в консоль.
// Игнорируется, если включен JsonFormat.
func WithColorMode(value bool) Option {
	return func(o *options) {
		o.colorMode = value
	}
}

// WithColorizeAttr - задаёт цвета для ключа/значения указанного атрибута.
// Параметры:
//   - attrKey - имя атрибута, которому назначаются цвета;
//   - colorKey, colorValue - ANSI-коды цветов из пакета color.
//
// При attrKey=AttrColorByDefaultKey цвета назначаются всем атрибутам,
// для которых явно не были определены цвета.
// Работает только при включенном ColorMode и отключенном JsonFormat.
//
// Цвета назначаются по имени атрибута без учёта группы, поэтому заданные цвета
// действуют и на вложенные атрибуты с тем же именем. Сообщение записи (msg)
// в этом смысле рядовой ключ: у него есть цвет по умолчанию, он перекрывается
// как любой другой, и его же получает одноимённый атрибут, включая вложенный.
func WithColorizeAttr(attrKey, colorKey, colorValue string) Option {
	return func(o *options) {
		color := attrColor{
			key:   colorKey,
			value: colorValue,
		}

		if attrKey == AttrColorByDefaultKey {
			o.attrColorByDefault = color

			return
		}

		o.attrKey2color[attrKey] = color
	}
}
