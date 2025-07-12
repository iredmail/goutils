package emailutils

import (
	"fmt"
	"io"
	"mime"

	"golang.org/x/text/encoding/htmlindex"
)

// DecodeHeader decodes a MIME encoded-word header according to RFC 2047,
// with all possible character set support.
//
// Note: mail.ParseAddress() supports just few character sets.
func DecodeHeader(v string) (string, error) {
	wd := &mime.WordDecoder{}

	// Handle iso-8859-9
	wd.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		enc, err := htmlindex.Get(charset)
		if err != nil {
			return nil, fmt.Errorf("unsupported charset: %s", charset)
		}

		return enc.NewDecoder().Reader(input), nil
	}

	return wd.DecodeHeader(v)
}
