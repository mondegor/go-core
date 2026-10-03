package password

import (
	"math"
)

// Сложность пароля.
const (
	PassStrengthNotRated PassStrength = iota // пароль без оценки
	PassStrengthWeak                         // слабый пароль
	PassStrengthMedium                       // средний пароль
	PassStrengthStrong                       // надёжный пароль
	PassStrengthBest                         // максимально надёжный пароль
)

const (
	passTypeSmallABC = iota
	passTypeBigABC
	passTypeNumeral
	passTypeSign
)

// Размеры групп символов для расчёта энтропии: совпадают с наборами генератора (pwCharSets),
// поэтому менять их нужно вместе с ними. Для пароля, придуманного пользователем, оценка
// получается чуть консервативнее: символы, исключённые из генератора, размер групп не увеличивают.
const (
	passTypeNumeralLen  = 9
	passTypeBigABCLen   = 25
	passTypeSmallABCLen = 25
	passTypeSignLen     = 27
)

type (
	// PassStrength - сложность (надёжность) пароля.
	PassStrength uint8
)

// TODO: заменить map на строку.
var passStrengthName = map[PassStrength]string{ //nolint:gochecknoglobals
	PassStrengthNotRated: "NOT_RATED",
	PassStrengthWeak:     "WEAK",
	PassStrengthMedium:   "MIDDLE",
	PassStrengthStrong:   "STRONG",
	PassStrengthBest:     "THE_BEST",
}

// CalcStrength - вычисляет сложность (надёжность) пароля на основе:
// - количества уникальных символов;
// - разнообразия наборов символов (цифры, буквы, знаки);
// - информационной энтропии.
// Возвращает PassStrengthNotRated для пустого пароля.
func CalcStrength(value string) PassStrength {
	length := len(value)

	if length == 0 {
		return PassStrengthNotRated
	}

	uniqChars := make(map[byte]bool, length)
	uniqTypeChars := [4]int{}

	for i := 0; i < length; i++ {
		uniqChars[value[i]] = true

		switch {
		case value[i] >= 48 && value[i] <= 57:
			uniqTypeChars[passTypeNumeral] = passTypeNumeralLen
		case value[i] >= 65 && value[i] <= 90:
			uniqTypeChars[passTypeBigABC] = passTypeBigABCLen
		case value[i] >= 97 && value[i] <= 122:
			uniqTypeChars[passTypeSmallABC] = passTypeSmallABCLen
		default:
			uniqTypeChars[passTypeSign] = passTypeSignLen
		}
	}

	var (
		totalLen  int
		totalSets int
	)

	for i := range uniqTypeChars {
		if uniqTypeChars[i] == 0 {
			continue
		}

		totalLen += uniqTypeChars[i]
		totalSets++
	}

	if totalSets > 1 { // минимально два набора символов должно использоваться
		// вычисление информационной энтропии;
		// в комментариях к порогам указано, сколько уникальных символов гарантирует уровень
		// при любом сочетании N наборов (расчёт по самому узкому сочетанию), в скобках - биты
		bits := uint64(float64(len(uniqChars)) * math.Log2(float64(totalLen)))

		if bits >= 76 && totalSets > 3 { // 12 uniq chars and 4 sets[77]
			return PassStrengthBest
		}

		if bits >= 63 && totalSets > 2 { // 10 uniq chars and 4 sets[64] OR 11 uniq chars and 3 sets[64]
			return PassStrengthStrong
		}

		if bits >= 56 { // 9 uniq chars and 4 sets[57] OR 10 uniq chars and 3 sets[58] OR 12 uniq chars and 2 sets[61]
			return PassStrengthMedium
		}

		if bits >= 44 { // 7 uniq chars and 4 sets[44] OR 8 uniq chars and 3 sets[47] OR 9 uniq chars and 2 sets[45]
			return PassStrengthWeak
		}
	}

	return PassStrengthNotRated
}

// String - возвращает значение в виде строки.
func (e PassStrength) String() string {
	return passStrengthName[e]
}
