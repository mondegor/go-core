package timezone_test

// база часовых поясов встраивается в тестовый бинарник, чтобы тесты
// проходили в минимальных образах, где она отсутствует в системе.
import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/timezone"
)

func TestLocationList_LocationByName(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Europe/Moscow"})

	tests := []struct {
		name    string
		value   string
		want    *time.Location
		wantErr bool
	}{
		{
			name:  "registered iana name",
			value: "Europe/Moscow",
			want:  mustLoadLocation(t, "Europe/Moscow"),
		},
		{
			name:  "utc is registered by default",
			value: "UTC",
			want:  time.UTC,
		},
		{
			// пояс процесса наружу не отдаётся: в список он не входит,
			// поэтому промах сводится к поясу по умолчанию
			name:    "local is not registered",
			value:   "Local",
			want:    mustLoadLocation(t, "Europe/Moscow"),
			wantErr: true,
		},
		{
			// промах сводится к поясу по умолчанию, а не к UTC: в этом списке
			// они различаются, поэтому подмена одного другим была бы заметна
			name:    "empty value returns the default zone and error",
			value:   "",
			want:    mustLoadLocation(t, "Europe/Moscow"),
			wantErr: true,
		},
		{
			name:    "known but not registered name returns the default zone and error",
			value:   "Asia/Tokyo",
			want:    mustLoadLocation(t, "Europe/Moscow"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := list.LocationByName(tc.value)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

// TestNewLocationList_SkipsInvalidNames - фиксирует, что негодные имена
// пропускаются, а не прерывают создание списка: имя, стоящее после негодного,
// должно быть зарегистрировано. Отвергать такой список - забота
// config.ValidateTimeZones, конструктор ошибку не возвращает.
func TestNewLocationList_SkipsInvalidNames(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"", "Nowhere/Nowhere", "Europe/Moscow"})

	t.Run("valid name after invalid ones is registered", func(t *testing.T) {
		t.Parallel()

		got, err := list.LocationByName("Europe/Moscow")

		require.NoError(t, err)
		assert.Equal(t, mustLoadLocation(t, "Europe/Moscow"), got)
	})

	t.Run("valid name after invalid ones is indexed by offset", func(t *testing.T) {
		t.Parallel()

		// Москва: +03:00 круглый год, перехода на летнее время нет
		got, ok := list.NameByOffset(3*time.Hour, false)

		require.True(t, ok)
		assert.Equal(t, "Europe/Moscow", got)
	})

	t.Run("unknown name is not registered", func(t *testing.T) {
		t.Parallel()

		got, err := list.LocationByName("Nowhere/Nowhere")

		require.Error(t, err)
		// негодные имена списком не зарегистрированы, поэтому поясом
		// по умолчанию стало первое годное имя
		assert.Equal(t, mustLoadLocation(t, "Europe/Moscow"), got)
	})

	t.Run("utc is registered by default", func(t *testing.T) {
		t.Parallel()

		utc, err := list.LocationByName("UTC")
		require.NoError(t, err)
		assert.Equal(t, time.UTC, utc)
	})

	t.Run("local is not registered", func(t *testing.T) {
		t.Parallel()

		// пояс процесса наружу не отдаётся даже как всегда доступное имя
		_, err := list.LocationByName("Local")
		require.Error(t, err)
	})
}

func TestLocationList_Default(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		names []string
		want  *time.Location
	}{
		{
			name:  "first name of the list",
			names: []string{"Europe/Moscow", "Asia/Tokyo"},
			want:  mustLoadLocation(t, "Europe/Moscow"),
		},
		{
			name:  "empty list falls back to utc",
			names: nil,
			want:  time.UTC,
		},
		{
			// Local наружу не отдаётся, поэтому поясом по умолчанию
			// становится следующее годное имя списка
			name:  "local is skipped",
			names: []string{"Local", "Europe/Moscow"},
			want:  mustLoadLocation(t, "Europe/Moscow"),
		},
		{
			name:  "unloadable names are skipped",
			names: []string{"", "Nowhere/Bad", "Asia/Tokyo"},
			want:  mustLoadLocation(t, "Asia/Tokyo"),
		},
		{
			// UTC зарегистрирован до обхода списка, но первым годным именем
			// остаётся именно он, поэтому он же становится поясом по умолчанию
			name:  "explicit utc is the first usable name",
			names: []string{"UTC", "Europe/Moscow"},
			want:  time.UTC,
		},
		{
			// повтор пропускается, но пояс по умолчанию задаётся первым вхождением
			name:  "duplicate does not shift the default",
			names: []string{"Asia/Tokyo", "Asia/Tokyo", "Europe/Moscow"},
			want:  mustLoadLocation(t, "Asia/Tokyo"),
		},
		{
			name:  "entirely unusable list falls back to utc",
			names: []string{"", "Nowhere/Bad"},
			want:  time.UTC,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, timezone.NewLocationList(tc.names).Default())
		})
	}
}

// TestLocationList_TimeZones - проверяет, что список имён отличается от переданного:
// UTC присутствует всегда, а пустые имена, "Local", не найденные в базе часовых поясов
// и повторы отброшены.
func TestLocationList_TimeZones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{
			name:  "utc is added to the list",
			names: []string{"Europe/Moscow", "Asia/Tokyo"},
			want:  []string{"UTC", "Europe/Moscow", "Asia/Tokyo"},
		},
		{
			// UTC регистрируется всегда, поэтому в именах не удваивается
			name:  "explicit utc is not duplicated",
			names: []string{"UTC", "Europe/Moscow"},
			want:  []string{"UTC", "Europe/Moscow"},
		},
		{
			// UTC регистрируется до обхода списка, поэтому своей позиции в нём не занимает
			name:  "explicit utc is hoisted to the front",
			names: []string{"Europe/Moscow", "UTC", "Asia/Tokyo"},
			want:  []string{"UTC", "Europe/Moscow", "Asia/Tokyo"},
		},
		{
			name:  "duplicates are collapsed",
			names: []string{"Europe/Moscow", "Europe/Moscow"},
			want:  []string{"UTC", "Europe/Moscow"},
		},
		{
			// пояс процесса наружу не отдаётся, пустое имя и неизвестный пояс негодны
			name:  "empty, local and unknown names are dropped",
			names: []string{"", "Local", "Mars/Olympus", "Europe/Moscow"},
			want:  []string{"UTC", "Europe/Moscow"},
		},
		{
			name:  "empty list still provides utc",
			names: nil,
			want:  []string{"UTC"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, timezone.NewLocationList(tc.names).TimeZones())
		})
	}
}

// TestLocationList_TimeZones_MatchLocationByName - проверяет, что список и строгий подбор
// опираются на одни и те же имена: каждое имя списка принимается LocationByName.
// Негодные имена в исходном списке заданы затем, чтобы отсев не разошёлся между ними.
func TestLocationList_TimeZones_MatchLocationByName(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"", "Local", "Mars/Olympus", "Europe/Moscow", "Asia/Tokyo"})

	names := list.TimeZones()
	require.NotEmpty(t, names) // иначе проверка выродилась бы в пустой цикл

	for _, name := range names {
		_, err := list.LocationByName(name)
		assert.NoError(t, err, name)
	}
}

// TestLocationList_TimeZones_ReturnsCopy - проверяет, что выданный срез принадлежит
// вызывающему: его изменение на состав поясов списка не влияет.
func TestLocationList_TimeZones_ReturnsCopy(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Europe/Moscow"})

	got := list.TimeZones()
	got[0] = "Mars/Olympus"

	assert.Equal(t, []string{"UTC", "Europe/Moscow"}, list.TimeZones())
}

// mustLoadLocation - загружает часовой пояс или прерывает тест.
func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()

	loc, err := time.LoadLocation(name)
	require.NoError(t, err)

	return loc
}
