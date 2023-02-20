package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	"golang.org/x/exp/maps"
)

var locales Locales

func Init(embedLocales embed.FS) (err error) {
	if locales != nil {
		return
	}

	paths, err := fs.Glob(embedLocales, "*/*_*")
	if err != nil {
		return
	}

	locales = make(map[string]*Locale)
	for _, pth := range paths {
		lang := filepath.Base(pth)
		l := NewFSLocale(embedLocales, pth, lang)

		if err = l.AddDomain("messages"); err != nil {
			return
		}

		locales[lang] = l
	}

	return
}

type Locales map[string]*Locale

func Languages() []string {
	return maps.Keys(locales)
}

func Translate(lang, s string, args ...any) string {
	locale, ok := locales[lang]
	if !ok {
		return fmt.Sprintf(s, args...)
	}

	return locale.Get(s, args...)
}
