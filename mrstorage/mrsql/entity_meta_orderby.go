package mrsql

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-core/mrtype/sortdirection"
)

const (
	// modelNameEntityMetaOrderBy - имя сущности для логирования и сообщений об ошибках.
	modelNameEntityMetaOrderBy = "EntityMetaOrderBy"

	// fieldTagSortByField - имя тега для указания полей/выражений сортировки.
	fieldTagSortByField = "sort"
)

type (
	// EntityMetaOrderBy - объект для управления порядком сортировки записей БД.
	// Информация о порядке считывается из тегов структуры с именем `sort`.
	// Формат тега: "имя_поля_БД[,default][,asc|desc]"
	//   - имя_поля_БД: обязательное, имя поля или выражения в БД;
	//   - default - необязательное, указывает поле сортировки по умолчанию;
	//   - asc|desc - необязательное, направление сортировки (по умолчанию: ASC).
	EntityMetaOrderBy struct {
		fieldMap    map[string]bool
		defaultSort mrtype.SortParams
	}

	// parsedTagSort - результат разбора тега sort.
	parsedTagSort struct {
		SortName      string
		IsDefault     bool
		SortDirection sortdirection.Enum
	}
)

// NewEntityMetaOrderBy - создаёт объект EntityMetaOrderBy для управления сортировкой.
// Анализирует теги `sort` структуры entity и формирует карту допустимых полей сортировки.
// Параметры:
//   - logger - логгер для вывода информации о разборе тегов;
//   - entity - структура для разбора (может быть указателем).
func NewEntityMetaOrderBy(logger mrlog.Logger, entity any) (*EntityMetaOrderBy, error) {
	rvt := reflect.TypeOf(entity)
	logger = mrlog.WithAttrs(logger, "object", fmt.Sprintf("[%s] %s", modelNameEntityMetaOrderBy, rvt.String()))

	for rvt.Kind() == reflect.Pointer {
		rvt = rvt.Elem()
	}

	if rvt.Kind() != reflect.Struct {
		return nil, errors.ErrInternalInvalidType.New(
			"type", rvt.Kind(),
			"expected", reflect.Struct,
		)
	}

	var debugInfo []string

	meta := EntityMetaOrderBy{
		fieldMap: make(map[string]bool),
	}

	for i, cnt := 0, rvt.NumField(); i < cnt; i++ {
		fieldType := rvt.Field(i)
		sort := fieldType.Tag.Get(fieldTagSortByField)

		if sort == "" {
			continue
		}

		parsed, err := parseTagSort(rvt, sort, meta.defaultSort.Column == "")
		if err != nil {
			logger.Warn(context.Background(), "parse tag sort warning, skipped")

			continue
		}

		var extMessage string

		if parsed.IsDefault {
			meta.defaultSort.Column = parsed.SortName
			meta.defaultSort.Direction = parsed.SortDirection
			extMessage = ", default"
		}

		meta.fieldMap[parsed.SortName] = true

		if mrlog.DebugEnabled(logger) {
			debugInfo = append(
				debugInfo,
				fmt.Sprintf(
					"- %s(%d, %s) -> %s %s%s;",
					rvt.Field(i).Name, i, rvt.Field(i).Type, parsed.SortName, parsed.SortDirection.String(), extMessage,
				),
			)
		}
	}

	if len(debugInfo) > 0 {
		logger.Debug(context.Background(), strings.Join(debugInfo, "\n"))
	}

	return &meta, nil
}

// HasColumn - сообщает, зарегистрировано ли указанное поле как допустимое для сортировки.
func (m *EntityMetaOrderBy) HasColumn(name string) bool {
	_, ok := m.fieldMap[name]

	return ok
}

// DefaultSort - возвращает параметры сортировки по умолчанию.
// Если в структуре не указано поле с флагом `default`, возвращает пустые параметры.
func (m *EntityMetaOrderBy) DefaultSort() mrtype.SortParams {
	return m.defaultSort
}

// parseTagSort - разбирает тег sort поля структуры.
func parseTagSort(rvt reflect.Type, value string, canBeDefault bool) (parsedTagSort, error) {
	parsed := strings.Split(value, ",")
	count := len(parsed)

	errFunc := func(errString string) (parsedTagSort, error) {
		return parsedTagSort{}, fmt.Errorf(
			"[%s] %s: parse error in %q: %s",
			modelNameEntityMetaOrderBy,
			rvt.String(),
			value,
			errString,
		)
	}

	if count > 3 {
		return errFunc("incorrect value")
	}

	if parsed[0] == "" {
		return errFunc("field name is required")
	}

	if !regexpDbName.MatchString(parsed[0]) {
		return errFunc("field name is incorrect")
	}

	isDefault := false

	if count > 1 {
		if parsed[1] != "default" {
			return errFunc("the second parameter can only be equal to 'default'")
		}

		isDefault = true
	}

	if !canBeDefault && isDefault {
		return errFunc("default field already exists")
	}

	tagSort := parsedTagSort{
		SortName:      parsed[0],
		IsDefault:     isDefault,
		SortDirection: sortdirection.ASC,
	}

	if count > 2 {
		sortDirection, err := sortdirection.Parse(strings.ToUpper(parsed[2]))
		if err != nil {
			return errFunc("the third parameter can only be equal to 'asc' or 'desc'")
		}

		tagSort.SortDirection = sortDirection
	}

	return tagSort, nil
}
