package enc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryption(t *testing.T) {
	plainBytes := []byte("plain text")
	pass := "this_is_passphrase"

	encryptedBytes, err := Encrypt(pass, plainBytes)
	assert.Nil(t, err)

	decryptedBytes, err := Decrypt(pass, encryptedBytes)
	assert.Nil(t, err)
	assert.Equal(t, decryptedBytes, plainBytes)
}
