package mrlocale_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/mondegor/go-core/mrlocale"
)

// TestNewBundle_DefaultLanguage - проверяет выбор языка по умолчанию.
// Язык сверяется со списком поддерживаемых по разобранному тегу, поэтому
// одна и та же локаль, записанная по-разному ("ru-RU", "ru_RU", "ru-ru"),
// опознаётся как один и тот же язык.
func TestNewBundle_DefaultLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		languages []string
		def       string
		want      string
	}{
		{
			name:      "default is not specified, first language of the list wins",
			languages: []string{"ru-RU", "en-US"},
			def:       "",
			want:      "ru-RU",
		},
		{
			name:      "default matches the list entry byte to byte",
			languages: []string{"ru-RU", "en-US"},
			def:       "en-US",
			want:      "en-US",
		},
		{
			// подчёркивание - разделитель, который порождает генератор gotext,
			// поэтому такая запись попадает в конфигурацию и должна опознаваться
			name:      "underscore separated default matches hyphen separated list",
			languages: []string{"ru-RU", "en-US"},
			def:       "en_US",
			want:      "en-US",
		},
		{
			name:      "hyphen separated default matches underscore separated list",
			languages: []string{"ru_RU", "en_US"},
			def:       "en-US",
			want:      "en-US",
		},
		{
			name:      "default is case insensitive",
			languages: []string{"ru-RU", "en-US"},
			def:       "en-us",
			want:      "en-US",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []mrlocale.BundleOption

			if tc.def != "" {
				opts = append(opts, mrlocale.WithDefaultLanguage(tc.def))
			}

			bundle, err := mrlocale.NewBundle(tc.languages, newStubProvider, opts...)
			require.NoError(t, err)

			assert.Equal(t, tc.want, mrlocale.NewPool(bundle).Localizer().Language())
		})
	}
}

// TestNewBundle_EmptyLanguages - проверяет подмену пустого списка языков.
// Пустой список - это не отказ: вместо него подставляется "en-US", и он становится
// и единственным языком списка, и языком по умолчанию. Поэтому WithDefaultLanguage
// здесь годится только для того же "en-US" (см. TestNewBundle_Errors).
func TestNewBundle_EmptyLanguages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		def  string
		want string
	}{
		{
			name: "default is not specified, the substituted language wins",
			def:  "",
			want: "en-US",
		},
		{
			name: "default matches the substituted language",
			def:  "en-US",
			want: "en-US",
		},
		{
			// подставленный язык - обычное имя списка, поэтому язык по умолчанию
			// сверяется с ним по разобранному тегу, а наружу выходит уже канон
			name: "underscore separated default matches the substituted language",
			def:  "en_US",
			want: "en-US",
		},
		{
			name: "case insensitive default matches the substituted language",
			def:  "en-us",
			want: "en-US",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []mrlocale.BundleOption

			if tc.def != "" {
				opts = append(opts, mrlocale.WithDefaultLanguage(tc.def))
			}

			bundle, err := mrlocale.NewBundle(nil, newStubProvider, opts...)
			require.NoError(t, err)

			pool := mrlocale.NewPool(bundle)

			assert.Equal(t, tc.want, pool.Localizer().Language())
			assert.Equal(t, []string{tc.want}, pool.Languages())
		})
	}
}

// TestNewBundle_Errors - проверяет отказы при создании бандла.
func TestNewBundle_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		languages []string
		def       string
	}{
		{
			name:      "language of the list is not well-formed",
			languages: []string{"ru-RU", "garbage!!!"},
		},
		{
			// регион EN не существует, поэтому такой тег отвергается
			name:      "region of the list language is unknown",
			languages: []string{"ru-RU", "en-EN"},
		},
		{
			name:      "duplicate language in the list",
			languages: []string{"ru-RU", "en-US", "ru-RU"},
		},
		{
			// разная запись одной локали - это тот же язык, поэтому тоже дубликат
			name:      "duplicate language written differently",
			languages: []string{"ru-RU", "ru_RU"},
		},
		{
			name:      "default language is not well-formed",
			languages: []string{"ru-RU", "en-US"},
			def:       "garbage!!!",
		},
		{
			name:      "default language is not in the list",
			languages: []string{"ru-RU", "en-US"},
			def:       "de-DE",
		},
		{
			// язык без региона не равен языку с регионом, поэтому не опознаётся
			name:      "default language without region does not match regional list entry",
			languages: []string{"ru-RU", "en-US"},
			def:       "en",
		},
		{
			// вместо пустого списка подставляется "en-US", поэтому язык
			// по умолчанию сверяется с ним и любой другой язык не подходит
			name:      "empty list with default language other than the substituted one",
			languages: nil,
			def:       "ru-RU",
		},
		{
			// негодная запись отвергается на разборе языка по умолчанию,
			// до сверки с подставленным языком дело не доходит
			name:      "empty list with not well-formed default language",
			languages: nil,
			def:       "garbage!!!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []mrlocale.BundleOption

			if tc.def != "" {
				opts = append(opts, mrlocale.WithDefaultLanguage(tc.def))
			}

			bundle, err := mrlocale.NewBundle(tc.languages, newStubProvider, opts...)

			require.Error(t, err)
			assert.Nil(t, bundle)
		})
	}
}

// TestNewBundle_UndefinedLanguageIsSupported - фиксирует, что "und" является
// обычным языком списка: он не должен трактоваться как признак незаданного
// языка по умолчанию.
func TestNewBundle_UndefinedLanguageIsSupported(t *testing.T) {
	t.Parallel()

	bundle, err := mrlocale.NewBundle(
		[]string{"ru-RU", "und"},
		newStubProvider,
		mrlocale.WithDefaultLanguage("und"),
	)
	require.NoError(t, err)

	assert.Equal(t, "und", mrlocale.NewPool(bundle).Localizer().Language())
}

// TestNewBundle_ProviderErrors - проверяет отказы, связанные с провайдером.
func TestNewBundle_ProviderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider func([]language.Tag) (mrlocale.MessageProvider, error)
		opts     []mrlocale.BundleOption
		wantErr  string
	}{
		{
			name: "provider has not messages domain",
			provider: func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return echoProvider{domains: []string{mrlocale.DefaultErrorsDomain}}, nil
			},
			wantErr: `messages domain "messages"`,
		},
		{
			name: "provider has not errors domain",
			provider: func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return echoProvider{domains: []string{mrlocale.DefaultMessagesDomain}}, nil
			},
			wantErr: `errors domain "errors"`,
		},
		{
			name: "provider has not custom messages domain",
			provider: func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
			opts:    []mrlocale.BundleOption{mrlocale.WithMessagesDomain("custom")},
			wantErr: `messages domain "custom"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bundle, err := mrlocale.NewBundle(
				[]string{"en"},
				tc.provider,
				tc.opts...,
			)

			require.ErrorContains(t, err, tc.wantErr)
			assert.Nil(t, bundle)
		})
	}
}

// TestNewBundle_ProviderFactoryFails - ошибка фабрики провайдера возвращается обёрнутой.
func TestNewBundle_ProviderFactoryFails(t *testing.T) {
	t.Parallel()

	errProvider := errors.New("provider failed")

	bundle, err := mrlocale.NewBundle(
		[]string{"en"},
		func(_ []language.Tag) (mrlocale.MessageProvider, error) {
			return nil, errProvider
		},
	)

	require.ErrorIs(t, err, errProvider)
	assert.Nil(t, bundle)
}

// TestNewBundle_ProviderGetsLanguages - фабрика провайдера получает разобранные языки бандла.
func TestNewBundle_ProviderGetsLanguages(t *testing.T) {
	t.Parallel()

	var got []language.Tag

	_, err := mrlocale.NewBundle(
		[]string{"ru_RU", "en"},
		func(languages []language.Tag) (mrlocale.MessageProvider, error) {
			got = languages

			return stubProvider{}, nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, []language.Tag{language.MustParse("ru-RU"), language.English}, got)
}
