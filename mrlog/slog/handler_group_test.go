package slog_test

import (
	"bytes"
	stdlog "log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog/color"
	"github.com/mondegor/go-core/mrlog/slog"
)

// TestNewLoggerAdapter_NestedBuiltinKeys - проверяет, что вложенные атрибуты
// с именами системных полей записи (time, level, msg) выводятся как обычные
// значения: настройки времени и оформление уровня к ним не применяются,
// а собственное сообщение записи ими не подменяется.
func TestNewLoggerAdapter_NestedBuiltinKeys(t *testing.T) {
	t.Parallel()

	// дата в прошлом и с ненулевым смещением: пересчёт в UTC сдвинет её на сутки назад
	// (1999-01-01T21:30:45Z), а собственное время записи (time.Now()) заведомо не содержит
	// ни одной из этих дат, поэтому проверки не зависят от момента запуска теста
	nested := time.Date(1999, time.January, 2, 2, 30, 45, 0, time.FixedZone("", 5*60*60))

	// служебный уровень Trace: каждый обработчик оформлял бы его по-своему
	// ("TRC" в цветном режиме, "TRACE" в text и json), а нетронутый атрибут
	// печатается родным представлением slog — "ERROR+8". Обычный DEBUG здесь
	// не годится: level.Enum и slog.Level дают для него одну и ту же строку,
	// поэтому проверка не отличила бы оформленный атрибут от нетронутого
	nestedLevel := stdlog.LevelError + 8

	modes := []struct {
		name       string
		jsonFormat bool
		colorMode  bool
	}{
		{name: "json", jsonFormat: true, colorMode: false},
		{name: "colored", jsonFormat: false, colorMode: true},
		{name: "text", jsonFormat: false, colorMode: false},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger, err := slog.NewLoggerAdapter(
				slog.WithWriter(&buf),
				slog.WithJsonFormat(mode.jsonFormat),
				slog.WithColorMode(mode.colorMode),
				slog.WithTimeFormat("RFC3339"),
				slog.WithTimeZone("UTC"),
			)
			require.NoError(t, err)

			logger.Info(
				t.Context(),
				"any message",
				stdlog.Group("req", "time", nested, "level", nestedLevel, "msg", "nested message"),
			)

			got := buf.String()

			// каждый обработчик печатает обычный атрибут-время по-своему,
			// поэтому проверяется не формат, а сам момент времени: он должен
			// остаться исходным (1999-01-02), а не быть пересчитан в UTC (1999-01-01)
			assert.Contains(t, got, "1999-01-02", "log record: %q", got)
			assert.NotContains(
				t,
				got,
				"1999-01-01",
				"nested time attr must not be converted to the logger timezone",
			)

			assert.Contains(t, got, "ERROR+8", "log record: %q", got)
			assert.NotContains(
				t,
				got,
				"TRC",
				"nested level attr must not be styled as the record level",
			)
			assert.NotContains(
				t,
				got,
				"TRACE",
				"nested level attr must not be styled as the record level",
			)

			// вложенный msg выводится как обычный атрибут и не подменяет
			// собственное сообщение записи: в выводе присутствуют оба
			assert.Contains(t, got, "nested message", "log record: %q", got)
			assert.Contains(t, got, "any message", "log record: %q", got)
		})
	}
}

// TestNewLoggerAdapter_TopLevelBuiltinKeyCollision - проверяет, что атрибут
// верхнего уровня, случайно названный именем системного поля записи, но несущий
// другой тип, оформляется как обычный атрибут. Системные поля записи всегда
// приходят со своим типом (time.Time и slog.Level), поэтому несовпадение типа
// однозначно означает пользовательский атрибут, и его нельзя оставлять
// без оформления только из-за совпадения имени.
func TestNewLoggerAdapter_TopLevelBuiltinKeyCollision(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(
		slog.WithWriter(&buf),
		slog.WithColorMode(true),
		slog.WithTimeFormat("RFC3339"),
		slog.WithTimeZone("UTC"),
	)
	require.NoError(t, err)

	// значения намеренно строковые: у системных полей здесь были бы
	// time.Time и slog.Level
	logger.Info(t.Context(), "any message", "level", "high", "time", "yesterday")

	got := buf.String()

	// цвет значения по умолчанию задан в NewLoggerAdapter (attrColorByDefault)
	for _, value := range []string{"high", "yesterday"} {
		assert.Contains(
			t,
			got,
			color.ColorizeText(color.LightGray, value),
			"attr %q must be colorized as a regular one, log record: %q",
			value,
			got,
		)
	}
}

// TestNewLoggerAdapter_MessageIsColorized - проверяет, что сообщение раскрашивается
// всегда. В отличие от time и level, у msg нет типового признака, по которому
// сообщение записи отличалось бы от пользовательского атрибута с таким же именем,
// поэтому оба оформляются одинаково, на общих основаниях.
func TestNewLoggerAdapter_MessageIsColorized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []slog.Option
		want string
	}{
		{
			// цвет сообщения по умолчанию задан в NewLoggerAdapter (attrKey2color)
			name: "record message",
			want: color.ColorizeText(color.White, "any message"),
		},
		{
			// именно этот случай ранее выводился без раскраски
			name: "top level attr named as the message key",
			want: color.ColorizeText(color.White, "attr value"),
		},
		{
			name: "default message color is overridden",
			opts: []slog.Option{
				slog.WithColorizeAttr(stdlog.MessageKey, color.Cyan, color.Magenta),
			},
			want: color.ColorizeText(color.Magenta, "any message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger, err := slog.NewLoggerAdapter(
				append(
					[]slog.Option{slog.WithWriter(&buf), slog.WithColorMode(true)},
					tt.opts...,
				)...,
			)
			require.NoError(t, err)

			logger.Info(t.Context(), "any message", stdlog.MessageKey, "attr value")

			assert.Contains(t, buf.String(), tt.want, "log record: %q", buf.String())
		})
	}
}

// TestNewLoggerAdapter_DroppedAttr - проверяет, что атрибут, отброшенный
// через WithReplaceAttrs, не попадает в вывод ни в одном из режимов.
// Цветной режим переписывает ключ атрибута, поэтому пустой атрибут должен
// возвращаться до раскраски: иначе ANSI-коды в ключе не дадут обработчику
// его отбросить, и запись получит поле с пустым именем и значением <nil>.
func TestNewLoggerAdapter_DroppedAttr(t *testing.T) {
	t.Parallel()

	modes := []struct {
		name       string
		jsonFormat bool
		colorMode  bool
	}{
		{name: "json", jsonFormat: true, colorMode: false},
		{name: "colored", jsonFormat: false, colorMode: true},
		{name: "text", jsonFormat: false, colorMode: false},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger, err := slog.NewLoggerAdapter(
				slog.WithWriter(&buf),
				slog.WithJsonFormat(mode.jsonFormat),
				slog.WithColorMode(mode.colorMode),
				slog.WithReplaceAttrs(func(attr stdlog.Attr) stdlog.Attr {
					if attr.Key == "secret" {
						return stdlog.Attr{}
					}

					return attr
				}),
			)
			require.NoError(t, err)

			logger.Info(t.Context(), "any message", "secret", "hunter2", "kept", "value")

			got := buf.String()

			assert.NotContains(t, got, "hunter2", "dropped attr value, log record: %q", got)
			assert.NotContains(t, got, "<nil>", "dropped attr placeholder, log record: %q", got)
			assert.Contains(t, got, "value", "log record: %q", got)
		})
	}
}

// TestNewLoggerAdapter_NestedMessageIsColorized - фиксирует принятый компромисс:
// цвета назначаются по имени атрибута без учёта группы, поэтому вложенный msg
// получает цвет сообщения записи, а не цвет по умолчанию.
func TestNewLoggerAdapter_NestedMessageIsColorized(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(slog.WithWriter(&buf), slog.WithColorMode(true))
	require.NoError(t, err)

	logger.Info(
		t.Context(),
		"any message",
		stdlog.Group("req", stdlog.MessageKey, "nested message", "other", "nested other"),
	)

	got := buf.String()

	assert.Contains(
		t,
		got,
		color.ColorizeText(color.White, "nested message"),
		"nested msg must get the message color, log record: %q",
		got,
	)
	assert.Contains(
		t,
		got,
		color.ColorizeText(color.LightGray, "nested other"),
		"other nested attrs must keep the default color, log record: %q",
		got,
	)
}
