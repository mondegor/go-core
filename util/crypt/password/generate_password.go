package password

import (
	"crypto/rand"
	"math"
)

// Символы используемые в пароле.
const (
	CharVowels      CharKinds = 1  // гласные буквы
	CharConsonants  CharKinds = 2  // согласные буквы
	CharNumerals    CharKinds = 4  // цифры
	CharSigns       CharKinds = 8  // знаки
	CharAbc         CharKinds = 3  // CharVowels + CharConsonants
	CharAbcNumerals CharKinds = 7  // CharVowels + CharConsonants + CharNumerals
	CharAll         CharKinds = 15 // CharVowels + CharConsonants + CharNumerals + CharSigns
)

var pwCharSets = [...]pwCharSet{ //nolint:gochecknoglobals
	{CharVowels, 2, true, 10, []byte("aeiuyAEIUY")}, // oO - символы удалены, чтобы не перепутать с нулём
	{CharConsonants, 2, true, 40, []byte("bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ")},
	{CharNumerals, 1, false, 9, []byte("123456789")},                 // 0 - символ удалён, чтобы не перепутать с символами oO
	{CharSigns, 1, false, 27, []byte("!#$%&()*+,-./:;<=>?@[]^_{}~")}, // "'`\| - символы удалены
}

type (
	// CharKinds - вид символов используемых в пароле.
	CharKinds uint8

	pwCharSet struct {
		kind            CharKinds
		successivelyMax int
		firstOrLast     bool
		lettersLen      uint8
		letters         []byte
	}

	// Generator - генератор паролей с настраиваемыми наборами символов.
	// Позволяет создавать пароли с ограничениями на последовательные символы.
	Generator struct {
		pwCharSets []pwCharSet
	}
)

// NewGenerator - создаёт генератор паролей с предустановленными наборами символов.
func NewGenerator() *Generator {
	return &Generator{
		pwCharSets: pwCharSets[:],
	}
}

// Generate - генерирует пароль указанной длины с заданным набором символов.
// Параметр charsKinds определяет используемые символы: гласные, согласные, цифры, знаки.
// Ограничивает количество последовательных символов одного типа для повышения надёжности.
func (pg *Generator) Generate(length int, charsKinds CharKinds) string {
	if length < 1 {
		length = 1
	}

	if charsKinds == 0 || charsKinds > CharAll {
		charsKinds = CharAll
	}

	abc := make([]pwCharSet, 0, len(pg.pwCharSets))

	for i := 0; i < len(pg.pwCharSets); i++ {
		if (pg.pwCharSets[i].kind & charsKinds) > 0 {
			abc = append(abc, pg.pwCharSets[i])
		}
	}

	// если указан только один набор символов
	if len(abc) == 1 {
		abc[0].successivelyMax = length // максимальная длина совпадает с длиной пароля
		abc[0].firstOrLast = true       // первый и последний символ не проверяется
	}

	result := make([]byte, length)

	lastAbc := struct {
		charSetIndex           uint8
		countSuccessivelySigns int
	}{}

	for i := 0; i < length; i++ {
		var abcIndex uint8

		for {
			abcIndex = pg.getRandValue(uint8(len(abc))) //nolint:gosec // abc никогда не превзойдёт 255

			// если выбранный тип можно использовать для генерации первого и последнего символа
			// или если символ не первый и не последний
			if abc[abcIndex].firstOrLast || (i != 0 && i != (length-1)) {
				// если предыдущий символ такого же типа
				if abcIndex != lastAbc.charSetIndex {
					lastAbc.charSetIndex = abcIndex
					lastAbc.countSuccessivelySigns = 1

					break
				}

				// если подряд идущих символов не превышает
				if lastAbc.countSuccessivelySigns < abc[abcIndex].successivelyMax {
					lastAbc.countSuccessivelySigns++

					break
				}
			}
		}

		// обращение к случайному символу типа
		result[i] = abc[abcIndex].letters[pg.getRandValue(abc[abcIndex].lettersLen)]
	}

	return string(result)
}

func (pg *Generator) getRandValue(maxValue uint8) uint8 {
	var tmp [1]byte

	bits100 := uint64(math.Log2(float64(maxValue)) * 100)
	bits := bits100 / 100

	if bits100%100 != 0 {
		bits++
	}

	mask := uint8(1<<bits) - 1

	for {
		if _, err := rand.Read(tmp[:]); err != nil {
			return 0
		}

		rnd := tmp[0] & mask

		if rnd < maxValue {
			return rnd
		}
	}
}
