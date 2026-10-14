package gotext

import (
	"strconv"
	"strings"

	"github.com/mondegor/go-core/mrmsg"
)

// NewMessageFormatter - создаёт MessageFormatter для преобразования
// плейсхолдеров в формат, совместимый с fmt.Sprintf.
// Параметры leftDelim и rightDelim - ограничители плейсхолдеров (например: "{{" и "}}").
// Плейсхолдеры преобразуются в формат "%[N]s", где N - порядковый номер аргумента.
// Литеральный символ '%' в остальном тексте не экранируется, это делает MessageConverter.
func NewMessageFormatter(leftDelim, rightDelim string) *mrmsg.MessageFormatter {
	return mrmsg.NewMessageFormatter(
		leftDelim,
		rightDelim,
		func(_ string, i int) (newPlaceholder string) {
			return "%[" + strconv.Itoa(i+1) + "]s"
		},
	)
}

// MessageConverter - возвращает функцию для предварительной обработки сообщения.
// Преобразует плейсхолдеры с указанными ограничителями в формат fmt.Sprintf ("%[N]s"),
// а литеральный символ '%' экранирует как "%%" (например: "100% done {Name}" → "100%% done %[1]s").
// Ограничители не должны содержать символ '%', иначе плейсхолдеры не будут распознаны.
func MessageConverter(leftDelim, rightDelim string) func(message string, args []any) (newMsg string, newArgs []any) {
	formatter := NewMessageFormatter(leftDelim, rightDelim)

	return func(message string, args []any) (newMsg string, newArgs []any) {
		newMsg, _ = formatter.Format(strings.ReplaceAll(message, "%", "%%"))

		return newMsg, args
	}
}
