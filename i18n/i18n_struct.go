package i18n

import (
	"io/fs"

	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

func NewFS(fsLocales fs.FS, supportedLanguages []any) (i *I18N, err error) {
	i = &I18N{
		supportedLanguages: supportedLanguages,
	}

	i.bundle, err = spreak.NewBundle(
		spreak.WithDomainFs(spreak.NoDomain, fsLocales),
		spreak.WithLanguage(i.supportedLanguages...),
	)

	return
}

func NewLocalePath(pathLocales string, supportedLanguages []any) (i *I18N, err error) {
	i = &I18N{
		pathLocales:        pathLocales,
		supportedLanguages: supportedLanguages,
	}

	i.bundle, err = spreak.NewBundle(
		spreak.WithDomainPath(spreak.NoDomain, pathLocales),
		spreak.WithLanguage(i.supportedLanguages...),
	)

	return
}

type I18N struct {
	pathLocales        string
	supportedLanguages []any
	bundle             *spreak.Bundle
}

func (i *I18N) ReloadByLocalePath() (err error) {
	i.bundle, err = spreak.NewBundle(
		spreak.WithDomainPath(spreak.NoDomain, i.pathLocales),
		spreak.WithLanguage(i.supportedLanguages...),
	)

	return
}

func (i *I18N) HasLocale(lang string) bool {
	if i.bundle == nil {
		return false
	}

	return spreak.NewKeyLocalizer(i.bundle, lang).HasLocale()
}

func (i *I18N) IsLanguageSupported(lang string) bool {
	if i.bundle == nil {
		return false
	}

	tag, _, err := language.ParseAcceptLanguage(lang)
	if err != nil {
		return false
	}

	return i.bundle.IsLanguageSupported(tag[0])
}

func (i *I18N) Translate(lang string, s string) string {
	if i.bundle == nil {
		return s
	}

	t := spreak.NewKeyLocalizer(i.bundle, lang)

	return t.Get(s)
}

func (i *I18N) TranslateF(lang string, s string, args ...any) string {
	if i.bundle == nil {
		return s
	}

	t := spreak.NewKeyLocalizer(i.bundle, lang)

	return t.Getf(s, args...)
}
