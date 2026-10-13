package xio_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xio"
)

type (
	fakeCloser struct {
		name   string
		err    error
		closed *[]string
	}
)

func (c *fakeCloser) Close() error {
	*c.closed = append(*c.closed, c.name)

	return c.err
}

func TestCloseManager(t *testing.T) {
	t.Parallel()

	var closed []string

	logger := &recordLogger{}
	manager := xio.NewCloseManager(logger)

	manager.Register(&fakeCloser{name: "first", closed: &closed})
	manager.Register(&fakeCloser{name: "second", err: errors.New("close failed"), closed: &closed})
	manager.Register(&fakeCloser{name: "third", closed: &closed})

	manager.Close()

	// ошибка закрытия одного ресурса не прерывает закрытие остальных
	assert.Equal(t, []string{"first", "second", "third"}, closed)
	assert.Equal(t, []string{"debug", "debug", "debug", "debug", "error", "debug"}, logger.levels())
}

func TestCloseManagerEmpty(t *testing.T) {
	t.Parallel()

	logger := &recordLogger{}

	xio.NewCloseManager(logger).Close()

	assert.Empty(t, logger.levels())
}
