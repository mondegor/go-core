package mraccess_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mraccess"
)

// stubRoleSource - тестовый источник прав по роли (роль -> права).
// Реализует неэкспортируемый интерфейс roleRightsSource структурно.
type stubRoleSource map[string][]string

func (s stubRoleSource) RoleRights(role string) ([]string, bool) {
	rights, ok := s[role]

	return rights, ok
}

// testSource - общий статичный конфиг прав по ролям (init-время).
func testSource() stubRoleSource {
	return stubRoleSource{
		"admin":   {"users.manage", "orders.manage", "public"},
		"manager": {"orders.manage"},
		"guest":   {"public"},
	}
}

// testGroups - набор групп ролей поверх testSource.
func testGroups() []mraccess.RoleGroup {
	return []mraccess.RoleGroup{
		{Name: "administrators", Roles: []string{"admin"}},
		{Name: "managers", Roles: []string{"manager"}},
		{Name: "guests", Roles: []string{"guest"}},
		{Name: "staff", Roles: []string{"manager", "guest"}}, // объединение прав двух ролей
		{Name: "empty", Roles: nil},                          // известная группа без прав (A11)
	}
}

// TestRolesGroupSet_AccessTable - поведенческая таблица доступа (A15).
// Строки: группа вызывающего пользователя; столбцы: запрашиваемое право.
// Ожидание: granted (true) / denied (false). Включены fail-closed случаи.
func TestRolesGroupSet_AccessTable(t *testing.T) {
	t.Parallel()

	getter, err := mraccess.NewRolesGroupSet(testGroups(), testSource())
	require.NoError(t, err)

	type testCase struct {
		name  string
		group string
		right string
		want  bool
	}

	tests := []testCase{
		// administrators (роль admin)
		{"admin: own privilege granted", "administrators", "users.manage", true},
		{"admin: orders granted", "administrators", "orders.manage", true},
		{"admin: public granted", "administrators", "public", true},
		{"admin: unknown right denied", "administrators", "billing.manage", false},

		// managers (роль manager)
		{"manager: orders granted", "managers", "orders.manage", true},
		{"manager: users.manage denied", "managers", "users.manage", false},
		{"manager: public denied (нет в роли)", "managers", "public", false},

		// guests (роль guest)
		{"guest: public granted", "guests", "public", true},
		{"guest: orders denied", "guests", "orders.manage", false},

		// staff = union(manager, guest)
		{"staff: orders из manager granted", "staff", "orders.manage", true},
		{"staff: public из guest granted", "staff", "public", true},
		{"staff: users.manage denied", "staff", "users.manage", false},

		// empty: группа известна, но прав нет — отличаем от опечатки (A11)
		{"empty group: privilege denied", "empty", "orders.manage", false},
		{"empty group: public denied", "empty", "public", false},

		// fail-closed: неизвестная группа -> доступ запрещён всегда
		{"unknown group: public denied (fail-closed)", "robots", "public", false},
		{"unknown group: privilege denied (fail-closed)", "robots", "orders.manage", false},

		// пустое имя права
		{"empty right name denied", "administrators", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getter.Rights(tt.group).Has(tt.right)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewRolesGroupSet_NonexistentRoleReturnsError(t *testing.T) {
	t.Parallel()

	_, err := mraccess.NewRolesGroupSet(
		[]mraccess.RoleGroup{{Name: "g", Roles: []string{"ghost"}}},
		testSource(),
	)

	require.ErrorContains(t, err, "ghost")
}

func TestNewRolesGroupSet_EmptyGroupIsKnown(t *testing.T) {
	t.Parallel()

	getter, err := mraccess.NewRolesGroupSet(
		[]mraccess.RoleGroup{{Name: "empty", Roles: nil}},
		testSource(),
	)
	require.NoError(t, err)

	// известная группа: не падает и не выдаёт прав
	assert.False(t, getter.Rights("empty").Has("public"))
}

func TestNewRolesGroupSet_NoGroups(t *testing.T) {
	t.Parallel()

	getter, err := mraccess.NewRolesGroupSet(nil, testSource())
	require.NoError(t, err)

	// любая группа неизвестна -> доступ запрещён (fail-closed)
	assert.False(t, getter.Rights("administrators").Has("public"))
}
