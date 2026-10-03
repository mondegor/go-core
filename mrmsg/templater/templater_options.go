package templater

type (
	// Option - функция для настройки объекта Templater.
	Option func(o *options)

	options struct {
		templater *Templater
	}
)

// WithMode - устанавливает режим рендеринга шаблона (по умолчанию ModeText).
func WithMode(value Mode) Option {
	return func(o *options) {
		o.templater.mode = value
	}
}
