package emailutils

import (
	"fmt"
	"io"
	"mime"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// DecodeHeader decodes a MIME encoded-word header according to RFC 2047,
// with extended support for ISO-8859-9 (Latin-5, Turkish) character set.
func DecodeHeader(v string) (string, error) {
	wd := &mime.WordDecoder{}

	// Handle iso-8859-9
	wd.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if strings.EqualFold(charset, "iso-8859-9") {
			return charmap.ISO8859_9.NewDecoder().Reader(input), nil
		}

		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}

	return wd.DecodeHeader(v)
}
