package config

import (
	"fmt"

	"golang.org/x/text/language"
)

const (
	// База неопределённого языка. С ней разбираются все записи, которые конкретный
	// язык не называют ("und", "und-RU", "x-private"), поэтому валидатором отвергаются.
	baseUndetermined = "und"
)

// ValidateLanguages - валидирует указанный список кодов языков приложения:
// каждый код должен разбираться, быть записан канонически, называть конкретный язык
// и встречаться в списке один раз. Проверка останавливается на первом негодном коде
// и сообщает о нём ошибкой.
//
// Пустой список ошибкой не считается: вместо него mrlocale.NewBundle подставляет "en-US",
// который становится единственным языком бандла и его языком по умолчанию
// (см. mrlocale.NewBundle).
//
// Каноничность требуется потому, что NewBundle принимает и неканоничную форму ("en_US",
// "ru-ru"), но наружу пул отдаёт уже канон (см. mrlocale.Pool.Languages): исходная запись
// после этого не совпадает ни с одним кодом, который приложение реально использует.
// Список из конфигурации в таком виде годится для NewBundle, но не годится ни как
// источник кодов для клиента, ни для сверки присланного кода.
//
// Неопределённый язык отвергается: language.Parse его принимает, NewBundle тоже,
// поэтому такой код дошёл бы до клиента как язык приложения, хотя переводов за ним нет.
// Язык определяется по базе тега (Tag.Raw), а не по Tag.Base: последний для "und"
// домысливает наиболее вероятный язык ("en") и потому неопределённость не показывает.
func ValidateLanguages(names []string) error {
	uniqNames := make(map[string]bool, len(names))

	for _, name := range names {
		tag, err := language.Parse(name)
		if err != nil {
			return fmt.Errorf("error parsing language (name='%s'): %w", name, err)
		}

		if code := tag.String(); name != code {
			return fmt.Errorf("language '%s' is not canonical: expected '%s'", name, code)
		}

		if base, _, _ := tag.Raw(); base.String() == baseUndetermined {
			return fmt.Errorf("language '%s' is undetermined", name)
		}

		// уникальность проверяется по исходной строке, а не по разобранному тегу
		// (в отличие от mrlocale.NewBundle): каноничность проверена выше, поэтому
		// строка совпадает со своим каноном и сравнение строк равносильно сравнению тегов.
		// По той же причине пара разных записей одной локали сюда не доходит -
		// неканоничная из них отвергается раньше
		if uniqNames[name] {
			return fmt.Errorf("duplicate language name '%s'", name)
		}

		uniqNames[name] = true
	}

	return nil
}
