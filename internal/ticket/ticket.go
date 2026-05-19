package ticket

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
)

func Decrypt(ticket, aesKey, aesIV string) ([]byte, error) {
	key := []byte(aesKey)
	iv := []byte(aesIV)

	normalizedTicket := strings.ReplaceAll(ticket, "-", "+")
	normalizedTicket = strings.ReplaceAll(normalizedTicket, "_", "/")

	padding := 4 - len(normalizedTicket)%4
	if padding < 4 {
		normalizedTicket += strings.Repeat("=", padding)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(normalizedTicket)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return nil, ErrInvalidCiphertext
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext), nil
}

var ErrInvalidCiphertext = errInvalidCiphertext{}

type errInvalidCiphertext struct{}

func (errInvalidCiphertext) Error() string { return "invalid ciphertext" }

func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > aes.BlockSize {
		return data
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return data
		}
	}
	return data[:len(data)-padding]
}
