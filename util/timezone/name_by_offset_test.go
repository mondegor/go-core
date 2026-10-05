package timezone_test

// база часовых поясов встраивается в тестовый бинарник, чтобы тесты
// проходили в минимальных образах, где она отсутствует в системе.
// Встроенная база используется, только если системной нет.
import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/timezone"
)

// TestLocationList_NameByOffset - проверяет подбор пояса по паре
// (смещение, признак летнего времени) на списке, в котором пары
// не пересекаются, поэтому результат не зависит от порядка имён.
func TestLocationList_NameByOffset(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{
		"Europe/Berlin",    // +01:00, летом +02:00
		"Asia/Tokyo",       // +09:00, перехода нет
		"Australia/Sydney", // +10:00, летом (январь) +11:00
		"America/New_York", // -05:00, летом -04:00
	})

	tests := []struct {
		name   string
		offset time.Duration
		isDST  bool
		want   string
	}{
		{
			name:   "zone with dst in standard state",
			offset: 1 * time.Hour,
			want:   "Europe/Berlin",
		},
		{
			name:   "zone with dst in daylight state",
			offset: 2 * time.Hour,
			isDST:  true,
			want:   "Europe/Berlin",
		},
		{
			name:   "zone without dst",
			offset: 9 * time.Hour,
			want:   "Asia/Tokyo",
		},
		{
			// южное полушарие: стандартное время приходится на июль
			name:   "southern hemisphere in standard state",
			offset: 10 * time.Hour,
			want:   "Australia/Sydney",
		},
		{
			// южное полушарие: летнее время приходится на январь
			name:   "southern hemisphere in daylight state",
			offset: 11 * time.Hour,
			isDST:  true,
			want:   "Australia/Sydney",
		},
		{
			name:   "negative offset in standard state",
			offset: -5 * time.Hour,
			want:   "America/New_York",
		},
		{
			name:   "negative offset in daylight state",
			offset: -4 * time.Hour,
			isDST:  true,
			want:   "America/New_York",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := list.NameByOffset(tc.offset, tc.isDST)

			require.True(t, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestLocationList_NameByOffset_NotFound - проверяет, что промах сообщается
// явно, а не подменяется каким-либо именем: вызывающий должен отличать
// "пояс не подобран" от любого успешного подбора.
func TestLocationList_NameByOffset_NotFound(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Europe/Berlin", "Asia/Tokyo"})

	tests := []struct {
		name   string
		offset time.Duration
		isDST  bool
	}{
		{
			// смещение зарегистрировано, но только как летнее
			name:   "daylight offset requested as standard",
			offset: 2 * time.Hour,
		},
		{
			// смещение зарегистрировано, но только как стандартное
			name:   "standard offset requested as daylight",
			offset: 1 * time.Hour,
			isDST:  true,
		},
		{
			// у пояса без перехода летнего состояния в индексе нет
			name:   "zone without dst requested as daylight",
			offset: 9 * time.Hour,
			isDST:  true,
		},
		{
			name:   "offset is not registered",
			offset: 13 * time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := list.NameByOffset(tc.offset, tc.isDST)

			assert.False(t, ok)
			assert.Empty(t, got, "the name must not be set when the zone is not found")
		})
	}
}

// TestLocationList_NameByOffset_FractionalDSTShift - проверяет пояс, у которого
// сдвиг на летнее время не равен часу: величина сдвига нигде не должна
// предполагаться, иначе такой пояс подбирался бы неверно.
func TestLocationList_NameByOffset_FractionalDSTShift(t *testing.T) {
	t.Parallel()

	// Australia/Lord_Howe: +10:30, летом +11:00, то есть сдвиг 30 минут
	list := timezone.NewLocationList([]string{"Australia/Lord_Howe"})

	assertZone(t, list, 10*time.Hour+30*time.Minute, false, "Australia/Lord_Howe")
	assertZone(t, list, 11*time.Hour, true, "Australia/Lord_Howe")

	// вычитание захардкоженного часа дало бы здесь +10:00
	_, ok := list.NameByOffset(10*time.Hour, false)
	assert.False(t, ok)
}

// TestLocationList_NameByOffset_PermanentDST - проверяет пояс, у которого признак
// летнего времени стоит не там, где его ждёт зондирование в двух точках года.
//
// Africa/Casablanca большую часть года живёт в +01, а в месяц Рамадана уходит
// на +00. Признак летнего времени зависит от формата базы часовых поясов:
//   - rearguard (системная база ряда дистрибутивов): +01 с признаком круглый год,
//     +00 без признака;
//   - основной (встроенная в Go time/tzdata, системная база Ubuntu): +01 без
//     признака, +00 с признаком (отрицательное летнее время).
//
// Какой формат загрузится, решает окружение: time.LoadLocation сначала читает
// системную базу, поэтому тест проверяет состояния того формата, что реально загружен.
func TestLocationList_NameByOffset_PermanentDST(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Africa/Casablanca")
	require.NoError(t, err)

	list := timezone.NewLocationList([]string{"Africa/Casablanca"})

	// в июле Рамадана нет, поэтому пояс находится в своём основном состоянии +01
	isRearguard := time.Date(2026, time.July, 1, 12, 0, 0, 0, loc).IsDST()

	if isRearguard {
		// круглогодичное состояние пояса подбирается
		assertZone(t, list, 1*time.Hour, true, "Africa/Casablanca")

		// второе реальное состояние пояса, (+00, isDST=false), совпадает с парой UTC,
		// а UTC регистрируется последним и потому выигрывает. Через NameByOffset это
		// состояние ненаблюдаемо, проверяется лишь то, что пара разрешима
		assertZone(t, list, 0, false, "UTC")
	} else {
		// оба реальных состояния пояса подбираются
		assertZone(t, list, 1*time.Hour, false, "Africa/Casablanca")
		assertZone(t, list, 0, true, "Africa/Casablanca")
	}

	// состояния (+01) с противоположным признаком у пояса не существует,
	// поэтому его в индексе быть не должно
	got, ok := list.NameByOffset(1*time.Hour, !isRearguard)
	assert.False(t, ok, "a state the zone is never in must not be registered")
	assert.Empty(t, got)
}

// TestLocationList_NameByOffset_ZonesWithoutDST - проверяет, что для пояса
// без перехода летняя запись в индексе не создаётся.
func TestLocationList_NameByOffset_ZonesWithoutDST(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Asia/Tokyo", "Africa/Lagos", "Asia/Tehran"})

	tests := []struct {
		name   string
		offset time.Duration
		want   string
	}{
		{name: "tokyo", offset: 9 * time.Hour, want: "Asia/Tokyo"},
		{name: "lagos", offset: 1 * time.Hour, want: "Africa/Lagos"},
		{name: "tehran", offset: 3*time.Hour + 30*time.Minute, want: "Asia/Tehran"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertZone(t, list, tc.offset, false, tc.want)

			_, ok := list.NameByOffset(tc.offset, true)
			assert.False(t, ok, "dst state must not be registered")
		})
	}
}

// TestLocationList_NameByOffset_Collision - фиксирует выбранное поведение при
// совпадении пары у нескольких поясов: выигрывает пояс, указанный в списке позднее
// (о списке с повторами см. TestLocationList_NameByOffset_DuplicateIgnored).
// Это документирование правила, а не проверка «правильного» ответа —
// Europe/Berlin и Europe/Paris по паре неразличимы.
func TestLocationList_NameByOffset_Collision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		names []string
		want  string
	}{
		{
			name:  "berlin is registered last",
			names: []string{"Europe/Paris", "Europe/Berlin"},
			want:  "Europe/Berlin",
		},
		{
			name:  "paris is registered last",
			names: []string{"Europe/Berlin", "Europe/Paris"},
			want:  "Europe/Paris",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			list := timezone.NewLocationList(tc.names)

			assertZone(t, list, 1*time.Hour, false, tc.want)
			assertZone(t, list, 2*time.Hour, true, tc.want)
		})
	}
}

// TestLocationList_NameByOffset_DuplicateIgnored - проверяет, что повтор имени в списке
// на подбор не влияет: пояс регистрируется по первому вхождению, поэтому повтор
// не отбирает пару ни у пояса, указанного после него, ни у самого себя, то есть
// сравниваются первые вхождения, а не позиции в списке
// (ср. TestLocationList_NameByOffset_Collision).
func TestLocationList_NameByOffset_DuplicateIgnored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		names []string
		want  string
	}{
		{
			// повтор Europe/Berlin пару себе не возвращает, хотя указан последним
			name:  "duplicate does not take the pair back",
			names: []string{"Europe/Berlin", "Europe/Paris", "Europe/Berlin"},
			want:  "Europe/Paris",
		},
		{
			// зеркальный список: последним указан Europe/Paris, но его первое
			// вхождение раньше, поэтому пара остаётся у Europe/Berlin
			name:  "the last name in the list does not win",
			names: []string{"Europe/Paris", "Europe/Berlin", "Europe/Paris"},
			want:  "Europe/Berlin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			list := timezone.NewLocationList(tc.names)

			assertZone(t, list, 1*time.Hour, false, tc.want)
			assertZone(t, list, 2*time.Hour, true, tc.want)
		})
	}
}

// TestLocationList_NameByOffset_LocalExcluded - проверяет, что часовой пояс
// процесса не подбирается: вызывающему требуется IANA-имя.
func TestLocationList_NameByOffset_LocalExcluded(t *testing.T) {
	t.Parallel()

	_, localOffset := time.Now().In(time.Local).Zone()

	tests := []struct {
		name  string
		names []string
	}{
		{
			name:  "local is registered implicitly",
			names: nil,
		},
		{
			name:  "local is listed explicitly",
			names: []string{"Local"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			list := timezone.NewLocationList(tc.names)

			// на машине с UTC смещение совпадёт с UTC, поэтому проверяется
			// не конкретное имя, иначе тест зависел бы от пояса машины
			for _, isDST := range []bool{false, true} {
				got, _ := list.NameByOffset(time.Duration(localOffset)*time.Second, isDST)
				assert.NotEqual(t, "Local", got, "isDST=%v", isDST)
			}
		})
	}
}

// TestLocationList_NameByOffset_UTC - проверяет, что UTC подбирается всегда:
// он регистрируется последним и потому выигрывает свою пару у любого пояса
// из списка. Правило "выигрывает последний" на перечисленные имена
// по-прежнему распространяется (см. TestLocationList_NameByOffset_Collision),
// особый здесь только UTC.
func TestLocationList_NameByOffset_UTC(t *testing.T) {
	t.Parallel()

	t.Run("utc is found as a regular zone", func(t *testing.T) {
		t.Parallel()

		list := timezone.NewLocationList(nil)

		assertZone(t, list, 0, false, "UTC")
	})

	t.Run("utc wins over a listed zone with the same offset", func(t *testing.T) {
		t.Parallel()

		// Atlantic/Reykjavik: +00:00 без перехода, то есть та же пара, что у UTC
		list := timezone.NewLocationList([]string{"Atlantic/Reykjavik"})

		assertZone(t, list, 0, false, "UTC")
	})
}

// TestLocationList_NameByOffset_Horizon - проверяет, что индекс покрывает весь
// год вперёд от создания списка, а не текущий календарный год: иначе у сервиса,
// запущенного в конце декабря, индекс устаревал бы через считанные сутки.
//
// Моменты отсчитываются от текущего дня, поэтому тест проверяет заявленное
// свойство в какой бы день года он ни был запущен.
func TestLocationList_NameByOffset_Horizon(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Europe/Berlin"})

	loc, err := list.LocationByName("Europe/Berlin")
	require.NoError(t, err)

	// пары берутся прямо у пояса, а не задаются константами: так тест
	// не зависит ни от текущего сезона, ни от правил конкретного года
	for _, months := range []int{1, 3, 6, 9, 11} {
		moment := time.Now().In(loc).AddDate(0, months, 0)
		_, offset := moment.Zone()

		got, ok := list.NameByOffset(time.Duration(offset)*time.Second, moment.IsDST())

		assert.True(t, ok, "zone is not found for a moment in %d month(s): %s", months, moment)
		assert.Equal(t, "Europe/Berlin", got)
	}
}

// TestLocationList_NameByOffset_SeasonIndependent - ключевое свойство дизайна:
// все состояния пояса лежат в индексе сразу, поэтому результат подбора
// не зависит от того, в какое время года выполняется запрос.
func TestLocationList_NameByOffset_SeasonIndependent(t *testing.T) {
	t.Parallel()

	list := timezone.NewLocationList([]string{"Europe/Berlin", "Australia/Sydney"})

	// северное полушарие: зима и лето разрешаются одним и тем же экземпляром
	assertZone(t, list, 1*time.Hour, false, "Europe/Berlin")
	assertZone(t, list, 2*time.Hour, true, "Europe/Berlin")

	// южное полушарие: сезоны обратные, но поведение то же
	assertZone(t, list, 10*time.Hour, false, "Australia/Sydney")
	assertZone(t, list, 11*time.Hour, true, "Australia/Sydney")
}

// assertZone - проверяет, что по указанной паре подобран ожидаемый пояс.
func assertZone(t *testing.T, list *timezone.LocationList, offset time.Duration, isDST bool, want string) {
	t.Helper()

	got, ok := list.NameByOffset(offset, isDST)

	assert.True(t, ok, "zone is not found (offset=%s, isDST=%v)", offset, isDST)
	assert.Equal(t, want, got)
}
