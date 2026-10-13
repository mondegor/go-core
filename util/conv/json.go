package conv

import (
	"strconv"
)

const (
	nilValue = "<NIL>"
)

// JSONValue - преобразует значение в строковое представление для JSON.
// Строки, error, fmt.Stringer и прочие нечисловые значения оборачиваются в кавычки
// через strconv.Quote, nil преобразуется в null.
func JSONValue(value any) string {
	str, needQuote := toString(value)

	if needQuote {
		return strconv.Quote(str)
	}

	if str == nilValue {
		return "null"
	}

	return str
}
