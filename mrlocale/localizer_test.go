package mrlocale_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/mondegor/go-core/mrlocale"
)

type (
	// echoProvider - провайдер, возвращающий домен, язык, сообщение и аргументы
	// одной строкой: проверяется, что именно бандл передаёт провайдеру.
	echoProvider struct {
		domains []string
	}
)

func (p echoProvider) Domains() []string {
	return p.domains
}

func (p echoProvider) Localize(domain string, lang language.Tag, msg string, args []any) string {
	return fmt.Sprintf("%s|%s|%s|%v", domain, lang, msg, args)
}

func newEchoLocalizer(t *testing.T, opts ...mrlocale.BundleOption) *mrlocale.Localizer {
	t.Helper()

	provider := echoProvider{domains: []string{"messages", "errors", "dict", "msg2", "err2"}}

	bundle, err := mrlocale.NewBundle(
		[]string{"ru-RU", "en"},
		func(_ []language.Tag) (mrlocale.MessageProvider, error) {
			return provider, nil
		},
		opts...,
	)
	require.NoError(t, err)

	return mrlocale.NewPool(bundle).Localizer()
}

func TestLocalizer_Translate(t *testing.T) {
	t.Parallel()

	localizer := newEchoLocalizer(t)

	assert.Equal(t, "ru-RU", localizer.Language())
	assert.Equal(t, "messages|ru-RU|Hello %s|[Ivan]", localizer.Translate("Hello %s", "Ivan"))
	assert.Equal(t, "dict|ru-RU|key|[]", localizer.TranslateInDomain("dict", "key"))
	assert.Equal(t, "errors|ru-RU|error text|[]", localizer.TranslateError(errors.New("error text")))
}

func TestLocalizer_CustomDomains(t *testing.T) {
	t.Parallel()

	localizer := newEchoLocalizer(
		t,
		mrlocale.WithMessagesDomain("msg2"),
		mrlocale.WithErrorsDomain("err2"),
	)

	assert.Equal(t, "msg2|ru-RU|msg|[]", localizer.Translate("msg"))
	assert.Equal(t, "err2|ru-RU|err|[]", localizer.TranslateError(errors.New("err")))
}

func TestLocalizer_FormatOptions(t *testing.T) {
	t.Parallel()

	localizer := newEchoLocalizer(
		t,
		mrlocale.WithFormatMessage(func(msg string, args []any) (string, []any) {
			return "formatted: " + msg, append(args, "extra")
		}),
		mrlocale.WithFormatError(func(err error) (string, []any) {
			return "code:" + err.Error(), []any{"arg"}
		}),
	)

	assert.Equal(t, "messages|ru-RU|formatted: msg|[1 extra]", localizer.Translate("msg", 1))
	assert.Equal(t, "dict|ru-RU|formatted: key|[extra]", localizer.TranslateInDomain("dict", "key"))

	// сообщение ошибки тоже проходит через formatMessage
	assert.Equal(t, "errors|ru-RU|formatted: code:err|[arg extra]", localizer.TranslateError(errors.New("err")))
}
