package mrlocale

type (
	// BundleOption - функция-опция для настройки Bundle.
	BundleOption func(o *bundleOptions)

	bundleOptions struct {
		formatMessage   func(msg string, args []any) (newMsg string, newArgs []any)
		formatError     func(err error) (msg string, args []any)
		messagesDomain  string
		errorsDomain    string
		defaultLanguage string
	}
)

// WithFormatMessage - задаёт функцию предварительной обработки сообщения и аргументов
// перед передачей провайдеру, в том числе сообщений ошибок (по умолчанию - без изменений).
func WithFormatMessage(fn func(msg string, args []any) (newMsg string, newArgs []any)) BundleOption {
	return func(o *bundleOptions) {
		o.formatMessage = fn
	}
}

// WithFormatError - задаёт функцию извлечения сообщения и аргументов из ошибки для
// Localizer.TranslateError (по умолчанию - err.Error() без аргументов).
func WithFormatError(fn func(err error) (msg string, args []any)) BundleOption {
	return func(o *bundleOptions) {
		o.formatError = fn
	}
}

// WithMessagesDomain - задаёт домен для локализации обычных сообщений (по умолчанию "messages").
func WithMessagesDomain(value string) BundleOption {
	return func(o *bundleOptions) {
		o.messagesDomain = value
	}
}

// WithErrorsDomain - задаёт домен для локализации сообщений об ошибках (по умолчанию "errors").
func WithErrorsDomain(value string) BundleOption {
	return func(o *bundleOptions) {
		o.errorsDomain = value
	}
}

// WithDefaultLanguage - задаёт язык по умолчанию.
// Если указанный язык не найден в списке поддерживаемых,
// будет возвращена ошибка при создании Bundle.
//
// Язык сверяется со списком поддерживаемых по разобранному тегу, поэтому его запись
// не обязана совпадать с записью в списке ("en-US", "en_US" и "en-us" - один язык).
// При этом язык без региона не равен языку с регионом: "en" не подойдёт для "en-US".
func WithDefaultLanguage(value string) BundleOption {
	return func(o *bundleOptions) {
		o.defaultLanguage = value
	}
}
