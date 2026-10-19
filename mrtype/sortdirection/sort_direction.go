package sortdirection

import (
	"fmt"
)

// Возможные направления сортировки.
const (
	ASC  Enum = iota // по возрастанию
	DESC             // по убыванию
)

const (
	enumName = "SortDirection"
)

type (
	// Enum - перечисление направлений сортировки.
	Enum uint8
)

//nolint:gochecknoglobals
var (
	enumKeys = map[Enum]string{
		ASC:  "ASC",
		DESC: "DESC",
	}

	enumValues = map[string]Enum{
		"ASC":  ASC,
		"DESC": DESC,
	}
)

// String - возвращает значение в виде строки.
func (e Enum) String() string {
	if v, ok := enumKeys[e]; ok {
		return v
	}

	return "UNKNOWN"
}

// Parse - преобразует строку в значение направления сортировки.
// Поддерживаемые значения: "ASC", "DESC".
// Возвращает найденное значение при успехе, или ошибку при нераспознанном значении.
func Parse(value string) (Enum, error) {
	if parsedValue, ok := enumValues[value]; ok {
		return parsedValue, nil
	}

	return 0, fmt.Errorf("key is not found in enum set (enum=%s, key=%q)", enumName, value)
}
