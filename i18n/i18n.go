package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
)

func New(embedLocales embed.FS) (Locales, error) {
	paths, err := fs.Glob(embedLocales, "*/*_*")
	if err != nil {
		return nil, err
	}

	locales := Locales{}
	for _, pth := range paths {
		lang := filepath.Base(pth)
		l := NewFSLocale(embedLocales, pth, lang)

		err = l.AddDomain("messages")
		if err != nil {
			return nil, err
		}

		locales[lang] = l
	}

	return locales, nil
}

type Locales map[string]*Locale

func (l Locales) Translate(lang, s string, args ...any) string {
	locale, ok := l[lang]
	if !ok {
		return fmt.Sprintf(s, args...)
	}

	return locale.Get(s, args...)
}
