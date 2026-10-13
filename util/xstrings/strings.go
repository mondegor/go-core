package xstrings

import (
	"slices"
	"strings"
)

// TrimBeforeSep - возвращает строку без её начальной части продолжающейся
// до сепаратора включая этот сепаратор.
// Если сепаратор не был найден, то возвращается исходная строка.
func TrimBeforeSep(s string, sep byte) string {
	if i := strings.IndexByte(s, sep); i >= 0 {
		return s[i+1:]
	}

	return s
}

// SortedUnique - сортирует s на месте и возвращает его же с удалёнными дубликатами (не копию).
func SortedUnique(s []string) []string {
	slices.Sort(s)

	return slices.Compact(s)
}
