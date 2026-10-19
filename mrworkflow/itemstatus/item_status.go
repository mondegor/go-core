package itemstatus

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
)

// Возможные статусы элемента.
const (
	Draft    Enum = iota + 1 // черновик
	Enabled                  // действующий
	Disabled                 // отключённый
)

const (
	enumLast = uint8(Disabled)
	enumName = "ItemStatus"
)

type (
	// Enum - статус элемента (сущности с жизненным циклом).
	// Поддерживаемые статусы: Draft, Enabled, Disabled.
	Enum uint8
)

//nolint:gochecknoglobals
var (
	enumKeys = map[Enum]string{
		Draft:    "DRAFT",
		Enabled:  "ENABLED",
		Disabled: "DISABLED",
	}

	enumValues = map[string]Enum{
		"DRAFT":    Draft,
		"ENABLED":  Enabled,
		"DISABLED": Disabled,
	}
)

// Set - устанавливает указанное значение, если оно является enum значением.
func (e *Enum) Set(value uint8) error {
	if value > 0 && value <= enumLast {
		*e = Enum(value)

		return nil
	}

	return fmt.Errorf("value is not found in enum set (value=%d, enum=%s)", value, enumName)
}

// String - возвращает значение в виде строки.
func (e Enum) String() string {
	if v, ok := enumKeys[e]; ok {
		return v
	}

	return "UNKNOWN"
}

// MarshalJSON - переводит enum значение в строковое представление.
func (e Enum) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(e.String())
	if err != nil {
		return nil, fmt.Errorf("marshal error (enum=%s): %w", enumName, err)
	}

	return bytes, nil
}

// UnmarshalJSON - переводит строковое значение в enum представление.
func (e *Enum) UnmarshalJSON(data []byte) error {
	var value string

	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("unmarshal error (enum=%s): %w", enumName, err)
	}

	val, err := Parse(value)
	if err != nil {
		return err
	}

	*e = val

	return nil
}

// Scan implements the Scanner interface.
func (e *Enum) Scan(value any) error {
	val, ok := value.(int64)
	if !ok {
		return fmt.Errorf("invalid type assertion (type=%s, value=%#v)", enumName, value)
	}

	if val < 0 || val > math.MaxUint8 {
		return fmt.Errorf("value is out of range (type=%s, value=%d)", enumName, val)
	}

	return e.Set(uint8(val))
}

// Value implements the driver.Valuer interface.
func (e Enum) Value() (driver.Value, error) {
	return uint8(e), nil
}

// Parse - парсит указанное значение и если оно валидно, то устанавливает его числовое значение.
func Parse(value string) (Enum, error) {
	if parsedValue, ok := enumValues[value]; ok {
		return parsedValue, nil
	}

	return 0, fmt.Errorf("key is not found in enum set (enum=%s, key=%q)", enumName, value)
}

// ParseList - парсит массив строковых значений и
// возвращает соответствующий массив enum значений.
func ParseList(values []string) ([]Enum, error) {
	parsedValues := make([]Enum, len(values))

	for i := range values {
		val, err := Parse(values[i])
		if err != nil {
			return nil, err
		}

		parsedValues[i] = val
	}

	return parsedValues, nil
}
