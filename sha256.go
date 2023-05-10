package goutils

import (
	"bytes"
	"crypto/sha256"
	"io"
)

func GenSha256(content []byte) (shasum []byte, err error) {
	h := sha256.New()
	if _, err = io.Copy(h, bytes.NewReader(content)); err != nil {
		return
	}

	return h.Sum(nil), nil
}

func GenSha256FromReader(r io.Reader) (shasum []byte, err error) {
	h := sha256.New()
	if _, err = io.Copy(h, r); err != nil {
		return
	}

	return h.Sum(nil), nil
}

func IsSameSha256(a, b []byte) bool {
	return bytes.Equal(a, b)
}
