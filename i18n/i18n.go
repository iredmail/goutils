package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
)

var locales Locales

func New(embedLocales embed.FS) (err error) {
	paths, err := fs.Glob(embedLocales, "*/*_*")
	if err != nil {
		return
	}

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

func Translate(lang, s string, args ...any) string {
	locale, ok := locales[lang]
	if !ok {
		return fmt.Sprintf(s, args...)
	}

	return locale.Get(s, args...)
}
