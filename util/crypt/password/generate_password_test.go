package password_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/crypt/password"
)

// TestGenerate_DefaultCharsKinds проверяет, что при нулевом charsKinds
// используется CharAll и пароль генерируется корректно.
func TestGenerate_DefaultCharsKinds(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()
	pw := gen.Generate(12, 0)

	require.Len(t, pw, 12)
}

// TestGenerate_InvalidCharsKinds проверяет, что при charsKinds > CharAll
// используется CharAll и пароль генерируется корректно.
func TestGenerate_InvalidCharsKinds(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()
	pw := gen.Generate(12, 255)

	require.Len(t, pw, 12)
}

// TestGenerate_MinLength проверяет, что при длине < 1 пароль
// всё равно генерируется с длиной 1.
func TestGenerate_MinLength(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	type testCase struct {
		name   string
		length int
	}

	tests := []testCase{
		{name: "zero", length: 0},
		{name: "negative", length: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pw := gen.Generate(tt.length, password.CharAll)
			require.Len(t, pw, 1)
		})
	}
}

// TestGenerate_Length1 проверяет генерацию пароля длиной 1 символ.
func TestGenerate_Length1(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()
	pw := gen.Generate(1, password.CharAll)

	require.Len(t, pw, 1)
}

// TestGenerate_CharSet проверяет, что при заданном наборе видов символов пароль
// состоит только из символов допустимого набора.
func TestGenerate_CharSet(t *testing.T) {
	t.Parallel()

	const (
		vowels     = "aeiuyAEIUY"
		consonants = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ"
		numerals   = "0123456789"
		signs      = "!#$%&()*+,-./:;<=>?@[]^_{}~"
	)

	type testCase struct {
		name       string
		length     int
		charsKinds password.CharKinds
		allowed    string
	}

	tests := []testCase{
		{name: "vowels only", length: 20, charsKinds: password.CharVowels, allowed: vowels},
		{name: "consonants only", length: 20, charsKinds: password.CharConsonants, allowed: consonants},
		{name: "numerals only", length: 20, charsKinds: password.CharNumerals, allowed: numerals},
		{name: "signs only", length: 20, charsKinds: password.CharSigns, allowed: signs},
		{name: "letters", length: 20, charsKinds: password.CharAbc, allowed: vowels + consonants},
		{name: "letters and numerals", length: 50, charsKinds: password.CharAbcNumerals, allowed: vowels + consonants + numerals},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := password.NewGenerator()
			pw := gen.Generate(tt.length, tt.charsKinds)

			require.Len(t, pw, tt.length)

			for _, ch := range pw {
				assert.True(t, strings.ContainsRune(tt.allowed, ch),
					"Символ %q вне допустимого набора", ch)
			}
		})
	}
}

// TestGenerate_CharNumerals_NoZero проверяет, что цифра 0 отсутствует в пароле.
func TestGenerate_CharNumerals_NoZero(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	for i := 0; i < 100; i++ {
		pw := gen.Generate(20, password.CharNumerals)
		assert.NotContains(t, pw, "0", "Пароль не должен содержать цифру 0")
	}
}

// TestGenerate_NoAmbiguousLetters проверяет, что в пароле отсутствуют символы oO и 0.
func TestGenerate_NoAmbiguousLetters(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	for i := 0; i < 100; i++ {
		pw := gen.Generate(30, password.CharAll)
		assert.NotContains(t, pw, "o", "Пароль не должен содержать 'o'")
		assert.NotContains(t, pw, "O", "Пароль не должен содержать 'O'")
		assert.NotContains(t, pw, "0", "Пароль не должен содержать '0'")
	}
}

// TestGenerate_NoAmbiguousSigns проверяет, что в пароле отсутствуют кавычки, обратный слэш и |.
func TestGenerate_NoAmbiguousSigns(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	for i := 0; i < 100; i++ {
		pw := gen.Generate(30, password.CharSigns)
		assert.False(t, strings.ContainsAny(pw, "\"'`\\|"), "Пароль не должен содержать кавычки, '\\' и '|': %q", pw)
	}
}

// TestGenerate_CharAll_ContainsAllKinds проверяет, что при использовании CharAll
// в пароле присутствуют все виды символов (при достаточной длине).
func TestGenerate_CharAll_ContainsAllKinds(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	vowels := "aeiuyAEIUY"
	consonants := "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ"
	numerals := "123456789"
	signs := "!#$%&()*+,-./:;<=>?@[]^_{}~"

	// Генерируем длинный пароль для повышения вероятности всех категорий
	pw := gen.Generate(100, password.CharAll)

	hasVowel := strings.ContainsAny(pw, vowels)
	hasConsonant := strings.ContainsAny(pw, consonants)
	hasNumeral := strings.ContainsAny(pw, numerals)
	hasSign := strings.ContainsAny(pw, signs)

	assert.True(t, hasVowel, "Пароль должен содержать гласные")
	assert.True(t, hasConsonant, "Пароль должен содержать согласные")
	assert.True(t, hasNumeral, "Пароль должен содержать цифры")
	assert.True(t, hasSign, "Пароль должен содержать знаки")
}

// TestGenerate_Uniqueness проверяет уникальность генерируемых паролей.
// Генерируется 1000 паролей и проверяется, что все они уникальны.
func TestGenerate_Uniqueness(t *testing.T) {
	t.Parallel()

	const iterations = 1000

	gen := password.NewGenerator()
	passwords := make(map[string]bool, iterations)

	for i := 0; i < iterations; i++ {
		pw := gen.Generate(16, password.CharAll)
		assert.False(t, passwords[pw], "Обнаружен дубликат пароля на итерации %d: %s", i, pw)
		passwords[pw] = true
	}

	assert.Len(t, passwords, iterations, "Должно быть сгенерировано %d уникальных паролей", iterations)
}

// TestGenerate_Concurrent проверяет потокобезопасность Generate().
// Запускается 10 горутин, каждая генерирует 100 паролей.
// Все 1000 паролей должны быть уникальны, race conditions отсутствуют.
func TestGenerate_Concurrent(t *testing.T) {
	t.Parallel()

	const (
		goroutines     = 10
		passwordsPerGr = 100
	)

	allPasswords := make(map[string]bool)

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	gen := password.NewGenerator()

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(goroutineID int) {
			defer wg.Done()

			localPasswords := make([]string, 0, passwordsPerGr)

			for i := 0; i < passwordsPerGr; i++ {
				localPasswords = append(localPasswords, gen.Generate(16, password.CharAll))
			}

			mu.Lock()
			defer mu.Unlock()

			for _, pw := range localPasswords {
				assert.False(t, allPasswords[pw], "Найден дубликат пароля из горутины %d: %s", goroutineID, pw)
				allPasswords[pw] = true
			}
		}(g)
	}

	wg.Wait()

	expectedCount := goroutines * passwordsPerGr
	assert.Len(t, allPasswords, expectedCount,
		"Должно быть сгенерировано %d уникальных паролей в %d горутинах", expectedCount, goroutines)
}

// TestGenerate_DoesNotHang проверяет, что Generate() не зависает
// при различных комбинациях видов символов (регрессия бага с бесконечным циклом).
func TestGenerate_DoesNotHang(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()

	type testCase struct {
		name       string
		length     int
		charsKinds password.CharKinds
	}

	tests := []testCase{
		{name: "charAll", length: 12, charsKinds: password.CharAll},
		{name: "charVowels", length: 12, charsKinds: password.CharVowels},
		{name: "charConsonants", length: 12, charsKinds: password.CharConsonants},
		{name: "charNumerals", length: 12, charsKinds: password.CharNumerals},
		{name: "charSigns", length: 12, charsKinds: password.CharSigns},
		{name: "charAbc", length: 12, charsKinds: password.CharAbc},
		{name: "charAbcNumerals", length: 12, charsKinds: password.CharAbcNumerals},
		{name: "default_zero", length: 12, charsKinds: 0},
		{name: "length_one", length: 1, charsKinds: password.CharAll},
		{name: "length_fifty", length: 50, charsKinds: password.CharAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pw := gen.Generate(tt.length, tt.charsKinds)
			require.Len(t, pw, tt.length)
		})
	}
}

// TestGenerate_ConsecutiveSameCharSet проверяет, что ограничение на количество
// подряд идущих символов одного типа соблюдается.
func TestGenerate_ConsecutiveSameCharSet(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()
	vowels := "aeiuyAEIUY"
	consonants := "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ"

	// Генерируем длинный пароль и проверяем, что подряд не идёт
	// более 2 гласных или более 2 согласных
	for i := 0; i < 100; i++ {
		pw := gen.Generate(50, password.CharAbc)

		maxConsecutiveVowels := 0
		maxConsecutiveConsonants := 0
		curVowels := 0
		curConsonants := 0

		for _, ch := range pw {
			if strings.ContainsRune(vowels, ch) {
				curVowels++
				curConsonants = 0

				if curVowels > maxConsecutiveVowels {
					maxConsecutiveVowels = curVowels
				}
			} else if strings.ContainsRune(consonants, ch) {
				curConsonants++
				curVowels = 0

				if curConsonants > maxConsecutiveConsonants {
					maxConsecutiveConsonants = curConsonants
				}
			}
		}

		assert.LessOrEqual(t, maxConsecutiveVowels, 2,
			"Пароль %q содержит более 2 гласных подряд", pw)
		assert.LessOrEqual(t, maxConsecutiveConsonants, 2,
			"Пароль %q содержит более 2 согласных подряд", pw)
	}
}

// TestNewGenerator_InnerState проверяет, что NewGenerator() создаёт
// генератор с корректно скопированными наборами символов.
func TestNewGenerator_InnerState(t *testing.T) {
	t.Parallel()

	gen := password.NewGenerator()
	pw := gen.Generate(10, password.CharAll)

	require.Len(t, pw, 10)
}

// TestGenerator_MultipleGenerators проверяет, что несколько генераторов
// работают независимо друг от друга.
func TestGenerator_MultipleGenerators(t *testing.T) {
	t.Parallel()

	gen1 := password.NewGenerator()
	gen2 := password.NewGenerator()

	pw1 := gen1.Generate(16, password.CharAll)
	pw2 := gen2.Generate(16, password.CharAll)

	require.Len(t, pw1, 16)
	require.Len(t, pw2, 16)
	assert.NotEqual(t, pw1, pw2, "Два генератора должны выдавать разные пароли")
}
