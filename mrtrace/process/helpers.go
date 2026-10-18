package process

import (
	"context"
)

type (
	// KeyGetIDWithID - тройка: текстовое название ID, функция получения ID из контекста и функция установки значения в контекст.
	KeyGetIDWithID struct {
		Key    string
		GetID  func(ctx context.Context) string
		WithID func(ctx context.Context, id string) context.Context
	}

	// KeyGetID - пара: текстовое название ID и функция получения ID из контекста.
	KeyGetID struct {
		Key   string
		GetID func(ctx context.Context) string
	}
)

// ExtractKeysValues - извлекает ID процессов из контекста по указанному списку ключей.
// Возвращает плоский слайс пар ключ/значение: ["key1", "value1", "key2", "value2", ...],
// ключи с пустыми значениями пропускаются. Если ни одного ID не найдено, возвращает nil.
func ExtractKeysValues(ctx context.Context, values []KeyGetID) (keyValue []any) {
	if ctx == context.Background() || len(values) == 0 {
		return nil
	}

	for _, v := range values {
		if value := v.GetID(ctx); value != "" {
			if keyValue == nil {
				// слайс выделяется только при первом найденном ID: на каждый ID - ключ и значение
				keyValue = make([]any, 0, 2*len(values))
			}

			keyValue = append(keyValue, v.Key, value)
		}
	}

	return keyValue[0:len(keyValue):len(keyValue)]
}

// ExtractCorrelationID - возвращает первый непустой ID из контекста по списку ключей
// (ID корреляции может храниться под разными ключами) или пустую строку.
func ExtractCorrelationID(ctx context.Context, values []KeyGetID) string {
	for _, v := range values {
		if id := v.GetID(ctx); id != "" {
			return id
		}
	}

	return ""
}

// ToKeyGetID - преобразует слайс KeyGetIDWithID в слайс KeyGetID,
// отбрасывая функции установки значений (WithID).
func ToKeyGetID(values []KeyGetIDWithID) []KeyGetID {
	if len(values) == 0 {
		return nil
	}

	list := make([]KeyGetID, 0, len(values))

	for _, v := range values {
		list = append(
			list,
			KeyGetID{
				Key:   v.Key,
				GetID: v.GetID,
			},
		)
	}

	return list
}
