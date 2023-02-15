package i18n

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed locales
var embedLocales embed.FS

func TestTranslate(t *testing.T) {
	locales, err := New(embedLocales)
	assert.Nil(t, err)

	assert.Equal(t, locales.Translate(DefaultLanguage, "title"), "Hello World!")
	assert.Equal(t, locales.Translate(DefaultLanguage, "Hello %s %s", "John", "Smith"), "Hello John Smith")

	assert.Equal(t, locales.Translate("zh_CN", "title"), "你好 世界！")
}
