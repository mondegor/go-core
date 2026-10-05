package mrstorage

import (
	"math"
	"strconv"
)

const (
	maxLimit = math.MaxInt32 - 1 // верхний предел размера страницы (см. PageLimit)
)

// ToSQL - преобразует часть SQL-запроса в строку.
// Если part равен nil, возвращает пустую строку.
// Игнорирует аргументы, возвращает только SQL-выражение.
func ToSQL(part SQLPart) string {
	if part == nil {
		return ""
	}

	sql, _ := part.ToSQL()

	return sql
}

// NonZeroLimit - формирует SQL-конструкцию LIMIT с гарантированным минимумом в одну запись.
// Если value меньше 1, лимит принудительно устанавливается равным 1,
// что исключает формирование некорректного выражения (LIMIT 0 или отрицательный лимит).
func NonZeroLimit(value int) string {
	if value < 1 {
		return " LIMIT 1"
	}

	return " LIMIT " + strconv.Itoa(value)
}

// PageLimit - возвращает размер страницы в диапазоне [1, math.MaxInt32-1]:
// верхний предел гарантирует, что Limit+1 (запрос лишней записи для hasNext) не переполняет int32.
func PageLimit(value int) int {
	if value < 1 {
		return 1
	}

	if value > maxLimit {
		return maxLimit
	}

	return value
}
