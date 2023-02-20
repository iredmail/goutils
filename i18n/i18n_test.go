package i18n

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed locales
var embedLocales embed.FS

func TestTranslate(t *testing.T) {
	err := Init(embedLocales)
	assert.Nil(t, err)

	assert.True(t, len(Languages()) == 2)

	assert.Equal(t, Translate(DefaultLanguage, "title"), "Hello World!")
	assert.Equal(t, Translate(DefaultLanguage, "Hello %s %s", "John", "Smith"), "Hello John Smith")

	assert.Equal(t, Translate("zh_CN", "title"), "你好 世界！")
}
