package gotext

import (
	"golang.org/x/text/message/catalog"
)

type (
	// Option - функция-опция для настройки Provider.
	Option func(o *options)

	options struct {
		domains  []string
		catalogs []catalog.Catalog
	}
)

// WithDomainCatalog - добавляет каталог переводов для указанного домена.
// Параметры:
//   - domain - имя домена (например: "messages" или "errors");
//   - cat - каталог с переводами типа catalog.Catalog из golang.org/x/text/message/catalog.
//
// Если cat равен nil, опция игнорируется. Повторный домен заменяет каталог, заданный ранее.
func WithDomainCatalog(domain string, cat catalog.Catalog) Option {
	return func(o *options) {
		if cat == nil {
			return
		}

		o.domains = append(o.domains, domain)
		o.catalogs = append(o.catalogs, cat)
	}
}
