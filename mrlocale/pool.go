package mrlocale

import (
	"slices"

	"golang.org/x/text/language"
)

type (
	// Pool - управляет набором Localizer для всех поддерживаемых языков.
	// Создаётся из Bundle и предоставляет удобный поиск локализатора
	// по близости к указанному языку через language.Matcher.
	Pool struct {
		bundle           *Bundle
		localizers       []*Localizer
		localizerByCode  map[string]*Localizer
		languages        []string
		defaultLocalizer *Localizer
	}
)

// NewPool - создаёт Pool из Bundle.
// Создаёт Localizer для каждого языка, поддерживаемого бандлом.
//
// Локализаторы хранятся срезом в том же порядке, что и Bundle.languages,
// чтобы индекс, возвращаемый language.Matcher.Match, адресовал их напрямую:
// матчер построен на том же срезе, поэтому индекс и позиция локализатора
// заведомо совпадают, и промежуточный поиск по тегу не нужен.
//
// Параллельно срезу собирается индекс локализаторов по коду: он обслуживает точный
// подбор (LocalizerByCode), которому матчер не нужен, и список кодов (см. Languages).
func NewPool(bundle *Bundle) *Pool {
	localizers := make([]*Localizer, len(bundle.languages))
	localizerByCode := make(map[string]*Localizer, len(bundle.languages))
	languages := make([]string, len(bundle.languages))
	defaultLocalizer := &Localizer{
		bundle:   bundle,
		language: bundle.defaultLanguage,
	}

	for i, lang := range bundle.languages {
		localizer := defaultLocalizer

		if lang != bundle.defaultLanguage {
			localizer = &Localizer{
				bundle:   bundle,
				language: lang,
			}
		}

		// код берётся у тега списка, а не у языка локализатора, чтобы ключ
		// localizerByCode отвечал позиции i независимо от того, попал в неё
		// общий локализатор по умолчанию или собственный. Дубликаты тегов отвергает
		// NewBundle, поэтому код тега и Localizer.Language() здесь заведомо совпадают
		code := lang.String()

		localizers[i] = localizer
		localizerByCode[code] = localizer
		languages[i] = code
	}

	return &Pool{
		bundle:           bundle,
		localizers:       localizers,
		localizerByCode:  localizerByCode,
		languages:        languages,
		defaultLocalizer: defaultLocalizer,
	}
}

// Languages - возвращает коды всех языков пула в том порядке, в котором они переданы
// в бандл. Коды канонические - ровно те, которые принимает LocalizerByCode и отдаёт
// Localizer.Language(), поэтому неканоничная запись из конфигурации ("en_US") здесь
// уже приведена к каноничной ("en-US").
//
// Пустым список не бывает: бандл без языков не создаётся (см. NewBundle).
func (p *Pool) Languages() []string {
	return slices.Clone(p.languages)
}

// LocalizerByCode - возвращает Localizer языка, код которого точно совпадает
// с указанным, и сообщает false, если такого языка нет.
//
// В отличие от Localizer подбор ближайшего языка не выполняется: код сверяется
// со списком как есть, поэтому "ru" при списке ["ru-RU"] совпадением не считается.
// Предназначен для источников, значения которых клиент получает от самого приложения:
// там промах означает ошибку клиента, а не повод подобрать похожий язык.
//
// При промахе возвращается nil, а не язык по умолчанию: флаг здесь обязателен к разбору.
// Вызывающему, которому подходит любой язык, предназначен Localizer - он подбирает язык
// сам и промахов не имеет, поэтому подмена в обоих методах сразу лишила бы приложение
// способа отличить неизвестный код от известного.
func (p *Pool) LocalizerByCode(code string) (*Localizer, bool) {
	localizer, ok := p.localizerByCode[code]

	return localizer, ok
}

// Localizer - возвращает Localizer для наиболее близкого языка из списка langs.
// Использует language.Matcher для поиска лучшего совпадения.
// Если совпадение не найдено, возвращает Localizer для языка по умолчанию.
// Если langs пуст, также возвращает Localizer по умолчанию.
//
// Язык выбирается по индексу, который возвращает Match, а не по тегу: для тегов
// с регионом Match возвращает тег с расширением (например, "en-US" -> "en-u-rg-uszzzz"),
// поэтому поиск по тегу промахивался бы мимо любого языка списка.
//
// Индекс на длину локализаторов не проверяется: матчер построен на Bundle.languages,
// из которого NewPool создаёт локализаторы, поэтому индекс заведомо в границах среза.
func (p *Pool) Localizer(langs ...language.Tag) *Localizer {
	if len(langs) > 0 {
		_, index, conf := p.bundle.languageMatcher.Match(langs...)

		// при conf == language.No матчер возвращает index == 0, то есть первый язык
		// списка, а он не обязан быть языком по умолчанию, поэтому промах отсекается
		// отдельно, а не разбирается вместе с удачными подборами
		if conf != language.No {
			return p.localizers[index]
		}
	}

	return p.defaultLocalizer
}
