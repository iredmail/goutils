package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
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
	bundle          *spreak.Bundle
	customLanguages []string
)

func Init(fsLocales fs.FS, supportedLanguages ...any) (err error) {
	bundle, err = spreak.NewBundle(
		spreak.WithDomainFs(domainDefault, fsLocales),
		spreak.WithLanguage(supportedLanguages...),
	)

	return err
}

// InitFSAndDir 同时从 fs.FS 和指定目录加载 JSON 格式的语言翻译文件。
func InitFSAndDir(fsLocales fs.FS, supportedLanguages []string, localesDir string) (_customLanguages []string, errCustomLocales, err error) {
	opts := []spreak.BundleOption{
		spreak.WithDomainFs(domainDefault, fsLocales),
	}

	for _, l := range supportedLanguages {
		opts = append(opts, spreak.WithLanguage(l))
	}

	// Load path locales
	if len(localesDir) == 0 {
		return
	}

	if !goutils.DestExists(localesDir) {
		return
	}

	_customLanguages, errCustomLocales, err = walkLocaleDirPath(localesDir)
	if err != nil {
		return
	}

	if len(_customLanguages) > 0 {
		opts = append(opts, spreak.WithDomainPath(domainCustom, localesDir))
		for _, localLanguage := range _customLanguages {
			opts = append(opts, spreak.WithLanguage(localLanguage))
		}

		customLanguages = slices.Clone(_customLanguages)
	}

	bundle, err = spreak.NewBundle(opts...)

	return
}

func walkLocaleDirPath(localesPath string) (customLanguages []string, errCustomLocales, err error) {
	err = filepath.WalkDir(localesPath, func(pth string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		lang := strings.TrimSuffix(d.Name(), ".json")
		_, _, err = language.ParseAcceptLanguage(lang)
		if err != nil {
			return err
		}

		jsonBytes, err := os.ReadFile(pth)
		if err != nil {
			return err
		}

		m := make(map[string]interface{})

		err = json.Unmarshal(jsonBytes, &m)
		if err != nil {
			errCustomLocales = errors.Join(
				errCustomLocales,
				fmt.Errorf("%s: %v", d.Name(), err),
			)

			return nil
		}

		customLanguages = slice.AddMissingElems(customLanguages, lang)

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
	if slices.Contains(customLanguages, lang) {
		t = spreak.NewKeyLocalizerForDomain(bundle, domainCustom, lang)
	} else {
		t = spreak.NewKeyLocalizerForDomain(bundle, domainDefault, lang)
	}

	return t.Getf(s, args...)
}
