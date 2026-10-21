package gotext_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"

	"github.com/mondegor/go-core/mrlocale/provider/gotext"
)

// TestProvider_Domains_ReturnsCopy - выданный срез принадлежит вызывающему:
// его изменение на состав доменов провайдера не влияет.
//
// Порядок доменов не проверяется: список собирается обходом карты каталогов
// и потому от запуска к запуску различается (см. NewProvider).
func TestProvider_Domains_ReturnsCopy(t *testing.T) {
	t.Parallel()

	provider, err := gotext.NewProvider(
		[]language.Tag{language.English},
		gotext.WithDomainCatalog("messages", catalog.NewBuilder()),
		gotext.WithDomainCatalog("errors", catalog.NewBuilder()),
	)
	require.NoError(t, err)

	want := []string{"messages", "errors"}

	got := provider.Domains()
	got[0] = "unknown"

	assert.ElementsMatch(t, want, provider.Domains())
}

// TestProvider_Domains_EmptyIsNil - провайдер без каталогов отдаёт пустой список
// как nil, а не как пустой срез.
func TestProvider_Domains_EmptyIsNil(t *testing.T) {
	t.Parallel()

	provider, err := gotext.NewProvider([]language.Tag{language.English})
	require.NoError(t, err)

	assert.Nil(t, provider.Domains())
}

// TestNewProvider_EmptyLanguages - провайдер без языков не создаётся.
func TestNewProvider_EmptyLanguages(t *testing.T) {
	t.Parallel()

	provider, err := gotext.NewProvider(nil, gotext.WithDomainCatalog("messages", catalog.NewBuilder()))
	require.ErrorContains(t, err, "languages is required")
	assert.Nil(t, provider)
}

// TestProvider_Localize - перевод по домену и языку; вне зарегистрированных
// доменов и языков сообщение только форматируется.
func TestProvider_Localize(t *testing.T) {
	t.Parallel()

	ru := language.MustParse("ru-RU")

	messages := catalog.NewBuilder()
	require.NoError(t, messages.SetString(language.English, "Hello %[1]s", "Hello, %[1]s!"))
	require.NoError(t, messages.SetString(ru, "Hello %[1]s", "Привет, %[1]s!"))

	errs := catalog.NewBuilder()
	require.NoError(t, errs.SetString(ru, "Hello %[1]s", "Ошибка: %[1]s"))

	provider, err := gotext.NewProvider(
		[]language.Tag{language.English, ru},
		gotext.WithDomainCatalog("messages", messages),
		gotext.WithDomainCatalog("errors", errs),
		gotext.WithDomainCatalog("ignored", nil),
	)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"messages", "errors"}, provider.Domains())

	type testCase struct {
		name   string
		domain string
		lang   language.Tag
		want   string
	}

	tests := []testCase{
		{name: "english", domain: "messages", lang: language.English, want: "Hello, Ivan!"},
		{name: "russian", domain: "messages", lang: ru, want: "Привет, Ivan!"},
		{name: "another domain", domain: "errors", lang: ru, want: "Ошибка: Ivan"},
		{name: "no translation in domain", domain: "errors", lang: language.English, want: "Hello Ivan"},
		{name: "unknown domain: formatted without translation", domain: "dict", lang: ru, want: "Hello Ivan"},
		{name: "unknown language: formatted without translation", domain: "messages", lang: language.German, want: "Hello Ivan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, provider.Localize(tt.domain, tt.lang, "Hello %[1]s", []any{"Ivan"}))
		})
	}
}

// TestProvider_LocalizeDuplicateDomain - повторный домен заменяет каталог, заданный ранее.
func TestProvider_LocalizeDuplicateDomain(t *testing.T) {
	t.Parallel()

	first := catalog.NewBuilder()
	require.NoError(t, first.SetString(language.English, "key", "first"))

	second := catalog.NewBuilder()
	require.NoError(t, second.SetString(language.English, "key", "second"))

	provider, err := gotext.NewProvider(
		[]language.Tag{language.English},
		gotext.WithDomainCatalog("messages", first),
		gotext.WithDomainCatalog("messages", second),
	)
	require.NoError(t, err)

	assert.Equal(t, []string{"messages"}, provider.Domains())
	assert.Equal(t, "second", provider.Localize("messages", language.English, "key", nil))
}
