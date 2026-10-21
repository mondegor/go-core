package config_test

// база часовых поясов встраивается в тестовый бинарник, чтобы тесты
// проходили в минимальных образах, где она отсутствует в системе:
// без неё валидатор отверг бы все IANA-имена.
import (
	"testing"
	_ "time/tzdata"

	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/util/timezone/config"
)

func TestValidateTimeZones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		names   []string
		wantErr bool
	}{
		{
			name:  "unique names",
			names: []string{"Europe/Moscow", "Asia/Tokyo", "UTC"},
		},
		{
			// пояс процесса LocationList не обслуживает, поэтому в списке
			// он бесполезен и отвергается валидатором
			name:    "process timezone name is not allowed",
			names:   []string{"Local"},
			wantErr: true,
		},
		{
			name:  "empty list",
			names: nil,
		},
		{
			// имя проверяется по базе часовых поясов: опечатку в конфиге
			// должна отсекать именно эта проверка, так как
			// timezone.NewLocationList негодное имя молча пропустит
			name:    "unknown name",
			names:   []string{"Europe/Moscow", "Nowhere/Nowhere"},
			wantErr: true,
		},
		{
			name:    "empty name",
			names:   []string{"Europe/Moscow", ""},
			wantErr: true,
		},
		{
			name:    "duplicate name",
			names:   []string{"Europe/Moscow", "Asia/Tokyo", "Europe/Moscow"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateTimeZones(tt.names)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
