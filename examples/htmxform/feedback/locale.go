package feedback

import (
	_ "embed" // for translation files

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed active.ru.json
var russianTranslation []byte

func AddLocalizationMessages(b *i18n.Bundle) (err error) {
	_, err = b.ParseMessageFileBytes(russianTranslation, "active.ru.json")
	if err != nil {
		return err
	}
	return nil
}
