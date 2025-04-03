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

const (
	// domainDefault 为程序 embed 里提供的默认语言包。
	domainDefault = "default"

	// domainCustom 为系统管理员自行翻译后放入指定目录的语言包。
	// 程序应该优先选用自定义语言包。
	domainCustom = "custom"
)

var (
	bundle                  *spreak.Bundle
	supportedLocalLanguages []string
)

func Init(fsLocales fs.FS, supportedLanguages ...any) (err error) {
	bundle, err = spreak.NewBundle(
		spreak.WithDomainFs(domainDefault, fsLocales),
		spreak.WithLanguage(supportedLanguages...),
	)

	return err
}

// InitFSAndPath 同时从 fs.FS 和指定目录加在语言包。
func InitFSAndPath(fsLocales fs.FS, supportedLanguages []string, localesPath string) (localLanguages []string, err error) {
	opts := []spreak.BundleOption{
		spreak.WithDomainFs(domainDefault, fsLocales),
	}

	for _, l := range supportedLanguages {
		opts = append(opts, spreak.WithLanguage(l))
	}

	// Load path locales
	if len(localesPath) == 0 {
		return
	}

	if !goutils.DestExists(localesPath) {
		return
	}

	supportedLocalLanguages, err = walkLocaleDirPath(localesPath)
	if err != nil {
		return
	}

	localLanguages = append(localLanguages, supportedLocalLanguages...)
	opts = append(opts, spreak.WithDomainPath(domainCustom, localesPath))
	for _, localLanguage := range supportedLocalLanguages {
		opts = append(opts, spreak.WithLanguage(localLanguage))
	}

	supportedLocalLanguages = append(supportedLocalLanguages, supportedLocalLanguages...)
	bundle, err = spreak.NewBundle(opts...)

	return
}

func walkLocaleDirPath(localesPath string) (localLanguages []string, err error) {
	err = filepath.WalkDir(localesPath, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
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

	t := spreak.NewKeyLocalizerForDomain(bundle, domainCustom, lang)
	if t.HasDomain(domainCustom) {
		return t.Get(s)
	}

	return spreak.NewKeyLocalizerForDomain(bundle, domainDefault, lang).Get(s)
}

func TranslateF(lang string, s string, args ...any) string {
	if bundle == nil {
		return s
	}

	var t *spreak.KeyLocalizer
	if slices.Contains(supportedLocalLanguages, lang) {
		t = spreak.NewKeyLocalizerForDomain(bundle, domainCustom, lang)
	} else {
		t = spreak.NewKeyLocalizerForDomain(bundle, domainDefault, lang)
	}

	return t.Getf(s, args...)
}
