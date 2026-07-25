package mrlocale_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/mondegor/go-core/mrlocale"
)

type (
	// stubProvider - провайдер, отдающий сообщение без перевода:
	// проверяется подбор локализатора, а не сама локализация.
	stubProvider struct{}
)

func (p stubProvider) Domains() []string {
	return []string{mrlocale.DefaultMessagesDomain, mrlocale.DefaultErrorsDomain}
}

func (p stubProvider) Localize(_ string, _ language.Tag, msg string, _ []any) string {
	return msg
}

// TestPool_Localizer - проверяет подбор локализатора по списку языков.
// Язык по умолчанию задан явно и не является первым в списке, поэтому промах
// матчера отличим от подбора первого языка.
func TestPool_Localizer(t *testing.T) {
	t.Parallel()

	pool := testPool(t)

	tests := []struct {
		name  string
		langs []language.Tag
		want  string
	}{
		{
			// матчер возвращает тег с расширением "en-u-rg-uszzzz",
			// поэтому подбор по тегу дал бы язык по умолчанию
			name:  "regional tag resolves to its base language",
			langs: []language.Tag{language.MustParse("en-US")},
			want:  "en",
		},
		{
			name:  "regional tag of another language",
			langs: []language.Tag{language.MustParse("fr-CA")},
			want:  "fr",
		},
		{
			name:  "exact tag",
			langs: []language.Tag{language.MustParse("en")},
			want:  "en",
		},
		{
			name:  "first supported language of the list wins",
			langs: []language.Tag{language.MustParse("es"), language.MustParse("en-GB")},
			want:  "en",
		},
		{
			// промах должен давать язык по умолчанию, а не первый язык списка (fr)
			name:  "unsupported language falls back to the default",
			langs: []language.Tag{language.MustParse("es")},
			want:  "de",
		},
		{
			name:  "empty list falls back to the default",
			langs: nil,
			want:  "de",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := pool.Localizer(tc.langs...)

			require.NotNil(t, got)
			assert.Equal(t, tc.want, got.Language())
		})
	}
}

// TestPool_LocalizerRegionalBundle - проверяет подбор локализатора, когда языки бандла
// заданы с регионом, как это сделано в приложениях ("ru-RU", "en-US").
//
// Фиксирует главное свойство: наружу отдаётся язык бандла, а не язык запроса, поэтому
// любая запись одной и той же локали ("ru", "ru_RU", "ru-UA") сводится к "ru-RU".
// Язык по умолчанию задан вторым, чтобы промах матчера был отличим от подбора первого языка.
func TestPool_LocalizerRegionalBundle(t *testing.T) {
	t.Parallel()

	bundle, err := mrlocale.NewBundle(
		[]string{"ru-RU", "en-US"},
		mrlocale.WithDefaultLanguage("en-US"),
		mrlocale.WithMessageProvider(
			func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
		),
	)
	require.NoError(t, err)

	pool := mrlocale.NewPool(bundle)

	tests := []struct {
		name  string
		langs []language.Tag
		want  string
	}{
		{
			name:  "language without region resolves to the regional language of the bundle",
			langs: []language.Tag{language.MustParse("ru")},
			want:  "ru-RU",
		},
		{
			name:  "exact regional tag",
			langs: []language.Tag{language.MustParse("ru-RU")},
			want:  "ru-RU",
		},
		{
			name:  "underscore separated tag is the same language",
			langs: []language.Tag{language.MustParse("ru_RU")},
			want:  "ru-RU",
		},
		{
			// матчер отдаёт "ru-RU-u-rg-uazzzz", поэтому подбор по тегу дал бы промах
			name:  "another region of the supported language",
			langs: []language.Tag{language.MustParse("ru-UA")},
			want:  "ru-RU",
		},
		{
			name:  "another supported language without region",
			langs: []language.Tag{language.MustParse("en")},
			want:  "en-US",
		},
		{
			name:  "another region of another supported language",
			langs: []language.Tag{language.MustParse("en-GB")},
			want:  "en-US",
		},
		{
			name:  "unsupported language falls back to the default",
			langs: []language.Tag{language.MustParse("fr")},
			want:  "en-US",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := pool.Localizer(tc.langs...)

			require.NotNil(t, got)
			assert.Equal(t, tc.want, got.Language())
		})
	}
}

// TestPool_LocalizerAcceptLanguage - проверяет подбор локализатора по списку языков,
// разобранному из заголовка Accept-Language: именно в таком виде языки приходят от клиента.
// Сам заголовок mrlocale не разбирает, это делает вызывающая сторона.
func TestPool_LocalizerAcceptLanguage(t *testing.T) {
	t.Parallel()

	bundle, err := mrlocale.NewBundle(
		[]string{"ru-RU", "en-US"},
		mrlocale.WithDefaultLanguage("en-US"),
		mrlocale.WithMessageProvider(
			func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
		),
	)
	require.NoError(t, err)

	pool := mrlocale.NewPool(bundle)

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "the most weighted language wins",
			header: "ru-RU,ru;q=0.9,en;q=0.8",
			want:   "ru-RU",
		},
		{
			name:   "language without region wins",
			header: "ru,en;q=0.8",
			want:   "ru-RU",
		},
		{
			// вес q=0 означает, что язык клиенту не подходит
			name:   "language rejected by zero weight is skipped",
			header: "ru;q=0,en;q=0.9",
			want:   "en-US",
		},
		{
			name:   "unsupported languages fall back to the default",
			header: "de-DE,fr;q=0.7",
			want:   "en-US",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			langs, _, err := language.ParseAcceptLanguage(tc.header)
			require.NoError(t, err)

			got := pool.Localizer(langs...)

			require.NotNil(t, got)
			assert.Equal(t, tc.want, got.Language())
		})
	}
}

// TestPool_LocalizerDefaultIsNotFirst - фиксирует, что язык по умолчанию берётся
// из настройки бандла, а не из позиции в списке языков.
func TestPool_LocalizerDefaultIsNotFirst(t *testing.T) {
	t.Parallel()

	pool := testPool(t)

	assert.Equal(t, "de", pool.Localizer().Language())
}

// testPool - создаёт Pool на языках [fr, en, de] с языком по умолчанию "de".
func testPool(t *testing.T) *mrlocale.Pool {
	t.Helper()

	bundle, err := mrlocale.NewBundle(
		[]string{"fr", "en", "de"},
		mrlocale.WithDefaultLanguage("de"),
		mrlocale.WithMessageProvider(
			func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
		),
	)
	require.NoError(t, err)

	return mrlocale.NewPool(bundle)
}

// testPoolNonCanonical - создаёт Pool на неканоничных записях ["en_US", "RU-ru"]:
// именно на них расхождение исходной записи и канона проявилось бы.
func testPoolNonCanonical(t *testing.T) *mrlocale.Pool {
	t.Helper()

	bundle, err := mrlocale.NewBundle(
		[]string{"en_US", "RU-ru"},
		mrlocale.WithMessageProvider(
			func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
		),
	)
	require.NoError(t, err)

	return mrlocale.NewPool(bundle)
}

func TestPool_LocalizerByCode(t *testing.T) {
	t.Parallel()

	pool := testPool(t)

	tests := []struct {
		name  string
		code  string
		want  string
		wants bool
	}{
		{
			name:  "registered code",
			code:  "en",
			want:  "en",
			wants: true,
		},
		{
			// язык по умолчанию отличим от промаха только флагом,
			// поэтому проверяется отдельно
			name:  "default language code",
			code:  "de",
			want:  "de",
			wants: true,
		},
		{
			// в отличие от Localizer подбор ближайшего языка не выполняется
			name:  "regional tag of registered language is not a match",
			code:  "en-US",
			wants: false,
		},
		{
			name:  "unknown code",
			code:  "ru",
			wants: false,
		},
		{
			name:  "empty code",
			code:  "",
			wants: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := pool.LocalizerByCode(tc.code)
			require.Equal(t, tc.wants, ok)

			if !tc.wants {
				assert.Nil(t, got)

				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tc.want, got.Language())
		})
	}
}

// TestPool_LocalizerByCode_RegionalLanguages - проверяет точное совпадение
// на бандле из региональных локалей: базовый язык совпадением не считается,
// а иная запись той же локали ("ru_RU") к канонической форме не приводится.
func TestPool_LocalizerByCode_RegionalLanguages(t *testing.T) {
	t.Parallel()

	bundle, err := mrlocale.NewBundle(
		[]string{"ru-RU", "en-US"},
		mrlocale.WithMessageProvider(
			func(_ []language.Tag) (mrlocale.MessageProvider, error) {
				return stubProvider{}, nil
			},
		),
	)
	require.NoError(t, err)

	pool := mrlocale.NewPool(bundle)

	tests := []struct {
		name  string
		code  string
		wants bool
	}{
		{
			name:  "exact regional code",
			code:  "ru-RU",
			wants: true,
		},
		{
			name:  "base language is not a match",
			code:  "ru",
			wants: false,
		},
		{
			name:  "underscore notation is not a match",
			code:  "ru_RU",
			wants: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := pool.LocalizerByCode(tc.code)
			assert.Equal(t, tc.wants, ok)

			if tc.wants {
				require.NotNil(t, got)
				assert.Equal(t, tc.code, got.Language())
			}
		})
	}
}

// TestPool_Languages - проверяет, что список отдаётся в порядке бандла.
func TestPool_Languages(t *testing.T) {
	t.Parallel()

	// язык по умолчанию (de) задан не первым, но на порядок списка не влияет
	assert.Equal(t, []string{"fr", "en", "de"}, testPool(t).Languages())
}

// TestPool_Languages_CanonicalCodes - проверяет, что коды приведены к канонической записи:
// неканоничная запись из конфигурации ("en_US", "RU-ru") в списке уже нормализована.
func TestPool_Languages_CanonicalCodes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"en-US", "ru-RU"}, testPoolNonCanonical(t).Languages())
}

// TestPool_Languages_MatchLocalizerByCode - проверяет, что список и точный подбор опираются
// на одни и те же коды: каждый код списка принимается LocalizerByCode. Бандл задан
// неканонично, т.к. именно на нём расхождение записи проявилось бы.
func TestPool_Languages_MatchLocalizerByCode(t *testing.T) {
	t.Parallel()

	pool := testPoolNonCanonical(t)

	codes := pool.Languages()
	require.NotEmpty(t, codes) // иначе проверка выродилась бы в пустой цикл

	for _, code := range codes {
		_, ok := pool.LocalizerByCode(code)
		assert.True(t, ok, code)
	}
}

// TestPool_Languages_ReturnsCopy - проверяет, что выданный срез принадлежит вызывающему:
// его изменение на состав языков пула не влияет.
func TestPool_Languages_ReturnsCopy(t *testing.T) {
	t.Parallel()

	pool := testPool(t)

	got := pool.Languages()
	got[0] = "xx"

	assert.Equal(t, []string{"fr", "en", "de"}, pool.Languages())
}
