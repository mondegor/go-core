package process

import (
	"context"
	"fmt"

	"github.com/mondegor/go-core/mrtrace"
)

type (
	contextManager struct {
		keyGetID  map[string]func(ctx context.Context) string
		keyWithID map[string]func(ctx context.Context, id string) context.Context

		extractKeysValues    []KeyGetID
		extractCorrelationID []KeyGetID

		idGenerator idGenerator
		logger      logger
	}

	idGenerator interface {
		GenerateID() string
	}

	logger interface {
		Warn(ctx context.Context, msg string, args ...any)
	}
)

// NewContextManager - создаёт менеджер контекста для управления ID процессов.
// Параметры:
//   - keyGetIDWithID - список триад (ключ, функция получения ID, функция установки ID);
//   - correlationKeys - ключи, значения которых могут служить ID корреляции, в порядке приоритета;
//   - idGenerator - генератор уникальных ID;
//   - logger - логгер для предупреждений об обращении к незарегистрированным ключам.
//
// Возвращает ошибку при дублировании ключей в keyGetIDWithID или если какого-либо ключа
// из correlationKeys нет в keyGetIDWithID (повторы в correlationKeys не проверяются).
func NewContextManager(
	keyGetIDWithID []KeyGetIDWithID,
	correlationKeys []string,
	idGenerator idGenerator,
	logger logger,
) (mrtrace.ContextManager, error) {
	cm := contextManager{
		logger:      logger,
		idGenerator: idGenerator,

		keyGetID:  make(map[string]func(ctx context.Context) string, len(keyGetIDWithID)),
		keyWithID: make(map[string]func(ctx context.Context, id string) context.Context, len(keyGetIDWithID)),

		extractKeysValues:    make([]KeyGetID, 0, len(keyGetIDWithID)),
		extractCorrelationID: make([]KeyGetID, 0, len(correlationKeys)),
	}

	for _, v := range keyGetIDWithID {
		if _, ok := cm.keyGetID[v.Key]; ok {
			return nil, fmt.Errorf("duplicated key %q", v.Key)
		}

		cm.keyGetID[v.Key] = v.GetID
		cm.keyWithID[v.Key] = v.WithID

		cm.extractKeysValues = append(
			cm.extractKeysValues,
			KeyGetID{
				Key:   v.Key,
				GetID: v.GetID,
			},
		)
	}

	for _, key := range correlationKeys {
		getID, ok := cm.keyGetID[key]
		if !ok {
			return nil, fmt.Errorf("correlation key not found (key='%s')", key)
		}

		cm.extractCorrelationID = append(
			cm.extractCorrelationID,
			KeyGetID{
				Key:   key,
				GetID: getID,
			},
		)
	}

	return &cm, nil
}

// ProcessID - извлекает ID процесса из контекста по указанному ключу.
// Если ключ не зарегистрирован, логирует предупреждение и возвращает пустую строку.
func (e *contextManager) ProcessID(ctx context.Context, key string) string {
	if fn, ok := e.keyGetID[key]; ok {
		return fn(ctx)
	}

	e.logger.Warn(ctx, "process key not found", "key", key)

	return ""
}

// WithProcessID - устанавливает указанный ID процесса в контекст по ключу.
// Если ключ не зарегистрирован, логирует предупреждение и возвращает исходный контекст без изменений.
func (e *contextManager) WithProcessID(ctx context.Context, key, value string) context.Context {
	if fn, ok := e.keyWithID[key]; ok {
		return fn(ctx, value)
	}

	e.logger.Warn(ctx, "process key not found", "key", key)

	return ctx
}

// WithGeneratedProcessID - генерирует новый уникальный ID процесса и устанавливает его в контекст.
// Если ключ не зарегистрирован, логирует предупреждение и возвращает исходный контекст без изменений.
func (e *contextManager) WithGeneratedProcessID(ctx context.Context, key string) context.Context {
	if fn, ok := e.keyWithID[key]; ok {
		return fn(ctx, e.idGenerator.GenerateID())
	}

	e.logger.Warn(ctx, "process key not found", "key", key)

	return ctx
}

// ExtractKeysValues - извлекает все зарегистрированные ID процессов из контекста (см. функцию ExtractKeysValues).
func (e *contextManager) ExtractKeysValues(ctx context.Context) []any {
	return ExtractKeysValues(ctx, e.extractKeysValues)
}

// ExtractCorrelationID - возвращает первый непустой ID по ключам корреляции
// в порядке, указанном при создании, или пустую строку.
func (e *contextManager) ExtractCorrelationID(ctx context.Context) string {
	return ExtractCorrelationID(ctx, e.extractCorrelationID)
}
