package ordered

import (
	"slices"
)

type (
	ordered interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
			~string
	}
)

// FilterFunc - фильтрует слайс целых/строк на месте (без выделения нового слайса).
// Удаляет элементы, для которых check возвращает false.
// Возвращает усечённый слайс.
func FilterFunc[T ordered](s []T, check func(el T) bool) []T {
	for i := 0; i < len(s); i++ {
		if check(s[i]) {
			continue
		}

		s2 := s[i:]
		for i2 := 1; i2 < len(s2); i2++ {
			if check(s2[i2]) {
				s[i] = s2[i2]
				i++
			}
		}

		clear(s[i:]) // zero/nil out the obsolete elements, for GC

		return s[:i]
	}

	return s
}

// SortedUnique - сортирует переданный слайс на месте и возвращает его срез
// без дубликатов (без выделения нового массива).
func SortedUnique[T ordered](s []T) []T {
	slices.Sort(s)

	return slices.Compact(s)
}

// SortedUniqueClone - возвращает новый отсортированный слайс без дубликатов;
// исходный слайс не изменяется.
func SortedUniqueClone[T ordered](s []T) []T {
	return SortedUnique(slices.Clone(s))
}

// BinaryIndex - возвращает индекс указанного элемента или -1, если его нет.
// Указанный массив обязан быть предварительно отсортирован.
func BinaryIndex[T ordered](s []T, value T) int {
	i := binaryIndex(s, value)

	if i == -1 || s[i] != value {
		return -1
	}

	return i
}

// BinaryContains - сообщает, содержится ли указанный элемент в массиве.
// Указанный массив обязан быть предварительно отсортирован.
func BinaryContains[T ordered](s []T, value T) bool {
	return BinaryIndex(s, value) >= 0
}

// BinaryAppend - добавляет элемент в массив, оставляя его отсортированным.
// Как и append, может изменить исходный массив, поэтому используется результат.
// Указанный массив обязан быть предварительно отсортирован.
func BinaryAppend[T ordered](s []T, value T) []T {
	return binaryAppend(s, value, false)
}

// UniqueBinaryAppend - добавляет элемент в массив, если такого ещё не было,
// оставляя массив отсортированным.
// Указанный массив обязан быть предварительно отсортирован.
func UniqueBinaryAppend[T ordered](s []T, value T) []T {
	return binaryAppend(s, value, true)
}

func binaryAppend[T ordered](s []T, value T, unique bool) []T {
	i := binaryIndex(s, value)

	if unique && i != -1 && value == s[i] {
		return s
	}

	if i == -1 || i == len(s)-1 && value > s[i] {
		return append(s, value)
	}

	var zero T

	c := append(s, zero) //nolint:gocritic
	copy(c[i+1:], s[i:])
	c[i] = value

	return c
}

func binaryIndex[T ordered](s []T, value T) int {
	i, j := 0, len(s)

	for i < j {
		// avoid overflow
		h := int(uint(i+j) >> 1)

		switch {
		case s[h] < value:
			i = h + 1
		case s[h] > value:
			j = h
		default:
			return h
		}
	}

	if i < len(s) {
		return i
	}

	return i - 1
}
