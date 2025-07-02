package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Original ISO-8859-9 tests
		{
			name:     "ISO-8859-9 encoded Turkish text",
			input:    "=?iso-8859-9?Q?=E7i=F0ek=20test?=", // Corrected encoding
			expected: "çiğek test",                        // Turkish characters
			wantErr:  false,
		},
		{
			name:     "ISO-8859-9 encoded Turkish text",
			input:    "=?iso-8859-9?Q?Merhaba!=20D=FCnya?=", // Corrected encoding
			expected: "Merhaba! Dünya",                      // Turkish characters
			wantErr:  false,
		},
		{
			name:     "UTF-8 Chinese with spaces (Simplified)",
			input:    "=?utf-8?Q?=E4=B8=AD=E6=96=87=20=E6=B5=8B=E8=AF=95?=",
			expected: "中文 测试",
			wantErr:  false,
		},
		{
			name:     "UTF-8 Taiwanese with mixed text",
			input:    "=?utf-8?Q?=E5=8F=B0=E7=81=A3=20(Taiwan)=20=E6=B8=AC=E8=A9=A6?=",
			expected: "台灣 (Taiwan) 測試",
			wantErr:  false,
		},
		// Other existing tests
		{
			name:     "Unencoded text",
			input:    "Plain text header",
			expected: "Plain text header",
			wantErr:  false,
		},
		{
			name:     "Unsupported charset",
			input:    "=?iso-8859-2?Q?test?=",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeHeader(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("DecodeHeader() = %v, want %v", got, tt.expected)
			}
		})
	}

	// decoded: `中文 中文 中文 <user@domain.com>`
	v, err := DecodeHeader("=?iso-8859-9?Q?Yavuz_Ma=FElak?= <user@domain.tr>")
	assert.Nil(t, err)
	assert.Equal(t, "Yavuz Maşlak <user@domain.tr>", v)
}
