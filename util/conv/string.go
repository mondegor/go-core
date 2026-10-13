package conv

import (
	"fmt"
	"strconv"
)

// String - преобразует значение произвольного типа в строку.
// Поддерживает: строки, целые числа, числа с плавающей точкой, bool, nil ("<NIL>"), error,
// fmt.Stringer и любые другие значения через fmt.Sprintf.
func String(value any) string {
	str, _ := toString(value)

	return str
}

func toString(value any) (string, bool) {
	switch val := value.(type) {
	case string:
		return val, true
	case int:
		return strconv.FormatInt(int64(val), 10), false
	case uint:
		return strconv.FormatUint(uint64(val), 10), false
	case int64:
		return strconv.FormatInt(val, 10), false
	case uint64:
		return strconv.FormatUint(val, 10), false
	case bool:
		return strconv.FormatBool(val), false
	case int8:
		return strconv.FormatInt(int64(val), 10), false
	case uint8:
		return strconv.FormatUint(uint64(val), 10), false
	case int16:
		return strconv.FormatInt(int64(val), 10), false
	case uint16:
		return strconv.FormatUint(uint64(val), 10), false
	case int32:
		return strconv.FormatInt(int64(val), 10), false
	case uint32:
		return strconv.FormatUint(uint64(val), 10), false
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), false
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), false
	case nil:
		return nilValue, false
	case error:
		return val.Error(), true
	case fmt.Stringer:
		return val.String(), true
	default:
		return fmt.Sprintf("%v", val), true
	}
}
