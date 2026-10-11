# Описание GoCore v0.15.3
Этот репозиторий содержит описание библиотеки GoCore.

## Статус библиотеки
Библиотека находится в стадии разработки.

## Описание библиотеки
Библиотека решает несколько основных задач:
- формирует пользовательские сообщения на различных языках, в том числе и сообщения об ошибках;
- предоставляет инструменты для более удобной обработки ошибок как пользовательских,
  так и программных согласующихся с Go подходом (подробнее см. [errors](errors/README.md));
- предоставляет систему логирования сообщений и ошибок на основе `slog`;
- содержит вспомогательные подсистемы: работа с хранилищами и PostgreSQL, фоновые процессы,
  контроль доступа, трассировка запросов, утилиты;

## Подключение библиотеки
`go get -u github.com/mondegor/go-core@v0.15.3`

Требуется Go 1.26 или новее.

## Установка библиотеки для её локальной разработки
- Выбрать рабочую директорию, где должна быть расположена библиотека
- `mkdir go-core && cd go-core` // создать и перейти в директорию проекта
- `git clone git@github.com:mondegor/go-core.git .`
- `cp .env.dist .env`
- `mrcmd go-dev deps` // загрузка зависимостей проекта
- Для работы утилит `golangci-lint` (линтеры и форматтеры), `mockgen`, `gotext` необходимо запустить
  `mrcmd go-dev install-tools`. По умолчанию `golangci-lint` устанавливается последней версии;
  чтобы закрепить версию, раскомментируйте переменную `GO_DEV_TOOLS_INSTALL_*` в `.env`.
  `mockgen` и `gotext` в go-dev по умолчанию выключены, их версии заданы в `.env.dist`;
- `go install ./cmd/gotext-catalog-fix` // установка утилиты, используемой при генерации каталогов локализации

### Запуск тестов
Тесты библиотеки запускаются командой `go test ./...`, внешние сервисы для них не нужны.

### Консольные команды используемые при разработке библиотеки

> Перед запуском консольных скриптов библиотеки необходимо скачать и установить утилиту Mrcmd.\
> Инструкция по её установке находится [здесь](https://github.com/mondegor/mrcmd#readme)

- `mrcmd go-dev help` // выводит список всех доступных go-dev команд;
- `mrcmd go-dev generate` // генерирует go файлы через встроенный механизм go:generate;
- `mrcmd go-dev fmt` // исправляет форматирование кода и imports (`golangci-lint fmt`: форматтеры из раздела `formatters` в `.golangci.yaml`, сгенерированные файлы пропускаются);
- `mrcmd go-dev fmt-diff` // показывает изменения, которые внесёт `fmt`, не меняя файлы (`golangci-lint fmt -d`);
- `mrcmd go-dev lint` // запускает линтеры для проверки кода (на основе `.golangci.yaml`);
- `mrcmd go-dev lint-verify` // проверяет `.golangci.yaml` по JSON-схеме (`golangci-lint config verify`, нужен доступ к golangci-lint.run);
- `mrcmd go-dev test` // запускает тесты библиотеки;
- `mrcmd go-dev test-report` // запускает тесты библиотеки с формированием отчёта о покрытии кода (`test-coverage-full.html`);
- `mrcmd plantuml build-all` // генерирует файлы изображений из `.puml` [подробнее](https://github.com/mondegor/mrcmd-plugins/blob/master/plantuml/README.md#%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%B0-%D1%81-%D0%B4%D0%BE%D0%BA%D1%83%D0%BC%D0%B5%D0%BD%D1%82%D0%B0%D1%86%D0%B8%D0%B5%D0%B9-%D0%BF%D1%80%D0%BE%D0%B5%D0%BA%D1%82%D0%B0-markdown--plantuml);

#### Короткий вариант выше приведённых команд (Makefile)
- `make deps` // аналог `mrcmd go-dev deps`
- `make deps-upgrade` // аналог `mrcmd go-dev get -u ./...` + `mrcmd go-dev tidy`
- `make generate` // аналог `mrcmd go-dev generate`
- `make lint` // аналог `mrcmd go-dev fmt` + `lint`
- `make lint-verify` // аналог `mrcmd go-dev lint-verify`
- `make test` // аналог `mrcmd go-dev test`
- `make test-report` // аналог `mrcmd go-dev test-report`
- `make plantuml` // аналог `mrcmd plantuml build-all`

Дополнительные команды (Makefile.mk):
- `make check-and-fix` // generate + форматирование + lint + test + plantuml;
- `make full` // `make deps` + `make check-and-fix`;
- `make full2` // `make deps-upgrade` + `make check-and-fix`;
- `make archive` // упаковывает проект в `../go-core.tar.gz`;

> Чтобы расширить список команд, необходимо создать Makefile.mk и добавить
> туда дополнительные команды, все они будут добавлены в единый список команд make утилиты.

## Обработка ошибок
Описание видов ошибок, архитектуры их обработки и C4-диаграммы: [errors](errors/README.md).
