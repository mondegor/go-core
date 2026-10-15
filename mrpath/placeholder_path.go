package mrpath

import (
	"fmt"
	"strings"
)

const (
	// defaultPlaceholder - стандартный плейсхолдер для NewPlaceholderPath.
	defaultPlaceholder = "{{path}}"
)

type (
	// placeholderBuilder - вставляет значение между префиксом и постфиксом базового пути.
	placeholderBuilder struct {
		basePath string
		postfix  string
	}
)

// NewPlaceholder - создаёт Builder, который заменяет плейсхолдер в basePath на переданный путь.
// Параметры:
//   - basePath - шаблон пути с плейсхолдером (например: "/dir/{{path}}/postfix");
//   - placeholder - строка-плейсхолдер для замены (например: "{{path}}").
//
// Заменяется только первое вхождение плейсхолдера.
// "/" на стыках с плейсхолдером сводятся к одному, но не добавляются.
// Если плейсхолдер стоит в конце basePath, результат эквивалентен NewTail(префикс).
// Возвращает ошибку, если плейсхолдер пуст или не найден в basePath.
// Пример: NewPlaceholder("/dir/{id}/postfix", "{id}").BuildPath("x") → "/dir/x/postfix".
func NewPlaceholder(basePath, placeholder string) (Builder, error) {
	if placeholder == "" {
		return nil, fmt.Errorf("placeholder is empty for path %q", basePath)
	}

	if i := strings.Index(basePath, placeholder); i >= 0 {
		postfix := collapseLeadingSlashes(basePath[i+len(placeholder):])
		if postfix == "" {
			return NewTail(basePath[0:i]), nil
		}

		return &placeholderBuilder{
			basePath: collapseTrailingSlashes(basePath[0:i]),
			postfix:  postfix,
		}, nil
	}

	return nil, fmt.Errorf("placeholder is not found in path (placeholder=%q, path=%q)", placeholder, basePath)
}

// NewPlaceholderPath - создаёт Builder, который заменяет плейсхолдер "{{path}}" в basePath на переданный путь.
// Правила те же, что у NewPlaceholder.
// Пример: NewPlaceholderPath("/dir/{{path}}/postfix").BuildPath("x") → "/dir/x/postfix".
func NewPlaceholderPath(basePath string) (Builder, error) {
	return NewPlaceholder(basePath, defaultPlaceholder)
}

// BuildPath - вставляет path без ведущих и завершающих "/" на место плейсхолдера.
// Возвращает пустую строку, если path пуст или состоит только из "/".
func (p *placeholderBuilder) BuildPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}

	return p.basePath + path + p.postfix
}
