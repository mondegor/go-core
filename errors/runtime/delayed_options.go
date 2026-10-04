package runtime

import (
	"sync"

	"github.com/mondegor/go-core/errors/kind"
)

type (
	// OptionsHandler - обработчик формирования списка опций для указанного типа ошибки.
	// Позволяет динамически настраивать опции ошибок при их создании.
	OptionsHandler interface {
		Options(kindErr kind.Enum, message string) []Option
	}

	// OptionsHandlerFunc - адаптер, позволяющий использовать обычную функцию как OptionsHandler.
	OptionsHandlerFunc func(kindErr kind.Enum, message string) []Option
)

// Options - реализует интерфейс OptionsHandler, вызывая саму функцию f.
// Позволяет использовать обычную функцию как обработчик опций.
func (f OptionsHandlerFunc) Options(kindErr kind.Enum, message string) []Option {
	return f(kindErr, message)
}

var defopts = struct { //nolint:gochecknoglobals
	mu sync.Mutex

	// delayed - массив собирает все ошибки созданные через NewDelayed(), до момента вызова InitDelayedOptions().
	// Это позволяет, в момент запуска приложения, отследить создание всех таких ошибок в переменных
	// вида 'var Err*' для того чтобы проинициализировать их опциями по умолчанию
	// при запуске приложения с помощью метода InitDelayedOptions(). Далее все таки ошибки инициализируются
	// опциями по умолчанию сразу же при их создании, а этот массив более не используется.
	delayed []*protoError

	// defaultProtoOptions - глобальный обработчик формирования списка опций по умолчанию
	// для создаваемых прототипов ошибок (при вызове New()).
	handler OptionsHandler
}{}

// NewDelayed - создаёт прототип ошибки с отложенной инициализацией опций.
// Прототип запоминается до вызова InitDelayedOptions(), который применит к нему
// опции по умолчанию через указанный OptionsHandler.
// Если InitDelayedOptions() уже вызван, опции применяются немедленно.
// Используется для глобальных переменных ошибок (var Err...), которые объявляются до инициализации приложения.
func NewDelayed(errKind kind.Enum, text string, opts ...Option) ProtoError {
	p := newProto(errKind, text, opts)

	defopts.mu.Lock()
	defer defopts.mu.Unlock()

	// сначала происходит сбор создаваемых глобальных ProtoError ошибок в момент запуска приложения
	// чтобы их инициализировать нужными опциями, которые определяются приложением позже
	// это будет происходить до тех пор, пока не будет вызвана функция InitDelayedOptions()
	// далее опции по умолчанию применяются сразу
	if defopts.handler == nil {
		defopts.delayed = append(defopts.delayed, p)

		return p
	}

	// устанавливаются опции по умолчанию
	applyDefaultOptions(p, defopts.handler.Options(p.kind, p.text))

	return p
}

// InitDelayedOptions - одноразово инициализирует опции всем прототипам ошибок,
// созданным через NewDelayed() до этого вызова.
// Параметр handler - обработчик, формирующий опции по умолчанию для каждого прототипа.
//
// Явно переданные в NewDelayed() опции не перезаписываются.
// После вызова handler сохраняется и применяется ко всем последующим вызовам NewDelayed().
// Повторный вызов игнорируется (обработчик устанавливается только один раз).
// Если будет nil handler, устанавливается заглушка, блокирующая повторную инициализацию.
func InitDelayedOptions(handler OptionsHandler) {
	defopts.mu.Lock()
	defer defopts.mu.Unlock()

	if defopts.handler != nil {
		return
	}

	// если обработчик не указан, то устанавливается заглушка,
	// которая блокирует повторный вызов функции InitDelayedOptions()
	if handler == nil {
		defopts.handler = OptionsHandlerFunc(
			func(_ kind.Enum, _ string) []Option {
				return nil
			},
		)

		defopts.delayed = nil

		return
	}

	// устанавливаются опции по умолчанию для каждой созданной ошибки, в момент инициализации приложения,
	// но только если эти опции уже не были установлены в момент создания этих ошибок
	for _, p := range defopts.delayed {
		applyDefaultOptions(p, handler.Options(p.kind, p.text))
	}

	defopts.handler = handler
	defopts.delayed = nil
}

// applyDefaultOptions - применяет к прототипу опции по умолчанию, не перезаписывая
// значения, явно установленные при его создании. Опции применяются к временному
// прототипу, из которого переносятся только незаданные поля.
// При добавлении новой опции её поле необходимо добавить сюда.
func applyDefaultOptions(p *protoError, opts []Option) {
	if len(opts) == 0 {
		return
	}

	var defaults protoError

	for _, opt := range opts {
		opt(&defaults)
	}

	if p.onCreate == nil {
		p.onCreate = defaults.onCreate
	}
}
