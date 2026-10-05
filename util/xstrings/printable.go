package xstrings

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SanitizePrintable - приводит недоверенную строку к безопасному для хранения и вывода виду:
// удаляет невалидный UTF-8, непечатаемые символы (управляющие и форматирующие Cf, включая
// bidi-override и ZWJ) и комбинируемые знаки в начале, схлопывает пробельные символы
// в один пробел и обрезает края. Длина ограничивается maxLen рунами (maxLen < 1 - без ограничения),
// удалённые символы в неё не входят, поэтому длину входа ограничивает вызывающий код.
func SanitizePrintable(s string, maxLen int) string {
	// под ASCII-результат хватает одной аллокации, для многобайтных рун буфер дорастёт
	capacity := len(s)
	if maxLen > 0 {
		capacity = min(capacity, maxLen)
	}

	var buf strings.Builder

	buf.Grow(capacity)

	count := 0
	pendingSpace := false

	for len(s) > 0 && (maxLen < 1 || count < maxLen) {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]

		// невалидный байт декодируется как RuneError размера 1
		if r == utf8.RuneError && size == 1 {
			continue
		}

		// пробел записывается только перед следующим символом, поэтому ведущие, повторные
		// (в том числе разделённые удалёнными символами) и концевые пробелы не попадают в результат
		if unicode.IsSpace(r) {
			pendingSpace = count > 0

			continue
		}

		if !unicode.IsPrint(r) {
			continue
		}

		// комбинируемый знак без базового символа
		if count == 0 && unicode.In(r, unicode.Mn, unicode.Me) {
			continue
		}

		if pendingSpace {
			// для пробела и символа после него нужно два места
			if maxLen > 0 && count+1 >= maxLen {
				break
			}

			buf.WriteByte(' ')

			count++
			pendingSpace = false
		}

		buf.WriteRune(r)

		count++
	}

	return buf.String()
}
