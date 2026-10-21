package slog_test

// база часовых поясов встраивается в тестовый бинарник, чтобы тесты
// проходили в минимальных образах, где она отсутствует в системе.
import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog/slog"
)

// rfc3339Regexp - выделяет время в формате RFC3339 из строки лога.
// Позволяет проверять все обработчики одинаково, независимо от того,
// как каждый из них оформляет запись (JSON, key=value, ANSI-коды).
var rfc3339Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:\d{2})`)

// localOffset - возвращает смещение часового пояса процесса в виде,
// в котором оно печатается форматом RFC3339 ("Z" для UTC, иначе "+03:00").
// Используется, чтобы тесты не зависели от пояса машины, на которой запущены.
func localOffset() string {
	return time.Now().Format("Z07:00")
}

// TestNewLoggerAdapter_TimeZoneByDefault - проверяет, что когда часовой пояс
// не задан явно, время выводится в UTC во всех режимах: режим вывода
// на выбор пояса не влияет.
func TestNewLoggerAdapter_TimeZoneByDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		jsonFormat bool
		colorMode  bool
	}{
		{
			name:       "colored mode",
			jsonFormat: false,
			colorMode:  true,
		},
		{
			name:       "json format",
			jsonFormat: true,
			colorMode:  false,
		},
		{
			name:       "json format has priority over color mode",
			jsonFormat: true,
			colorMode:  true,
		},
		{
			name:       "text format",
			jsonFormat: false,
			colorMode:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := logTimeValue(
				t,
				slog.WithJsonFormat(tt.jsonFormat),
				slog.WithColorMode(tt.colorMode),
				slog.WithTimeFormat("RFC3339"),
			)

			assert.True(
				t,
				strings.HasSuffix(got, "Z"),
				"time %q must have suffix %q",
				got,
				"Z",
			)
		})
	}
}

// TestNewLoggerAdapter_TimeZoneLocal - проверяет, что часовой пояс процесса
// используется только когда он задан явно значением "Local".
// На UTC-машине проверка вырождается в ожидание UTC, но случаи
// Europe/Moscow и Asia/Tokyo в TestNewLoggerAdapter_TimeZone уже
// гарантируют, что настройка не игнорируется.
func TestNewLoggerAdapter_TimeZoneLocal(t *testing.T) {
	t.Parallel()

	got := logTimeValue(
		t,
		slog.WithTimeFormat("RFC3339"),
		slog.WithTimeZone("Local"),
	)

	wantSuffix := localOffset()

	assert.True(
		t,
		strings.HasSuffix(got, wantSuffix),
		"time %q must have suffix %q",
		got,
		wantSuffix,
	)
}

// TestNewLoggerAdapter_TimeZone - проверяет, что заданный часовой пояс
// одинаково учитывается всеми обработчиками. Проверяются несколько поясов,
// поэтому тест не зависит от пояса машины, на которой запущен:
// игнорирование настройки завалит как минимум один из случаев.
func TestNewLoggerAdapter_TimeZone(t *testing.T) {
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

	zones := []struct {
		timeZone   string
		wantSuffix string
	}{
		{timeZone: "UTC", wantSuffix: "Z"},
		{timeZone: "Europe/Moscow", wantSuffix: "+03:00"},
		{timeZone: "Asia/Tokyo", wantSuffix: "+09:00"},
	}

	for _, mode := range modes {
		for _, zone := range zones {
			t.Run(mode.name+"/"+zone.timeZone, func(t *testing.T) {
				t.Parallel()

				got := logTimeValue(
					t,
					slog.WithJsonFormat(mode.jsonFormat),
					slog.WithColorMode(mode.colorMode),
					slog.WithTimeFormat("RFC3339"),
					slog.WithTimeZone(zone.timeZone),
				)

				assert.True(
					t,
					strings.HasSuffix(got, zone.wantSuffix),
					"time %q must have suffix %q",
					got,
					zone.wantSuffix,
				)
			})
		}
	}
}

func TestNewLoggerAdapter_TimeZoneError(t *testing.T) {
	t.Parallel()

	logger, err := slog.NewLoggerAdapter(
		slog.WithWriter(&bytes.Buffer{}),
		slog.WithTimeZone("Nowhere/Nowhere"),
	)

	require.Error(t, err)
	assert.Nil(t, logger)
}

// logTimeValue - логирует сообщение с указанными опциями
// и возвращает время, выделенное из полученной записи.
func logTimeValue(t *testing.T, opts ...slog.Option) string {
	t.Helper()

	var buf bytes.Buffer

	logger, err := slog.NewLoggerAdapter(append([]slog.Option{slog.WithWriter(&buf)}, opts...)...)
	require.NoError(t, err)

	logger.Info(t.Context(), "any message")

	got := rfc3339Regexp.FindString(buf.String())
	require.NotEmpty(t, got, "RFC3339 time is not found in the log record: %q", buf.String())

	return got
}
