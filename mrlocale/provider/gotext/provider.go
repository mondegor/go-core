package gotext

import (
	"errors"
	"maps"
	"slices"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type (
	// Provider - провайдер локализации, построенный на golang.org/x/text/message.
	// Поддерживает несколько доменов и языков, используя каталоги переводов.
	Provider struct {
		defaultPrinter     *message.Printer
		domains            []string
		domain2langCatalog map[string]map[language.Tag]*message.Printer
	}
)

// NewProvider - создаёт Provider локализации сообщений.
// Параметры:
//   - languages - список языков для которых будет настроена локализация;
//   - opts - опции для добавления каталогов переводов через WithDomainCatalog;
//
// Первый язык в списке становится языком по умолчанию.
func NewProvider(
	languages []language.Tag,
	opts ...Option,
) (*Provider, error) {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	if len(languages) == 0 {
		return nil, errors.New("gotext provider:languages is required")
	}

	p := &Provider{
		defaultPrinter: message.NewPrinter(languages[0]),
	}

	if len(o.domains) == 0 {
		return p, nil
	}

	p.domain2langCatalog = make(map[string]map[language.Tag]*message.Printer, len(o.domains))

	for i, domain := range o.domains {
		p.domain2langCatalog[domain] = make(map[language.Tag]*message.Printer, len(languages))

		for _, lang := range languages {
			p.domain2langCatalog[domain][lang] = message.NewPrinter(lang, message.Catalog(o.catalogs[i]))
		}
	}

	// uniq domains
	p.domains = slices.Collect(maps.Keys(p.domain2langCatalog))

	return p, nil
}

// Domains - возвращает список доступных доменов для локализации.
func (p *Provider) Domains() []string {
	return slices.Clone(p.domains)
}

// Localize - возвращает локализованное сообщение с подставленными аргументами.
// Если для указанной пары domain/lang не найден каталог, используется язык по умолчанию.
func (p *Provider) Localize(domain string, lang language.Tag, msg string, args []any) string {
	printer := p.defaultPrinter

	if pr, ok := p.domain2langCatalog[domain][lang]; ok {
		printer = pr
	}

	return printer.Sprintf(msg, args...)
}
