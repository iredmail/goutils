package i18n

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

//go:embed test_locales/*.json
var fsEmbedLocales embed.FS

func TestTranslate(t *testing.T) {
	fsLocales, err := fs.Sub(fsEmbedLocales, "test_locales")
	assert.Nil(t, err)

	err = Init(fsLocales, language.English, language.English, language.SimplifiedChinese)
	assert.Nil(t, err)

	assert.True(t, IsLanguageSupported(language.SimplifiedChinese.String()))
	assert.True(t, IsLanguageSupported(language.English.String()))
	assert.False(t, IsLanguageSupported(language.Slovenian.String()))

	assert.Equal(t, Translate("en_US", "Change world."), "Change world.")
	assert.Equal(t, Translate("en_US", "hello"), "Hello")
	assert.Equal(t, TranslateF("en_US", "Hello %s %s", "John", "Smith"), "Hello John Smith")

	assert.Equal(t, TranslateF("zh_CN", "hello"), "你好")
	assert.Equal(t, TranslateF("zh", "hello"), "你好")
}
