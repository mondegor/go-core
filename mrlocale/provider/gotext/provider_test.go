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
