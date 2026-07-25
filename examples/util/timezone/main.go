package main

// база часовых поясов встраивается в бинарник, чтобы пример работал
// в минимальных образах, где она отсутствует в системе.
import (
	"os"
	"time"
	_ "time/tzdata"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/slog"
	"github.com/mondegor/go-core/util/timezone"
	"github.com/mondegor/go-core/util/timezone/config"
)

func main() {
	logger, _ := slog.NewLoggerAdapter(slog.WithWriter(os.Stdout))

	names := getTimeZoneListFromConfig()

	if err := config.ValidateTimeZones(names); err != nil {
		mrlog.Error(logger, "invalid timezone config", "error", err)

		return
	}

	// все пояса загружаются один раз при старте приложения
	list := timezone.NewLocationList(names)

	// список поясов приложения берётся у самого списка, а не у конфигурации:
	// UTC в нём есть всегда, даже если его не указывали, а негодные имена и повторы
	// отброшены, поэтому только эти имена принимает LocationByName
	mrlog.Info(logger, "registered timezones", "zones", list.TimeZones(), "default", list.Default().String())

	tm := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)

	resultWrapper := func(loc *time.Location, err error) string {
		if err != nil {
			return "ERROR!"
		}

		return tm.In(loc).Format(time.RFC3339)
	}

	mrlog.Info(logger, "Europe/Moscow", "time", resultWrapper(list.LocationByName("Europe/Moscow")))
	mrlog.Info(logger, "Asia/Tokyo", "time", resultWrapper(list.LocationByName("Asia/Tokyo")))
	mrlog.Info(logger, "UTC", "time", resultWrapper(list.LocationByName("UTC")))
	mrlog.Info(logger, "Nowhere/Nowhere", "time", resultWrapper(list.LocationByName("Nowhere/Nowhere")))

	printNameByOffset(logger, list)
	printReverseLookup(logger, list)
}

// printNameByOffset - подбор пояса по смещению и признаку летнего времени.
// Одно и то же смещение даёт разные результаты в зависимости от признака:
// +02:00 летом - это Europe/Berlin, а зимой такого пояса в списке нет.
func printNameByOffset(logger mrlog.Logger, list *timezone.LocationList) {
	lookups := []struct {
		offset time.Duration
		isDST  bool
	}{
		{offset: 1 * time.Hour, isDST: false},
		{offset: 2 * time.Hour, isDST: true},
		{offset: 2 * time.Hour, isDST: false},
		{offset: 3 * time.Hour, isDST: false},
		{offset: -4 * time.Hour, isDST: true},
		{offset: 13 * time.Hour, isDST: false},
	}

	for _, lookup := range lookups {
		name, ok := list.NameByOffset(lookup.offset, lookup.isDST)
		if !ok {
			mrlog.Info(logger, "zone is not found", "offset", lookup.offset, "isDST", lookup.isDST)

			continue
		}

		mrlog.Info(
			logger,
			"name by offset",
			"offset", lookup.offset,
			"isDST", lookup.isDST,
			"zone", name,
		)
	}
}

// printReverseLookup - обратный подбор: по конкретному моменту вычисляются
// смещение и признак летнего времени, а по ним восстанавливается имя пояса.
// Зимний и летний моменты одного пояса дают разные пары, но одно имя.
func printReverseLookup(logger mrlog.Logger, list *timezone.LocationList) {
	moments := []time.Time{
		time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC),
	}

	for _, name := range []string{"Europe/Berlin", "Asia/Tokyo"} {
		loc, err := list.LocationByName(name)
		if err != nil {
			mrlog.Error(logger, "timezone is not registered", "name", name, "error", err)

			continue
		}

		for _, moment := range moments {
			local := moment.In(loc)
			_, offset := local.Zone()

			found, ok := list.NameByOffset(time.Duration(offset)*time.Second, local.IsDST())
			if !ok {
				found = "<not found>"
			}

			mrlog.Info(
				logger,
				"reverse lookup",
				"source", name,
				"at", local.Format(time.RFC3339),
				"isDST", local.IsDST(),
				"zone", found,
			)
		}
	}
}

func getTimeZoneListFromConfig() []string {
	return []string{
		"Europe/Moscow",
		"Europe/Berlin",
		"Asia/Tokyo",
		"America/New_York",
	}
}
