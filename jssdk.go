package openservice

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	oauthPath     = "/api/login/oauth"
	jssdkSignPath = "/api/jssdk/signature"
)

func decryptTicket(ticket, aesKey, aesIV string) (*OAuthUserInfo, error) {
	if ticket == "" {
		return nil, fmt.Errorf("%w: ticket is required", ErrInvalidRequest)
	}
	if aesKey == "" {
		return nil, fmt.Errorf("%w: aes_key is required", ErrInvalidRequest)
	}
	if aesIV == "" {
		return nil, fmt.Errorf("%w: aes_iv is required", ErrInvalidRequest)
	}

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
		return nil, fmt.Errorf("%w: decode ticket: %v", ErrInvalidResponse, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: create cipher: %v", ErrInvalidResponse, err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrInvalidResponse)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("%w: ciphertext is not a multiple of block size", ErrInvalidResponse)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext = pkcs7Unpad(plaintext)

	var userInfo OAuthUserInfo
	if err := json.Unmarshal(plaintext, &userInfo); err != nil {
		return nil, fmt.Errorf("%w: parse user info: %v", ErrInvalidResponse, err)
	}

	return &userInfo, nil
}

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

func (r OAuthRequest) payload(mid string) map[string]any {
	return map[string]any{
		"mch_id":       resolveString(r.MID, mid),
		"scope":        string(r.Scope),
		"redirect_url": r.RedirectURL,
	}
}

func (r JSSDKSignatureRequest) payload(mid string) map[string]any {
	return map[string]any{
		"url":       r.URL,
		"mid":       resolveString(r.MID, mid),
		"nonce_str": r.NonceStr,
		"timestamp": r.Timestamp,
	}
}
