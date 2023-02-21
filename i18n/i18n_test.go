package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

//go:embed locale/*
var embedLocales embed.FS

func TestTranslate(t *testing.T) {
	fsys, err := fs.Sub(embedLocales, "locale")
	assert.Nil(t, err)

	err = Init(fsys, language.Chinese)
	assert.Nil(t, err)

	fmt.Println(Languages())
	assert.True(t, len(Languages()) == 2)

	tt := spreak.NewKeyLocalizer(bundle, language.English)
	fmt.Println(tt.Get("title"))
	//fmt.Println(Translate(language.English, "title"))
	//assert.Equal(t, TranslateF(language.English, "title"), "Hello World!")
	//assert.Equal(t, TranslateF(language.English, "Hello %s %s", "John", "Smith"), "Hello John Smith")
	//
	//assert.Equal(t, TranslateF(language.Chinese, "title"), "你好 世界！")
}
