package config

import (
	"fmt"
	"strings"

	"github.com/mondegor/go-core/util/mime"
)

// ValidateMimeTypes - валидирует указанный список типов файлов.
// Расширения сравниваются без точки в начале, как их хранит mime.NewTypeList.
func ValidateMimeTypes(mimeTypes []mime.Type) error {
	uniqExtensions := make(map[string]bool, len(mimeTypes))

	for _, mt := range mimeTypes {
		ext := strings.TrimPrefix(mt.Extension, ".")

		if ext == "" {
			return fmt.Errorf("mimeType extension is required for content type '%s'", mt.ContentType)
		}

		if uniqExtensions[ext] {
			return fmt.Errorf("duplicate mimeType extension for content type (ext='%s', type='%s')", mt.Extension, mt.ContentType)
		}

		uniqExtensions[ext] = true
	}

	return nil
}
