package mrlocale

import (
	"golang.org/x/text/language"
)

type (
	// Localizer - предоставляет методы для локализации сообщений на определённом языке.
	// Получается из Pool для конкретного языка или языка по умолчанию.
	Localizer struct {
		bundle   *Bundle
		language language.Tag
	}
)

// Language - возвращает код текущего языка локализации в формате BCP 47 (например: "en", "ru").
func (l *Localizer) Language() string {
	return l.language.String()
}

// Translate - локализует сообщение из домена сообщений (messages).
// Параметры:
//   - msg - ключ или исходный текст сообщения;
//   - args - аргументы для подстановки в сообщение.
func (l *Localizer) Translate(msg string, args ...any) string {
	return l.bundle.localize(l.bundle.messagesDomain, l.language, msg, args)
}

// TranslateError - локализует сообщение об ошибке из домена ошибок (errors).
// Текст и аргументы извлекаются из ошибки через formatError (см. WithFormatError),
// затем обрабатываются как обычное сообщение (см. WithFormatMessage).
func (l *Localizer) TranslateError(err error) string {
	msg, args := l.bundle.formatError(err)

	return l.bundle.localize(l.bundle.errorsDomain, l.language, msg, args)
}

// TranslateInDomain - локализует сообщение в указанном домене.
// Параметры:
//   - domain - домен для поиска перевода;
//   - msg - ключ или исходный текст сообщения;
//   - args - аргументы для подстановки в сообщение.
func (l *Localizer) TranslateInDomain(domain, msg string, args ...any) string {
	return l.bundle.localize(domain, l.language, msg, args)
}
