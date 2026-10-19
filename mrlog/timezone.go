package mrlog

import (
	"errors"
	"fmt"
	"time"
)

// ParseTimeZone - преобразует имя часового пояса в *time.Location.
// Поддерживаются IANA-имена (например: "Europe/Moscow"), а также "UTC" и "Local".
// Возвращает *time.Location при успехе, или time.UTC и ошибку
// при пустом или нераспознанном значении. Пустое значение проверяется
// отдельно, так как time.LoadLocation("") молча возвращает UTC без ошибки,
// что скрыло бы незаполненную настройку от вызывающего.
//
// ВНИМАНИЕ: для IANA-имён требуется база часовых поясов в системе;
// в минимальных образах (scratch, distroless) её нужно встроить
// в бинарник импортом `_ "time/tzdata"`.
// Значения "UTC" и "Local" доступны всегда.
//
// ВАЖНО: функция является хелпером логгера и рассчитана на однократный вызов
// при старте, так как time.LoadLocation читает базу часовых поясов с диска.
// Для прикладного кода следует использовать util/timezone.LocationList,
// который загружает все требуемые пояса заранее. Сводить их в общий хелпер
// не следует: mrlog намеренно не зависит ни от чего, кроме стандартной
// библиотеки, чтобы логгер поднимался раньше прикладных утилит.
func ParseTimeZone(value string) (*time.Location, error) {
	if value == "" {
		return time.UTC, errors.New("the timezone value is empty")
	}

	loc, err := time.LoadLocation(value)
	if err != nil {
		return time.UTC, fmt.Errorf("error loading timezone (value=%q): %w", value, err)
	}

	return loc, nil
}
