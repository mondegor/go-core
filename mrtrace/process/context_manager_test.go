package process_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrtrace"
	mrcontext "github.com/mondegor/go-core/mrtrace/context"
	"github.com/mondegor/go-core/mrtrace/process"
	"github.com/mondegor/go-core/util/crypt"
)

func newContextManager(t *testing.T, logger *warnLogger) mrtrace.ContextManager {
	t.Helper()

	cm, err := process.NewContextManager(
		keyGetIDWithID(),
		[]string{mrtrace.KeyCorrelationID, mrtrace.KeyRequestID},
		crypt.IDGeneratorFunc(func() string { return "generated-id" }),
		logger,
	)
	require.NoError(t, err)

	return cm
}

func TestNewContextManager_Errors(t *testing.T) {
	t.Parallel()

	idGenerator := crypt.IDGeneratorFunc(process.GenerateID)

	_, err := process.NewContextManager(
		append(keyGetIDWithID(), keyGetIDWithID()[0]),
		nil,
		idGenerator,
		&warnLogger{},
	)
	require.Error(t, err)

	_, err = process.NewContextManager(
		keyGetIDWithID(),
		[]string{mrtrace.KeyTaskID},
		idGenerator,
		&warnLogger{},
	)
	require.Error(t, err)
}

func TestContextManager_ProcessID(t *testing.T) {
	t.Parallel()

	logger := &warnLogger{}
	cm := newContextManager(t, logger)

	ctx := cm.WithProcessID(t.Context(), mrtrace.KeyRequestID, "req-1")
	assert.Equal(t, "req-1", cm.ProcessID(ctx, mrtrace.KeyRequestID))
	assert.Equal(t, "req-1", mrcontext.RequestID(ctx))
	assert.Empty(t, cm.ProcessID(ctx, mrtrace.KeyProcessID))

	ctx = cm.WithGeneratedProcessID(ctx, mrtrace.KeyProcessID)
	assert.Equal(t, "generated-id", cm.ProcessID(ctx, mrtrace.KeyProcessID))

	assert.Empty(t, logger.all())
}

func TestContextManager_UnknownKey(t *testing.T) {
	t.Parallel()

	logger := &warnLogger{}
	cm := newContextManager(t, logger)
	ctx := t.Context()

	assert.Empty(t, cm.ProcessID(ctx, mrtrace.KeyTaskID))
	assert.Equal(t, ctx, cm.WithProcessID(ctx, mrtrace.KeyTaskID, "task-1"))
	assert.Equal(t, ctx, cm.WithGeneratedProcessID(ctx, mrtrace.KeyWorkerID))

	assert.Equal(
		t,
		[]string{
			"process key not found: " + mrtrace.KeyTaskID,
			"process key not found: " + mrtrace.KeyTaskID,
			"process key not found: " + mrtrace.KeyWorkerID,
		},
		logger.all(),
	)
}

func TestContextManager_Extract(t *testing.T) {
	t.Parallel()

	cm := newContextManager(t, &warnLogger{})

	ctx := cm.WithProcessID(t.Context(), mrtrace.KeyRequestID, "req-1")
	ctx = cm.WithProcessID(ctx, mrtrace.KeyProcessID, "proc-1")

	assert.Equal(t, "req-1", cm.ExtractCorrelationID(ctx))
	assert.Equal(t, []any{mrtrace.KeyRequestID, "req-1", mrtrace.KeyProcessID, "proc-1"}, cm.ExtractKeysValues(ctx))

	ctx = cm.WithProcessID(ctx, mrtrace.KeyCorrelationID, "corr-1")
	assert.Equal(t, "corr-1", cm.ExtractCorrelationID(ctx))
}
