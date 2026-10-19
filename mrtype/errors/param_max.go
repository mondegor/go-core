package errors

import (
	"fmt"
)

type (
	// ParamMaxValueError - ошибка превышения максимального допустимого значения параметра.
	ParamMaxValueError struct {
		Type     string
		MaxValue any
	}
)

// NewParamMaxValueError - создаёт ошибку ParamMaxValueError для указанного типа параметра.
// maxValue - максимально допустимое значение.
func NewParamMaxValueError(paramType string, maxValue any) error {
	return &ParamMaxValueError{
		Type:     paramType,
		MaxValue: maxValue,
	}
}

func (e *ParamMaxValueError) Error() string {
	return fmt.Sprintf("param contains value greater than max (type=%s, maxValue=%v)", e.Type, e.MaxValue)
}
