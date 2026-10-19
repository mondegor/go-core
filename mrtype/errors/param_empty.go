package errors

import (
	"fmt"
)

type (
	// ParamEmptyError - ошибка пустого значения параметра.
	// Возникает, когда обязательный параметр не передан.
	ParamEmptyError struct {
		Type string
	}
)

// NewParamEmptyError - создаёт ошибку ParamEmptyError для указанного типа параметра.
func NewParamEmptyError(paramType string) error {
	return &ParamEmptyError{
		Type: paramType,
	}
}

func (e *ParamEmptyError) Error() string {
	return fmt.Sprintf("param value is empty (type=%s)", e.Type)
}
