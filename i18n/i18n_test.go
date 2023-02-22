package i18n

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

//go:embed locale/*
var embedLocales embed.FS

func TestTranslate(t *testing.T) {
	fsys, err := fs.Sub(embedLocales, "locale")
	assert.Nil(t, err)

	err = Init(fsys, language.English, language.Chinese)
	assert.Nil(t, err)

	assert.True(t, IsLanguageSupported(language.Chinese.String()))

	assert.Equal(t, Translate("en_US", "Change world."), "Change world.")
	assert.Equal(t, Translate("en_US", "title"), "Hello World!")
	assert.Equal(t, TranslateF("en_US", "Hello %s %s", "John", "Smith"), "Hello John Smith")

	assert.Equal(t, TranslateF("zh_CN", "title"), "你好 世界！")
}
