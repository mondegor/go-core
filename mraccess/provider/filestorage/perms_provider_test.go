package filestorage_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mraccess/provider/filestorage"
)

// writeRole - создаёт YAML-файл роли с указанными привилегиями и разрешениями.
func writeRole(t *testing.T, dir, name string, privileges, permissions []string) {
	t.Helper()

	var b strings.Builder

	b.WriteString("privileges:\n")

	for _, p := range privileges {
		b.WriteString("  - " + p + "\n")
	}

	b.WriteString("permissions:\n")

	for _, p := range permissions {
		b.WriteString("  - " + p + "\n")
	}

	path := filepath.Join(dir, name+".yaml")
	require.NoError(t, os.WriteFile(path, []byte(b.String()), 0o600))
}

func TestNewPermsProvider(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeRole(t, dir, "admin", []string{"users.manage"}, []string{"orders.edit"})
	writeRole(t, dir, "manager", []string{"orders.manage"}, nil)

	provider, err := filestorage.NewPermsProvider(dir, []string{"admin", "manager"}, []string{"any-user", "everyone"})
	require.NoError(t, err)

	// права роли = объединение её привилегий, разрешений и системных разрешений,
	// переданных в NewPermsProvider и выдаваемых каждой роли автоматически
	rights, ok := provider.RoleRights("admin")
	require.True(t, ok)
	assert.ElementsMatch(
		t,
		[]string{"users.manage", "orders.edit", "any-user", "everyone"},
		rights,
	)

	_, ok = provider.RoleRights("unknown-role")
	assert.False(t, ok)
}

func TestPermsProvider_IsRegistered(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeRole(t, dir, "admin", []string{"users.manage"}, []string{"orders.edit"})
	writeRole(t, dir, "manager", []string{"orders.manage"}, nil)

	provider, err := filestorage.NewPermsProvider(dir, []string{"admin", "manager"}, []string{"any-user", "everyone"})
	require.NoError(t, err)

	type testCase struct {
		name  string
		right string
		want  bool
	}

	tests := []testCase{
		{"privilege of admin", "users.manage", true},
		{"permission of admin", "orders.edit", true},
		{"privilege of manager", "orders.manage", true},
		{"system permission any-user", "any-user", true},
		{"system permission everyone", "everyone", true},
		{"unregistered right (fail-closed)", "billing.manage", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, provider.IsRegistered(tt.right))
		})
	}
}

func TestNewPermsProvider_Errors(t *testing.T) {
	t.Parallel()

	t.Run("empty roles", func(t *testing.T) {
		t.Parallel()

		_, err := filestorage.NewPermsProvider(t.TempDir(), nil, nil)
		require.ErrorContains(t, err, "roles is required")
	})

	t.Run("duplicate role", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeRole(t, dir, "admin", []string{"x"}, nil)

		_, err := filestorage.NewPermsProvider(dir, []string{"admin", "admin"}, nil)
		require.ErrorContains(t, err, "duplicate role")
	})

	t.Run("missing role file", func(t *testing.T) {
		t.Parallel()

		_, err := filestorage.NewPermsProvider(t.TempDir(), []string{"ghost"}, nil)
		require.Error(t, err)
	})
}
