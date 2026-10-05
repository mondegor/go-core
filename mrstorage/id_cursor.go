package mrstorage

import (
	"math"
	"strconv"

	"github.com/mondegor/go-core/mrtype"
)

type (
	// IDCursor - позиция keyset-выборки по положительному целочисленному ключу, хранимому в bigint (int8 в PostgreSQL):
	// ID - ключ последней полученной записи (0 - первая страница), Limit - размер страницы.
	// Нулевое значение означает первую страницу при любом направлении выборки: граница
	// для запроса берётся методом AfterID или BeforeID, а не из поля ID напрямую.
	IDCursor struct {
		ID    int64
		Limit int
	}
)

// NewIDCursor - создаёт объект IDCursor из параметров курсорной пагинации: значение
// курсора - ключ последней полученной записи; неразбираемое, отрицательное или выходящее
// за диапазон bigint значение означает первую страницу.
func NewIDCursor(params mrtype.CursorParams) IDCursor {
	id, err := strconv.ParseInt(params.Value, 10, 64)
	if err != nil || id < 0 {
		id = 0
	}

	return IDCursor{
		ID:    id,
		Limit: PageLimit(params.Limit),
	}
}

// AfterID - возвращает границу выборки по возрастанию ключа (условие `id > $n`):
// для первой страницы - 0, меньше любого ключа.
func (c IDCursor) AfterID() int64 {
	return c.ID
}

// BeforeID - возвращает границу выборки по убыванию ключа (условие `id < $n`):
// для первой страницы - максимум bigint (math.MaxInt64).
func (c IDCursor) BeforeID() int64 {
	if c.ID == 0 {
		return math.MaxInt64
	}

	return c.ID
}
