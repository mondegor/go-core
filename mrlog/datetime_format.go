package mrlog

import (
	"fmt"
	"time"
)

// ParseDateTimeFormat - преобразует имя формата времени в константу time.
// Поддерживаемые значения: "RFC3339", "RFC3339Nano", "DateTime", "TimeOnly", "Kitchen".
// Возвращает строку формата для time.Time.Format при успехе,
// или time.RFC3339 и ошибку при нераспознанном значении.
//
// ВАЖНО: функция является хелпером логгера. Это не утилита общего
// назначения: для прикладного кода её использовать не следует.
func ParseDateTimeFormat(value string) (string, error) {
	switch value {
	case "RFC3339":
		return time.RFC3339, nil
	case "RFC3339Nano":
		return time.RFC3339Nano, nil
	case "DateTime":
		return time.DateTime, nil
	case "TimeOnly":
		return time.TimeOnly, nil
	case "Kitchen":
		return time.Kitchen, nil
	}

	return time.RFC3339, fmt.Errorf("the value is not a datetime format value (value=%q)", value)
}
