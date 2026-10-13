package crypt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/util/crypt"
)

func TestIDGeneratorFunc(t *testing.T) {
	t.Parallel()

	var gen crypt.IDGenerator = crypt.IDGeneratorFunc(func() string { return "id-1" })

	assert.Equal(t, "id-1", gen.GenerateID())
}
