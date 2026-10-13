package xio_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/xio"
)

type (
	failWriter struct{}
)

func (failWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWrite(t *testing.T) {
	t.Parallel()

	logger := &recordLogger{}

	var buf bytes.Buffer

	xio.Write(context.Background(), logger, &buf, []byte("data"))
	assert.Equal(t, "data", buf.String())
	assert.Empty(t, logger.levels())

	xio.Write(context.Background(), logger, failWriter{}, nil) // пустые данные не пишутся
	assert.Empty(t, logger.levels())

	xio.Write(context.Background(), logger, failWriter{}, []byte("data"))
	assert.Equal(t, []string{"error"}, logger.levels())
}
