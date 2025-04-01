package i18n

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/iredmail/goutils"
	"github.com/iredmail/goutils/slice"
	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

var (
	bundleFS                *spreak.Bundle
	bundlePath              *spreak.Bundle
	supportedLocalLanguages []string
)

func Init(fsLocales fs.FS, supportedLanguages ...any) (err error) {
	bundleFS, err = spreak.NewBundle(
		spreak.WithDomainFs(spreak.NoDomain, fsLocales),
		spreak.WithLanguage(supportedLanguages...),
	)

	return err
}

func InitFSAndPath(fsLocales fs.FS, supportedLanguages []string, localesPath ...string) (localLanguages []string, err error) {
	opts := []spreak.BundleOption{
		spreak.WithDomainFs(spreak.NoDomain, fsLocales),
	}

	for _, l := range supportedLanguages {
		opts = append(opts, spreak.WithLanguage(l))
	}

	bundleFS, err = spreak.NewBundle(opts...)
	if err != nil {
		return
	}

	// Load path locales
	if len(localesPath) == 0 {
		return
	}

	if !goutils.DestExists(localesPath[0]) {
		return
	}

	supportedLocalLanguages, err = walkLocaleDirPath(localesPath[0])
	if err != nil {
		return
	}

	localLanguages = append(localLanguages, supportedLocalLanguages...)
	opts = []spreak.BundleOption{
		spreak.WithDomainPath(spreak.NoDomain, localesPath[0]),
	}

	for _, localLanguage := range supportedLocalLanguages {
		opts = append(opts, spreak.WithLanguage(localLanguage))
	}

	bundlePath, err = spreak.NewBundle(opts...)

	return
}

func walkLocaleDirPath(localesPath string) (localLanguages []string, err error) {
	err = filepath.WalkDir(localesPath, func(path string, d fs.DirEntry, err error) error {
		if d != nil && d.IsDir() {
			return nil
		}

		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		lang := strings.TrimSuffix(d.Name(), ".json")
		_, _, err = language.ParseAcceptLanguage(lang)
		if err == nil {
			localLanguages = slice.AddMissingElems(localLanguages, lang)
		}

		return nil
	})

	return
}

func IsLanguageSupported(lang string) bool {
	tag, _, err := language.ParseAcceptLanguage(lang)
	if err != nil {
		return false
	}

	if bundlePath != nil {
		if bundlePath.IsLanguageSupported(tag[0]) {
			return true
		}
	}

	if bundleFS != nil {
		return bundleFS.IsLanguageSupported(tag[0])
	}

	return false
}

func Translate(lang string, s string) string {
	if bundleFS == nil && bundlePath == nil {
		return s
	}

	var t *spreak.KeyLocalizer
	if slices.Contains(supportedLocalLanguages, lang) {
		t = spreak.NewKeyLocalizer(bundlePath, lang)
	} else {
		t = spreak.NewKeyLocalizer(bundleFS, lang)
	}

	return t.Get(s)
}

func TranslateF(lang string, s string, args ...any) string {
	if bundleFS == nil && bundlePath == nil {
		return s
	}

	var t *spreak.KeyLocalizer
	if slices.Contains(supportedLocalLanguages, lang) {
		t = spreak.NewKeyLocalizer(bundlePath, lang)
	} else {
		t = spreak.NewKeyLocalizer(bundleFS, lang)
	}

	return t.Getf(s, args...)
}
