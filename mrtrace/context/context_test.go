package context_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	mrcontext "github.com/mondegor/go-core/mrtrace/context"
)

type (
	idAccessor struct {
		name   string
		withID func(ctx context.Context, id string) context.Context
		getID  func(ctx context.Context) string
	}
)

func idAccessors() []idAccessor {
	return []idAccessor{
		{name: "correlation", withID: mrcontext.WithCorrelationID, getID: mrcontext.CorrelationID},
		{name: "process", withID: mrcontext.WithProcessID, getID: mrcontext.ProcessID},
		{name: "request", withID: mrcontext.WithRequestID, getID: mrcontext.RequestID},
		{name: "task", withID: mrcontext.WithTaskID, getID: mrcontext.TaskID},
		{name: "worker", withID: mrcontext.WithWorkerID, getID: mrcontext.WorkerID},
	}
}

func TestIDs(t *testing.T) {
	t.Parallel()

	for _, tt := range idAccessors() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			assert.Empty(t, tt.getID(ctx))

			ctx = tt.withID(ctx, "id-1")
			assert.Equal(t, "id-1", tt.getID(ctx))

			// новое значение перекрывает предыдущее только в дочернем контексте
			child := tt.withID(ctx, "id-2")
			assert.Equal(t, "id-2", tt.getID(child))
			assert.Equal(t, "id-1", tt.getID(ctx))
		})
	}
}

func TestIDs_Independent(t *testing.T) {
	t.Parallel()

	accessors := idAccessors()

	ctx := t.Context()
	for _, a := range accessors {
		ctx = a.withID(ctx, a.name)
	}

	for _, a := range accessors {
		assert.Equal(t, a.name, a.getID(ctx))
	}
}
