package mraccess_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mraccess"
)

// stubUserProvider - тестовый провайдер, возвращающий заданного пользователя или ошибку.
type stubUserProvider struct {
	user mraccess.User
	err  error
}

func (s stubUserProvider) UserByToken(_ context.Context, _ string) (mraccess.User, error) {
	return s.user, s.err
}

// typeByTokenPrefix - выбирает тип провайдера по префиксу токена "<type>:...".
func typeByTokenPrefix(token string) string {
	if i := strings.IndexByte(token, ':'); i > 0 {
		return token[:i]
	}

	return ""
}

func TestUserProviderGroup_UserByToken(t *testing.T) {
	t.Parallel()

	getter, err := mraccess.NewRolesGroupSet(testGroups(), testSource())
	require.NoError(t, err)

	dbUser := mraccess.NewUser([16]byte{1}, "administrators", "", "en", "UTC", getter)
	providerErr := errors.New("jwt provider failed")

	providers := []mraccess.TypedUserProvider{
		{Type: "db", Value: stubUserProvider{user: dbUser}},
		{Type: "jwt", Value: stubUserProvider{err: providerErr}},
	}

	group := mraccess.NewUserProviderGroup(providers, typeByTokenPrefix)

	type testCase struct {
		name      string
		token     string
		wantUser  mraccess.User
		wantErr   bool
		errSubstr string
	}

	tests := []testCase{
		{
			name:     "routed to db provider",
			token:    "db:abc",
			wantUser: dbUser,
		},
		{
			name:      "delegates provider error",
			token:     "jwt:abc",
			wantErr:   true,
			errSubstr: "jwt provider failed",
		},
		{
			name:      "empty token rejected (fail-closed)",
			token:     "",
			wantErr:   true,
			errSubstr: "token value is empty",
		},
		{
			name:      "unknown token type rejected (fail-closed)",
			token:     "plain-token-without-prefix",
			wantErr:   true,
			errSubstr: "provider not found",
		},
		{
			name:      "unregistered type rejected (fail-closed)",
			token:     "ldap:abc",
			wantErr:   true,
			errSubstr: "provider not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := group.UserByToken(t.Context(), tt.token)

			if tt.wantErr {
				require.ErrorContains(t, err, tt.errSubstr)
				assert.Nil(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, got)
		})
	}
}
