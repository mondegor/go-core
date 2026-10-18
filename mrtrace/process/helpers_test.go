package process_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrtrace"
	mrcontext "github.com/mondegor/go-core/mrtrace/context"
	"github.com/mondegor/go-core/mrtrace/process"
)

func TestExtractKeysValues(t *testing.T) {
	t.Parallel()

	values := process.ToKeyGetID(keyGetIDWithID())

	ctx := mrcontext.WithRequestID(t.Context(), "req-1")
	ctx = mrcontext.WithProcessID(ctx, "proc-1")

	got := process.ExtractKeysValues(ctx, values)
	assert.Equal(t, []any{mrtrace.KeyRequestID, "req-1", mrtrace.KeyProcessID, "proc-1"}, got)
	assert.Equal(t, len(got), cap(got))

	// контекст без ID
	assert.Nil(t, process.ExtractKeysValues(t.Context(), values))
}

func TestExtractKeysValues_Empty(t *testing.T) {
	t.Parallel()

	values := process.ToKeyGetID(keyGetIDWithID())
	ctx := mrcontext.WithRequestID(t.Context(), "req-1")

	assert.Nil(t, process.ExtractKeysValues(context.Background(), values))
	assert.Nil(t, process.ExtractKeysValues(ctx, nil))
}

func TestExtractCorrelationID(t *testing.T) {
	t.Parallel()

	values := []process.KeyGetID{
		{Key: mrtrace.KeyCorrelationID, GetID: mrcontext.CorrelationID},
		{Key: mrtrace.KeyRequestID, GetID: mrcontext.RequestID},
	}

	ctx := mrcontext.WithRequestID(t.Context(), "req-1")
	assert.Equal(t, "req-1", process.ExtractCorrelationID(ctx, values))

	// приоритет у ключа, указанного раньше
	ctx = mrcontext.WithCorrelationID(ctx, "corr-1")
	assert.Equal(t, "corr-1", process.ExtractCorrelationID(ctx, values))

	assert.Empty(t, process.ExtractCorrelationID(t.Context(), values))
	assert.Empty(t, process.ExtractCorrelationID(ctx, nil))
}

func TestToKeyGetID(t *testing.T) {
	t.Parallel()

	assert.Nil(t, process.ToKeyGetID(nil))

	got := process.ToKeyGetID(keyGetIDWithID())
	require.Len(t, got, 3)

	ctx := mrcontext.WithRequestID(t.Context(), "req-1")

	for i, v := range keyGetIDWithID() {
		assert.Equal(t, v.Key, got[i].Key)
		assert.Equal(t, v.GetID(ctx), got[i].GetID(ctx))
	}
}
