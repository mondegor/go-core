package mrmsg_test

import (
	"errors"
)

type (
	// failWriter - писатель, всегда возвращающий ошибку.
	failWriter struct{}
)

var errWrite = errors.New("write failed")

func (failWriter) Write([]byte) (int, error) {
	return 0, errWrite
}
