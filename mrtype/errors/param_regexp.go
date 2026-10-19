package errors

import (
	"fmt"
)

type (
	// ParamRegexpError - ошибка несоответствия значения параметра регулярному выражению.
	ParamRegexpError struct {
		Type   string
		Regexp string
	}
)

// NewParamRegexpError - создаёт ошибку ParamRegexpError для указанного типа параметра.
// regexp - строка регулярного выражения, которому значение не соответствует.
func NewParamRegexpError(paramType, regexp string) error {
	return &ParamRegexpError{
		Type:   paramType,
		Regexp: regexp,
	}
}

func (e *ParamRegexpError) Error() string {
	return fmt.Sprintf("param value doesn't match regexp (type=%s, regexp=%q)", e.Type, e.Regexp)
}
