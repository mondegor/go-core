package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	// Имя пояса процесса. В список LocationList не входит, поэтому валидатором отвергается.
	nameLocal = "Local"
)

// ValidateTimeZones - валидирует указанный список имён часовых поясов:
// имена проверяются на непустоту, уникальность и существование в базе
// часовых поясов. Проверка останавливается на первом негодном имени
// и сообщает о нём ошибкой.
//
// Пустой список ошибкой не считается: timezone.NewLocationList создаёт из него список
// из одного UTC, он же становится поясом по умолчанию (см. timezone.NewLocationList,
// timezone.LocationList.Default).
//
// "Local" (пояс процесса) отвергается: timezone.LocationList его не обслуживает
// (см. timezone.NewLocationList), поэтому в списке он бесполезен - кому нужен пояс
// процесса, берёт time.Local из стандартной библиотеки напрямую.
//
// ВНИМАНИЕ: проверка существования требует базу часовых поясов в системе;
// в минимальных образах (scratch, distroless) её нужно встроить
// в бинарник импортом `_ "time/tzdata"`, иначе будут отвергнуты все
// IANA-имена. Значение "UTC" доступно всегда.
func ValidateTimeZones(names []string) error {
	uniqNames := make(map[string]bool, len(names))

	for _, name := range names {
		if name == "" {
			return errors.New("timezone name is required")
		}

		if name == nameLocal {
			return errors.New("timezone 'Local' is not allowed in the list")
		}

		if uniqNames[name] {
			return fmt.Errorf("duplicate timezone name '%s'", name)
		}

		if _, err := time.LoadLocation(name); err != nil {
			return fmt.Errorf("error loading timezone (name='%s'): %w", name, err)
		}

		uniqNames[name] = true
	}

	return nil
}
