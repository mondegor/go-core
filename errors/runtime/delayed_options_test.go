package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/errors/kind"
	"github.com/mondegor/go-core/errors/runtime"
)

// TestInitDelayedOptions_DoesNotOverrideExplicitOptions - InitDelayedOptions одноразовый и глобальный,
// поэтому все сценарии проверяются в одном тесте: до и после инициализации.
func TestInitDelayedOptions_DoesNotOverrideExplicitOptions(t *testing.T) {
	t.Parallel()

	const (
		explicitHint = "explicit"
		defaultHint  = "default"
	)

	withHint := func(value string) runtime.Option {
		return runtime.WithOnCreate(func(_ kind.Enum, _ error) any {
			return value
		})
	}

	beforeExplicit := runtime.NewDelayed(kind.Internal, "before explicit", withHint(explicitHint))
	beforeDefault := runtime.NewDelayed(kind.Internal, "before default")

	runtime.InitDelayedOptions(
		runtime.OptionsHandlerFunc(func(_ kind.Enum, _ string) []runtime.Option {
			return []runtime.Option{withHint(defaultHint)}
		}),
	)

	afterExplicit := runtime.NewDelayed(kind.System, "after explicit", withHint(explicitHint))
	afterDefault := runtime.NewDelayed(kind.System, "after default")

	type testCase struct {
		name  string
		proto runtime.ProtoError
		want  string
	}

	tests := []testCase{
		{name: "explicit option before init is kept", proto: beforeExplicit, want: explicitHint},
		{name: "default option applied on init", proto: beforeDefault, want: defaultHint},
		{name: "explicit option after init is kept", proto: afterExplicit, want: explicitHint},
		{name: "default option applied after init", proto: afterDefault, want: defaultHint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var hintProvider interface{ Hint() any }

			require.ErrorAs(t, tt.proto.New(), &hintProvider)
			assert.Equal(t, tt.want, hintProvider.Hint())
		})
	}
}
