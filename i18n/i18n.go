package i18n

import (
	"io/fs"

	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

var (
	bundle *spreak.Bundle
)

func Init(fsLocales fs.FS, supportedLanguages ...any) (err error) {
	bundle, err = spreak.NewBundle(
		spreak.WithDomainFs(spreak.NoDomain, fsLocales),
		spreak.WithLanguage(supportedLanguages...),
	)

	return err
}

func IsLanguageSupported(lang string) bool {
	if bundle == nil {
		return false
	}

	tag, _, err := language.ParseAcceptLanguage(lang)
	if err != nil {
		return false
	}

	return bundle.IsLanguageSupported(tag[0])
}

func Translate(lang string, s string) string {
	if bundle == nil {
		return s
	}

	t := spreak.NewKeyLocalizer(bundle, lang)

	return t.Get(s)
}

func TranslateF(lang string, s string, args ...any) string {
	if bundle == nil {
		return s
	}

	t := spreak.NewKeyLocalizer(bundle, lang)

	return t.Getf(s, args...)
}
