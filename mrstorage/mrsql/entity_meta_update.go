package mrsql

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrlog"
)

const (
	// modelNameEntityMetaUpdate - имя сущности для логирования и сообщений об ошибках.
	modelNameEntityMetaUpdate = "EntityMetaUpdate"

	// fieldTagDBFieldName - имя тега для указания имени поля в БД.
	fieldTagDBFieldName = "db"

	// fieldTagFieldUpdate - имя тега для указания полей, доступных для обновления.
	fieldTagFieldUpdate = "upd"
)

type (
	// EntityMetaUpdate - объект для управления динамическим обновлением записей в БД.
	// Информация об обновлении считывается из тегов структуры:
	//   - `db`: имя поля в БД, если в `upd` установлено значение "+";
	//   - `upd`: иначе в самом теге `upd` должно быть указано имя поля в БД;
	// Если в поле структуры нет нужного тега, то оно не используется.
	EntityMetaUpdate struct {
		structName string
		fieldsInfo map[int]fieldInfo // field index -> fieldInfo
	}

	// fieldInfo - информация о поле структуры для обновления.
	fieldInfo struct {
		kind      reflect.Kind // kind - тип поля (int, string, time.Time и т.д.)
		isPointer bool         // isPointer - является ли поле указателем
		dbName    string       // dbName - имя поля в БД
	}
)

var regexpDbName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// NewEntityMetaUpdate - создаёт объект EntityMetaUpdate для управления обновлением записей.
// Анализирует теги структуры entity и формирует карту полей для обновления.
// Параметры:
//   - logger - логгер для вывода информации о разборе тегов;
//   - entity - структура для разбора (может быть указателем).
func NewEntityMetaUpdate(logger mrlog.Logger, entity any) (*EntityMetaUpdate, error) {
	rvt := reflect.TypeOf(entity)
	logger = mrlog.WithAttrs(logger, "object", fmt.Sprintf("[%s] %s", modelNameEntityMetaUpdate, rvt.String()))

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

	meta := EntityMetaUpdate{
		structName: rvt.String(),
		fieldsInfo: make(map[int]fieldInfo),
	}

	for i, cnt := 0, rvt.NumField(); i < cnt; i++ {
		field := rvt.Field(i)
		update := field.Tag.Get(fieldTagFieldUpdate)
		dbName := field.Tag.Get(fieldTagDBFieldName)

		if update == "" {
			continue
		}

		dbName, err := parseTagUpdate(rvt, update, dbName)
		if err != nil {
			logger.Warn(context.Background(), "parse tag update warning, skipped", "error", err)

			continue
		}

		fieldType := field.Type
		isPointer := false

		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
			isPointer = true
		}

		if !checkEntityMetaUpdateFieldType(fieldType) {
			logger.Warn(
				context.Background(),
				"check field type warning, skipped",
				"error", fmt.Errorf("field is not supported (field=%s, type=%s)", rvt.Field(i).Name, fieldType.Kind()),
			)

			continue
		}

		meta.fieldsInfo[i] = fieldInfo{
			kind:      fieldType.Kind(),
			isPointer: isPointer,
			dbName:    dbName,
		}

		if mrlog.DebugEnabled(logger) {
			debugInfo = append(
				debugInfo,
				fmt.Sprintf(
					"- %s(%d, %s) -> %s;",
					rvt.Field(i).Name, i, rvt.Field(i).Type, dbName,
				),
			)
		}
	}

	if len(debugInfo) > 0 {
		logger.Debug(context.Background(), strings.Join(debugInfo, "\n"))
	}

	return &meta, nil
}

// parseTagUpdate - разбирает тег обновления поля структуры.
func parseTagUpdate(rvt reflect.Type, value, dbName string) (string, error) {
	errFunc := func(errString string) (string, error) {
		return "", fmt.Errorf(
			"[%s] %s: parse error in %q: %s",
			modelNameEntityMetaUpdate,
			rvt.String(),
			value,
			errString,
		)
	}

	if value == "+" {
		if dbName == "" {
			return errFunc(fmt.Sprintf("tag %s is empty", fieldTagDBFieldName))
		}

		if !regexpDbName.MatchString(dbName) {
			return errFunc(fmt.Sprintf("value %q from %s is incorrect", dbName, fieldTagDBFieldName))
		}

		return dbName, nil
	}

	if !regexpDbName.MatchString(value) {
		return errFunc(fmt.Sprintf("value %q from %s is incorrect", value, fieldTagFieldUpdate))
	}

	return value, nil
}

// FieldsForUpdate - возвращает список полей и их значения для формирования SQL-запроса UPDATE.
func (m *EntityMetaUpdate) FieldsForUpdate(entity any) ([]string, []any, error) {
	rv := reflect.ValueOf(entity)

	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil, errors.NewInternalError(
			"reflect field is invalid",
			"entity", rv,
		)
	}

	if rv.Type().String() != m.structName {
		return nil, nil, errors.ErrInternalInvalidType.New(
			"type", rv.Type(),
			"expected", m.structName,
		)
	}

	fields := make([]string, 0, len(m.fieldsInfo))
	args := make([]any, 0, cap(fields))

	for i, info := range m.fieldsInfo {
		field := rv.Field(i)

		if !field.IsValid() {
			return nil, nil, errors.NewInternalError(
				"reflect field is invalid",
				"field", field,
			)
		}

		if info.isPointer {
			if field.IsNil() {
				continue
			}

			field = rv.Field(i).Elem()
		}

		switch info.kind {
		case reflect.String, reflect.Slice: // empty slice === nil
			if !info.isPointer && field.Len() == 0 {
				continue
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Bool, reflect.Array:
			if !info.isPointer && field.IsZero() {
				continue
			}

		case reflect.Struct:
			v := field.Interface()

			value, ok := v.(time.Time)
			if !ok {
				return nil, nil, errors.NewInternalError(
					"reflect field is unknown struct",
					"field", field,
				)
			}

			if !info.isPointer && value.IsZero() {
				continue
			}

		default:
			return nil, nil, errors.NewInternalError(
				"reflect field is undefined",
				"field", field,
			)
		}

		fields = append(fields, info.dbName)
		args = append(args, field.Interface())
	}

	return fields, args, nil
}

// checkEntityMetaUpdateFieldType - проверяет, поддерживается ли тип поля для обновления.
// Поддерживаемые типы:
//   - Примитивы: string, int*, uint*, float*, bool
//   - Массивы: uuid.UUID
//   - Слайсы: []byte
//   - Структуры: time.Time
func checkEntityMetaUpdateFieldType(fieldType reflect.Type) bool {
	switch fieldType.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return true

	case reflect.Array:
		return fieldType.String() == "uuid.UUID"

	case reflect.Slice:
		return fieldType.Elem().Name() == "uint8" // byte

	case reflect.Struct:
		return fieldType.String() == "time.Time"
	}

	return false
}
