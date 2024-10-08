package mrjulienrouter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrjulienrouter"
)

// Make sure the RouterAdapter conforms with the mrserver.HttpRouter interface.
func TestRouterAdapterImplementsHttpRouter(t *testing.T) {
	assert.Implements(t, (*mrserver.HttpRouter)(nil), &mrjulienrouter.RouterAdapter{})
}
