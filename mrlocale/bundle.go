package mrlocale

import (
	"fmt"
	"slices"

	"golang.org/x/text/language"
)

// Предопределённые названия доменов по умолчанию.
const (
	DefaultMessagesDomain = "messages"
	DefaultErrorsDomain   = "errors"
)

const (
	fallbackLanguage = "en-US"
)

type (
	// Bundle - управляет локализацией сообщений, ошибок и справочников.
	// Загружает переводы из провайдера, выбирает подходящий язык
	// и форматирует сообщения с учётом переданных аргументов.
	Bundle struct {
		provider       MessageProvider
		messagesDomain string
		errorsDomain   string
		formatMessage  func(msg string, args []any) (newMsg string, newArgs []any)
		formatError    func(err error) (msg string, args []any)
		// languages - языки в том же порядке, в котором они переданы
		// в languageMatcher: индекс, возвращаемый Match, адресует именно этот срез,
		// и по нему же адресуются локализаторы пула (см. NewPool)
		languages       []language.Tag
		languageMatcher language.Matcher
		defaultLanguage language.Tag
	}
)

// NewBundle - создаёт Bundle для локализации сообщений.
// Параметры:
//   - languages - список поддерживаемых языков в формате BCP 47 (например: "en", "ru", "en-US", "en_US");
//   - opts - дополнительные опции настройки (WithMessageProvider, WithDefaultLanguage и др.);
//
// Первый язык в списке становится языком по умолчанию, если не задан явно через WithDefaultLanguage.
//
// Пустой список отказом не является: вместо него подставляется "en-US", и он становится
// единственным языком бандла и его языком по умолчанию. Опция WithDefaultLanguage при этом
// годится только для того же "en-US" (запись любая: "en_US", "en-us"), а любой другой язык
// отвергается обычной сверкой со списком.
//
// Запись языка принимается любая разбираемая, в т.ч. неканоничная ("en_US", "ru-ru"),
// но наружу пул отдаёт уже канон ("en-US", "ru-RU"), см. Pool.Languages. Поэтому список
// из конфигурации годится для создания бандла, но не годится ни как источник кодов
// для клиента, ни для сверки присланного кода: исходная запись не совпадёт ни с одним
// кодом, который приложение реально использует. Чтобы расхождение записи не дошло
// до приложения, список из конфигурации проверяется config.ValidateLanguages -
// это единственное место, где неканоничная запись отвергается.
func NewBundle(languages []string, opts ...BundleOption) (*Bundle, error) {
	o := bundleOptions{
		messagesDomain: DefaultMessagesDomain,
		errorsDomain:   DefaultErrorsDomain,
	}

	for _, opt := range opts {
		opt(&o)
	}

	if len(languages) == 0 {
		languages = []string{fallbackLanguage}
	}

	languageTags := make([]language.Tag, len(languages))

	for i, lang := range languages {
		tag, err := language.Parse(lang)
		if err != nil {
			return nil, fmt.Errorf("bundle create: parsing language '%s': %w", lang, err)
		}

		// уникальность проверяется по разобранному тегу, а не по строке: одна и та же
		// локаль записывается по-разному ("ru-RU", "ru_RU"), и дубликат в списке
		// породил бы лишний локализатор и неоднозначность подбора языка.
		// Сравнение тегов через == корректно: language.Parse приводит тег
		// к канонической форме, поэтому разные записи одной локали равны побитово
		if slices.Contains(languageTags[:i], tag) {
			return nil, fmt.Errorf("bundle create: duplicate language '%s'", lang)
		}

		languageTags[i] = tag
	}

	// если в опции явно не указан язык по умолчанию,
	// то по умолчанию используется первый язык в списке
	defaultLanguage := languageTags[0]

	if o.defaultLanguage != "" {
		tag, err := language.Parse(o.defaultLanguage)
		if err != nil {
			return nil, fmt.Errorf("bundle create: parsing default language '%s': %w", o.defaultLanguage, err)
		}

		// языки сверяются разобранными тегами, а не исходными строками: одна и та же
		// локаль записывается по-разному ("ru-RU", "ru_RU", "ru-ru"), и сравнение строк
		// отвергало бы язык по умолчанию, записанный не так, как он записан в списке
		if !slices.Contains(languageTags, tag) {
			return nil, fmt.Errorf("bundle create: default language is not supported (lang='%s')", o.defaultLanguage)
		}

		defaultLanguage = tag
	}

	pr, err := o.createProvider(languageTags)
	if err != nil {
		return nil, fmt.Errorf("bundle create: %w", err)
	}

	if !isDomainInArray(o.messagesDomain, pr.Domains()) {
		return nil, fmt.Errorf("bundle create: provider has not messages domain '%s'", o.messagesDomain)
	}

	if !isDomainInArray(o.errorsDomain, pr.Domains()) {
		return nil, fmt.Errorf("bundle create: provider has not errors domain '%s'", o.errorsDomain)
	}

	if o.formatMessage == nil {
		o.formatMessage = func(msg string, args []any) (newMsg string, newArgs []any) {
			return msg, args
		}
	}

	if o.formatError == nil {
		o.formatError = func(err error) (string, []any) {
			return err.Error(), nil
		}
	}

	return &Bundle{
		provider:        pr,
		messagesDomain:  o.messagesDomain,
		errorsDomain:    o.errorsDomain,
		formatMessage:   o.formatMessage,
		formatError:     o.formatError,
		languages:       languageTags,
		languageMatcher: language.NewMatcher(languageTags),
		defaultLanguage: defaultLanguage,
	}, nil
}

// localize - выполняет локализацию сообщения в указанном домене.
// Предварительно обрабатывает сообщение и аргументы через formatMessage.
func (b *Bundle) localize(domain string, lang language.Tag, msg string, args []any) string {
	msg, args = b.formatMessage(msg, args)

	return b.provider.Localize(domain, lang, msg, args)
}

// isDomainInArray - проверяет наличие домена в массиве доменов.
func isDomainInArray(domain string, domains []string) bool {
	for _, val := range domains {
		if val == domain {
			return true
		}
	}

	return false
}
