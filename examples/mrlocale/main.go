package main

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/errors/helper"
	"github.com/mondegor/go-core/examples/mrlocale/internal/dict/errcat"
	"github.com/mondegor/go-core/examples/mrlocale/internal/dict/fruitcat"
	"github.com/mondegor/go-core/examples/mrlocale/internal/dict/msgcat"
	"github.com/mondegor/go-core/mrlocale"
	"github.com/mondegor/go-core/mrlocale/config"
	"github.com/mondegor/go-core/mrlocale/provider/gotext"
)

// main - пример формирование сообщений, ошибок и свойств объектов на указанном языке.
func main() {
	// языки из конфигурации: годятся для создания бандла, но не как источник списка
	// языков приложения - для него есть Pool.Languages (см. обход локализаторов ниже)
	languages := []string{"ru-RU", "en-US"}

	// неканоничную запись ("ru_RU") бандл принял бы, но наружу пул отдавал бы канон,
	// поэтому список из конфигурации проверяется до создания бандла
	if err := config.ValidateLanguages(languages); err != nil {
		panic(err)
	}

	bundle, err := mrlocale.NewBundle(
		languages,
		mrlocale.WithDefaultLanguage("ru-RU"),
		mrlocale.WithFormatError(helper.ExtractMessageForLocalization),
		mrlocale.WithMessageProvider(
			func(languages []language.Tag) (mrlocale.MessageProvider, error) {
				return gotext.NewProvider(
					languages,
					gotext.WithDomainCatalog(mrlocale.DefaultMessagesDomain, msgcat.NewCatalog()),
					gotext.WithDomainCatalog(mrlocale.DefaultErrorsDomain, errcat.NewCatalog()),
					gotext.WithDomainCatalog("fruitcat", fruitcat.NewCatalog()),
				)
			},
		),
	)
	if err != nil {
		panic(err)
	}

	pool := mrlocale.NewPool(bundle)
	lz := pool.Localizer(language.MustParse("ru-RU"))

	fmt.Printf("language: %s\n", lz.Language())
	fmt.Printf("message: %s\n", lz.Translate("Message example"))

	fmt.Println("--------------------------------------------------")

	errorMessage := errcat.ParseErrorMessage(lz.TranslateError(errors.NewInternalError("my-error")))
	fmt.Printf("error: %s\n", errorMessage.Reason)
	fmt.Printf("error-details: %s\n", errorMessage.Details)

	fmt.Println("--------------------------------------------------")

	// список языков берётся у пула, а не у конфигурации: коды в нём канонические,
	// поэтому клиенту отдаются ровно те коды, которые принимает LocalizerByCode
	for _, code := range pool.Languages() {
		// код взят у самого пула, поэтому точный подбор заведомо удачен
		lz, _ := pool.LocalizerByCode(code)

		fmt.Printf("language: %s\n", code)
		fmt.Println("..................................................")

		fmt.Println(lz.Translate("Message example with param %s", "param1"))
		fmt.Println(lz.Translate("Message example with two params %[1]s and %[2]s", "param1", "param2"))
		fmt.Println(lz.Translate("Total %d message(s)", 0))
		fmt.Println(lz.Translate("Total %d message(s)", 1))
		fmt.Println(lz.Translate("Total %d message(s)", 21))
		fmt.Println(lz.Translate("Total %d message(s)", 43))
		fmt.Println(lz.Translate("Total %d message(s)", 67))

		fmt.Println("..................................................")

		errMess := errcat.ParseErrorMessage(lz.TranslateError(errors.ErrInternalNilPointer.New()))
		fmt.Println("reason: '" + errMess.Reason + "'; details: '" + errMess.Details + "'")

		errMess = errcat.ParseErrorMessage(lz.TranslateError(errors.ErrSystemStorageConnectionFailed.New()))
		fmt.Println("reason: '" + errMess.Reason + "'; details: '" + errMess.Details + "'")

		fmt.Println(lz.TranslateError(errors.ErrRecordNotFound))

		fmt.Println("..................................................")

		fruitMess := fruitcat.ParseMessage(lz.TranslateInDomain("fruitcat", "id7391"))
		fmt.Println("ID: '" + fruitMess.ID + "'; Name: '" + fruitMess.Name + "'; Description: '" + fruitMess.Description + "'")

		fruitMess = fruitcat.ParseMessage(lz.TranslateInDomain("fruitcat", "id97104"))
		fmt.Println("ID: '" + fruitMess.ID + "'; Name: '" + fruitMess.Name + "'; Description: '" + fruitMess.Description + "'")

		fmt.Println("--------------------------------------------------")
	}

	// fr - not found
	lz = pool.Localizer(language.MustParse("fr"))
	fmt.Println(lz.Language()) // print default lang
}
