package mrtrace_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrtrace"
)

func TestNopTracer(t *testing.T) {
	t.Parallel()

	tracer := mrtrace.NopTracer()

	assert.NotPanics(t, func() {
		tracer.Trace(t.Context(), "key", "value")
		tracer.Trace(t.Context())
	})
}
