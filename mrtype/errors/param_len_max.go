package errors

import (
	"fmt"
)

type (
	// ParamLenMaxError - ошибка превышения максимальной длины значения параметра.
	ParamLenMaxError struct {
		Type   string
		MaxLen int
	}
)

// NewParamLenMaxError - создаёт ошибку ParamLenMaxError для указанного типа параметра.
// maxLen - максимально допустимая длина значения.
func NewParamLenMaxError(paramType string, maxLen int) error {
	return &ParamLenMaxError{
		Type:   paramType,
		MaxLen: maxLen,
	}
}

func (e *ParamLenMaxError) Error() string {
	return fmt.Sprintf("param has value length greater than max characters (type=%s, maxLen=%d)", e.Type, e.MaxLen)
}
