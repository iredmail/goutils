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

	assert.True(t, len(Languages()) == 2)

	assert.Equal(t, Translate(language.English, "Change world."), "Change world.")
	assert.Equal(t, Translate(language.English, "title"), "Hello World!")
	assert.Equal(t, TranslateF(language.English, "Hello %s %s", "John", "Smith"), "Hello John Smith")

	assert.Equal(t, TranslateF(language.Chinese, "title"), "你好 世界！")
}
