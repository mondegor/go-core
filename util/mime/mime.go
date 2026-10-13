package mime

import (
	"errors"
	"fmt"
	"strings"
)

type (
	// TypeList - хранит соответствие расширений их типам файлов (в обе стороны).
	TypeList struct {
		contentTypeMap map[string]string
		extensionMap   map[string]string
	}

	// Type - хранит расширение и соответствующий ему тип файла.
	Type struct {
		ContentType string `yaml:"type"`
		Extension   string `yaml:"ext"`
	}
)

// NewTypeList - создаёт объект TypeList на основе списка соответствий расширений и типов файлов.
// Точка в начале расширения отбрасывается.
func NewTypeList(items []Type) *TypeList {
	mimeMap := make(map[string]string, len(items))
	extMap := make(map[string]string, len(items))

	for _, item := range items {
		item.Extension = strings.TrimPrefix(item.Extension, ".")
		extMap[item.Extension] = item.ContentType

		// т.к. у одного типа может быть несколько расширений,
		// то в индекс попадает только первый зарегистрированный
		if _, ok := mimeMap[item.ContentType]; !ok {
			mimeMap[item.ContentType] = item.Extension
		}
	}

	return &TypeList{
		contentTypeMap: mimeMap,
		extensionMap:   extMap,
	}
}

// TypesByExts - возвращает типы для указанных расширений (Extension сохраняется в переданном виде).
// Если хотя бы одно расширение не зарегистрировано, возвращается ошибка.
func (mt *TypeList) TypesByExts(values []string) ([]Type, error) {
	mime := make([]Type, len(values))

	for i, ext := range values {
		contentType, err := mt.ContentTypeByExt(ext)
		if err != nil {
			return nil, err
		}

		mime[i] = Type{
			ContentType: contentType,
			Extension:   ext,
		}
	}

	return mime, nil
}

// ContentTypeByExt - возвращает тип файла по расширению (допускается точка в начале).
// Если тип не найден или расширение пустое, возвращается ошибка.
func (mt *TypeList) ContentTypeByExt(value string) (string, error) {
	if value == "" || len(value) == 1 && value[0] == '.' {
		return "", errors.New("arg 'value' is empty")
	}

	if value[0] == '.' { // если указано расширение с точкой в начале
		value = value[1:]
	}

	if ext, ok := mt.extensionMap[value]; ok {
		return ext, nil
	}

	return "", fmt.Errorf("mime not found for arg '%s'", value)
}

// ExtByContentType - возвращает расширение по типу файла (первое из зарегистрированных для этого типа).
// Если расширение не найдено или тип пустой, возвращается ошибка.
func (mt *TypeList) ExtByContentType(value string) (string, error) {
	if value == "" {
		return "", errors.New("arg 'value' is empty")
	}

	if contentType, ok := mt.contentTypeMap[value]; ok {
		return contentType, nil
	}

	return "", fmt.Errorf("ext not found for arg '%s'", value)
}
