package instance_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/errors/runtime/instance"
)

// TestGenerateID_Format проверяет базовый формат генерируемого ID.
// ID должен состоять из 3 частей, разделённых дефисами: XXXXXXXXXX-XXXX-XXXX.
// Формат: 10 символов timestamp, дефис, 4 символа random, дефис, 4 символа random.
func TestGenerateID_Format(t *testing.T) {
	t.Parallel()

	id := instance.GenerateID()

	// Проверяем общую длину
	require.Len(t, id, 20, "Длина ID должна быть 20 символов")

	parts := strings.Split(id, "-")
	require.Len(t, parts, 3, "ID должен содержать ровно 3 части, разделённые дефисами")

	// Проверяем длину каждой части
	assert.Len(t, parts[0], 10, "Первая часть (timestamp) должна быть 10 символов")
	assert.Len(t, parts[1], 4, "Вторая часть (random1) должна быть 4 символа")
	assert.Len(t, parts[2], 4, "Третья часть (random2) должна быть 4 символа")
}

// TestGenerateID_TimestampDecodes проверяет, что первая часть ID - это UnixNano
// момента генерации в base36 без старшего и младшего разрядов (12 разрядов -> 10 символов).
func TestGenerateID_TimestampDecodes(t *testing.T) {
	t.Parallel()

	const (
		base   = 36
		modulo = 3656158440062976 // 36^10
	)

	now := uint64(time.Now().UnixNano())
	id := instance.GenerateID()

	got, err := strconv.ParseUint(id[:10], base, 64)
	require.NoError(t, err)

	// метка усечена по модулю 36^10, поэтому берётся кратчайшее смещение между got и now по кругу;
	// допуск с запасом покрывает время вызова и небольшой перевод системных часов
	diff := (now/base + modulo - got) % modulo
	if diff > modulo/2 {
		diff = modulo - diff
	}

	assert.LessOrEqual(t, diff*base, uint64(time.Minute))
}

// TestGenerateID_ContainsOnlyUppercaseAlphanumeric проверяет, что ID содержит
// только заглавные буквы латинского алфавита (A-Z) и цифры (0-9).
func TestGenerateID_ContainsOnlyUppercaseAlphanumeric(t *testing.T) {
	t.Parallel()

	id := instance.GenerateID()

	for _, ch := range id {
		assert.True(t, (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || ch == '-')
	}
}

// TestGenerateID_DashesAtCorrectPositions проверяет наличие и расположение дефисов.
// В ID должно быть ровно 2 дефиса, разделяющих timestamp и случайные компоненты.
// Формат: XXXXXXXXXX-XXXX-XXXX (дефисы на позициях 10 и 15).
func TestGenerateID_DashesAtCorrectPositions(t *testing.T) {
	t.Parallel()

	id := instance.GenerateID()

	require.Len(t, id, 20, "Длина ID должна быть 20 символов")
	assert.Equal(t, uint8('-'), id[10], "Дефис должен быть на позиции 10")
	assert.Equal(t, uint8('-'), id[15], "Дефис должен быть на позиции 15")
}

// TestGenerateID_Unique проверяет уникальность генерируемых ID.
// Генерируется 1000 ID и проверяется, что все они уникальны.
func TestGenerateID_Unique(t *testing.T) {
	t.Parallel()

	const iterations = 1000

	ids := make(map[string]bool, iterations)

	for i := 0; i < iterations; i++ {
		id := instance.GenerateID()
		assert.False(t, ids[id], "Обнаружен дубликат ID на итерации %d: %s", i, id)
		ids[id] = true
	}

	assert.Len(t, ids, iterations, "Должно быть сгенерировано %d уникальных ID", iterations)
}

// TestGenerateID_Concurrent проверяет потокобезопасность GenerateID().
// Запускается 10 горутин, каждая генерирует 100 ID.
// Все 1000 ID должны быть уникальны, race conditions отсутствуют.
func TestGenerateID_Concurrent(t *testing.T) {
	t.Parallel()

	const (
		goroutines      = 10
		idsPerGoroutine = 100
	)

	allIDs := make(map[string]bool)

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	// Запускаем 10 горутин
	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(goroutineID int) {
			defer wg.Done()

			localIDs := make([]string, 0, idsPerGoroutine)

			// Каждая горутина генерирует 100 ID
			for i := 0; i < idsPerGoroutine; i++ {
				localIDs = append(localIDs, instance.GenerateID())
			}

			// Синхронизируем доступ к общей карте
			mu.Lock()
			defer mu.Unlock()

			for _, id := range localIDs {
				assert.False(t, allIDs[id], "Найден дубликат ID из горутины %d: %s", goroutineID, id)
				allIDs[id] = true
			}
		}(g)
	}

	wg.Wait()

	expectedCount := goroutines * idsPerGoroutine
	assert.Len(t, allIDs, expectedCount, "Должно быть сгенерировано %d уникальных ID в %d горутинах", expectedCount, goroutines)
}

// TestGenerateID_TimeBasedComponent проверяет временную компоненту ID.
// ID, сгенерированные с задержкой, должны иметь одинаковый формат структуры.
func TestGenerateID_TimeBasedComponent(t *testing.T) {
	t.Parallel()

	// Генерируем первый ID
	id1 := instance.GenerateID()

	// Небольшая задержка для изменения timestamp компоненты
	for i := 0; i < 1000; i++ {
		_ = i * i // занимаем процессорное время
	}

	// Генерируем второй ID
	id2 := instance.GenerateID()

	// Оба ID должны иметь одинаковую структуру (2 дефиса, 3 части)
	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	assert.Len(t, parts1, 3, "Первый ID должен состоять из 3 частей")
	assert.Len(t, parts2, 3, "Второй ID должен состоять из 3 частей")

	// Оба ID должны иметь одинаковое количество частей
	assert.Len(t, parts1, len(parts2), "Оба ID должны иметь одинаковое количество частей")
}

// TestGenerateID_RandomComponent проверяет случайную компоненту ID.
// Даже при генерации в один момент времени (одинаковый timestamp),
// ID должны быть уникальны благодаря crypto/rand.
func TestGenerateID_RandomComponent(t *testing.T) {
	t.Parallel()

	// Генерируем множество ID быстро (вероятно одинаковый timestamp)
	const count = 100

	ids := make([]string, count)

	for i := 0; i < count; i++ {
		ids[i] = instance.GenerateID()
	}

	// Проверяем, что все ID уникальны
	// Случайная компонента обеспечивает уникальность даже при одинаковом timestamp
	seen := make(map[string]bool)
	for i, id := range ids {
		assert.False(t, seen[id], "Дубликат ID на индексе %d: %s", i, id)
		seen[id] = true
	}
}

// TestGenerateID_NoLowercaseLetters проверяет отсутствие строчных букв.
// Все буквы в ID должны быть в верхнем регистре (A-Z, не a-z).
func TestGenerateID_NoLowercaseLetters(t *testing.T) {
	t.Parallel()

	// Генерируем 100 ID и проверяем каждый символ
	for i := 0; i < 100; i++ {
		id := instance.GenerateID()
		for j, ch := range id {
			assert.True(t, ch < 'a' || ch > 'z', "Строчная буква найдена на позиции %d: %c", j, ch)
		}
	}
}

// TestGenerateID_StatisticalDistribution проверяет статистическое распределение
// символов в ID. Base36 использует 36 символов (0-9, A-Z).
// Проверяем, что используются и буквы, и цифры.
func TestGenerateID_StatisticalDistribution(t *testing.T) {
	t.Parallel()

	const sampleSize = 100

	charCounts := make(map[rune]int)

	// Генерируем выборку ID
	for i := 0; i < sampleSize; i++ {
		id := instance.GenerateID()
		for _, ch := range id {
			if ch != '-' {
				charCounts[ch]++
			}
		}
	}

	// Ожидаем разумное распределение символов
	// Base36 использует 0-9 и A-Z (36 символов)
	assert.Greater(t, len(charCounts), 5, "В выборке должно быть минимум 5 различных символов")

	// Проверяем, что используются и буквы, и цифры
	hasLetter := false
	hasNumber := false

	for ch := range charCounts {
		if ch >= 'A' && ch <= 'Z' {
			hasLetter = true
		}

		if ch >= '0' && ch <= '9' {
			hasNumber = true
		}
	}

	assert.True(t, hasLetter, "ID должны содержать заглавные буквы")
	assert.True(t, hasNumber, "ID должны содержать цифры")
}

// TestGenerateID_PrefixVariation проверяет вариативность префикса ID
// (первая часть перед первым дефисом). Эта часть основана на timestamp,
// поэтому должна варьироваться при генерации в разное время.
func TestGenerateID_PrefixVariation(t *testing.T) {
	t.Parallel()

	// Генерируем ID и проверяем, что префиксы варьируются
	const count = 50

	prefixes := make(map[string]bool)

	for i := 0; i < count; i++ {
		id := instance.GenerateID()
		// Берём первую часть (timestamp компонента)
		parts := strings.Split(id, "-")
		prefix := parts[0]
		prefixes[prefix] = true
	}

	// С временной компонентой ожидаем вариативность префиксов
	// (если только все ID не сгенерированы в одну наносекунду)
	assert.Greater(t, len(prefixes), 1, "Должна быть вариативность в префиксах ID")
}

// TestGenerateID_SuffixVariation проверяет вариативность суффикса ID
// (случайная часть после первого дефиса). Благодаря crypto/rand
// все суффиксы должны быть уникальны.
func TestGenerateID_SuffixVariation(t *testing.T) {
	t.Parallel()

	// Генерируем ID и проверяем, что случайные части варьируются
	const count = 50

	suffixes := make(map[string]bool)

	for i := 0; i < count; i++ {
		id := instance.GenerateID()
		// Берём части после первого дефиса (random компоненты)
		parts := strings.Split(id, "-")
		suffix := parts[1] + "-" + parts[2]
		suffixes[suffix] = true
	}

	// Случайная компонента должна обеспечивать уникальность всех суффиксов
	assert.Len(t, suffixes, count, "Все суффиксы должны быть уникальны благодаря случайности")
}

// TestGenerateID_MatchesExpectedPattern проверяет полное соответствие ID
// ожидаемому паттерну: [A-Z0-9]+-[A-Z0-9]+-[A-Z0-9]+
// Тест запускается на 100 итерациях для надёжности.
func TestGenerateID_MatchesExpectedPattern(t *testing.T) {
	t.Parallel()

	// Паттерн: одна или более base36 символов, дефис, повторить 3 раза
	const iterations = 100

	for i := 0; i < iterations; i++ {
		id := instance.GenerateID()

		// Проверяем, что ID содержит ровно 2 дефиса
		dashCount := strings.Count(id, "-")
		assert.Equal(t, 2, dashCount, "Итерация %d: ID должен содержать ровно 2 дефиса", i)

		// Проверяем каждый символ
		for j, ch := range id {
			if ch == '-' {
				continue // дефисы ожидаемы
			}

			assert.True(t,
				(ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z'),
				"Итерация %d: Позиция %d должна содержать заглавную букву или цифру, получено %c", i, j, ch)
		}
	}
}
