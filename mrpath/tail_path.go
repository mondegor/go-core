package mrpath

import (
	"strings"
)

type (
	// tailBuilder - добавляет значение в конец базового пути.
	tailBuilder struct {
		basePath string
	}
)

// NewTail - создаёт Builder, который добавляет переданный путь в конец basePath.
// Завершающие "/" в basePath сводятся к одному, но не добавляются.
// Пример: NewTail("/dir/").BuildPath("x") → "/dir/x", NewTail("/dir-").BuildPath("x") → "/dir-x".
func NewTail(basePath string) Builder {
	return &tailBuilder{
		basePath: collapseTrailingSlashes(basePath),
	}
}

// BuildPath - добавляет path без ведущих "/" в конец базового пути (завершающие "/" сохраняются).
// Возвращает пустую строку, если path пуст или состоит только из "/".
func (p *tailBuilder) BuildPath(path string) string {
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return ""
	}

	return p.basePath + path
}
