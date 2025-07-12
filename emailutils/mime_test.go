package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {
	// Turkish ISO-8859-9 cases
	v, err := DecodeHeader("=?iso-8859-9?Q?=E7i=F0ek=20test?=")
	assert.Nil(t, err)
	assert.Equal(t, "çiğek test", v)

	v, err = DecodeHeader("=?iso-8859-9?Q?Merhaba!=20D=FCnya?=")
	assert.Nil(t, err)
	assert.Equal(t, "Merhaba! Dünya", v)

	// UTF-8 Chinese cases
	v, err = DecodeHeader("=?utf-8?Q?=E4=B8=AD=E6=96=87=20=E6=B5=8B=E8=AF=95?=")
	assert.Nil(t, err)
	assert.Equal(t, "中文 测试", v)

	v, err = DecodeHeader("=?utf-8?Q?=E5=8F=B0=E7=81=A3=20(Taiwan)=20=E6=B8=AC=E8=A9=A6?=")
	assert.Nil(t, err)
	assert.Equal(t, "台灣 (Taiwan) 測試", v)

	// Plain text case
	v, err = DecodeHeader("Plain text header")
	assert.Nil(t, err)
	assert.Equal(t, "Plain text header", v)

	// Additional Turkish case
	v, err = DecodeHeader("=?iso-8859-9?Q?Yavuz_Ma=FElak?= <user@domain.tr>")
	assert.Nil(t, err)
	assert.Equal(t, "Yavuz Maşlak <user@domain.tr>", v)
}
