package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlocale/config"
)

func TestValidateLanguages(t *testing.T) {
	t.Parallel()

	// wantErr - фрагмент сообщения: он показывает, какая именно проверка отвергла код,
	// т.к. одна и та же запись может быть отвергнута и как неразбираемая,
	// и как неканоничная, и как повтор
	tests := []struct {
		name    string
		names   []string
		wantErr string
	}{
		{
			name:  "canonical codes",
			names: []string{"ru-RU", "en-US", "ru", "zh-Hans", "es-419"},
		},
		{
			name:  "empty list",
			names: nil,
		},
		{
			// NewBundle такую запись принимает, но пул отдаёт наружу канон ("en-US"),
			// и исходный код перестаёт совпадать с тем, что использует приложение
			name:    "underscore form is not canonical",
			names:   []string{"en_US"},
			wantErr: "language 'en_US' is not canonical: expected 'en-US'",
		},
		{
			name:    "region in lower case is not canonical",
			names:   []string{"ru-ru"},
			wantErr: "language 'ru-ru' is not canonical: expected 'ru-RU'",
		},
		{
			name:    "language in upper case is not canonical",
			names:   []string{"RU-RU"},
			wantErr: "language 'RU-RU' is not canonical: expected 'ru-RU'",
		},
		{
			// проверка доходит до конца списка, а не ограничивается первым кодом
			name:    "not canonical code in the middle",
			names:   []string{"ru-RU", "en_US", "fr-FR"},
			wantErr: "language 'en_US' is not canonical: expected 'en-US'",
		},
		{
			name:    "duplicate code is rejected",
			names:   []string{"ru-RU", "en-US", "ru-RU"},
			wantErr: "duplicate language name 'ru-RU'",
		},
		{
			// две записи одной локали до проверки дубликатов не доходят:
			// неканоничная из них отвергается раньше, о чём и сообщает ошибка
			name:    "duplicate in non-canonical form is rejected",
			names:   []string{"en-US", "en_US"},
			wantErr: "language 'en_US' is not canonical: expected 'en-US'",
		},
		{
			// код разбирается и канонически записан, но конкретный язык не называет
			name:    "undetermined code is rejected",
			names:   []string{"und"},
			wantErr: "language 'und' is undetermined",
		},
		{
			// регион неопределённость не снимает: язык по-прежнему не назван
			name:    "undetermined code with region is rejected",
			names:   []string{"und-RU"},
			wantErr: "language 'und-RU' is undetermined",
		},
		{
			// приватная запись разбирается в тег с той же неопределённой базой
			name:    "private use code is rejected",
			names:   []string{"x-private"},
			wantErr: "language 'x-private' is undetermined",
		},
		{
			// "root" - неканоничная запись того же "und", поэтому отвергается раньше,
			// чем дело доходит до проверки неопределённости
			name:    "root is rejected as not canonical",
			names:   []string{"root"},
			wantErr: "language 'root' is not canonical: expected 'und'",
		},
		{
			// неразбираемый код сообщается ошибкой разбора, а не неканоничности
			name:    "malformed code is rejected",
			names:   []string{"not-a-language-tag!!!"},
			wantErr: "error parsing language (name='not-a-language-tag!!!')",
		},
		{
			name:    "empty code is rejected",
			names:   []string{""},
			wantErr: "error parsing language (name='')",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateLanguages(tc.names)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
