package mrpath

import (
	"strings"
)

type (
	// Builder - строит полный путь из базового пути и переданного значения.
	// Общие правила:
	//   - "/" на стыках базового пути со значением сводятся к одному, но не добавляются
	//     ("/" внутри базового пути и значения, а также "/" схемы URL, например "https://", не меняются);
	//   - ведущие "/" значения срезаются всегда (значение относительно базового пути),
	//     завершающие - только если за значением в базовом пути что-то следует;
	//   - если значение пусто или состоит только из "/", возвращается пустая строка.
	Builder interface {
		BuildPath(path string) string
	}
)

// collapseTrailingSlashes - сводит завершающие "/" к одному, если они есть.
// "/" сразу после ":" (схема URL, например "https://") не трогаются.
func collapseTrailingSlashes(s string) string {
	if t := strings.TrimRight(s, "/"); len(t) < len(s) && !strings.HasSuffix(t, ":") {
		return t + "/"
	}

	return s
}

// collapseLeadingSlashes - сводит ведущие "/" к одному, если они есть.
func collapseLeadingSlashes(s string) string {
	if t := strings.TrimLeft(s, "/"); len(t) < len(s) {
		return "/" + t
	}

	return s
}
