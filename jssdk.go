package openservice

import (
	"encoding/json"
	"fmt"

	"github.com/modern-magic-go/openservice-sdk/internal/ticket"
)

const (
	oauthPath     = "/api/login/oauth"
	jssdkSignPath = "/api/jssdk/signature"
)

func decryptTicket(encryptedTicket, aesKey, aesIV string) (*OAuthUserInfo, error) {
	if encryptedTicket == "" {
		return nil, fmt.Errorf("%w: ticket is required", ErrInvalidRequest)
	}
	if aesKey == "" {
		return nil, fmt.Errorf("%w: aes_key is required", ErrInvalidRequest)
	}
	if aesIV == "" {
		return nil, fmt.Errorf("%w: aes_iv is required", ErrInvalidRequest)
	}

	plaintext, err := ticket.Decrypt(encryptedTicket, aesKey, aesIV)
	if err != nil {
		return nil, fmt.Errorf("%w: decode ticket: %v", ErrInvalidResponse, err)
	}

	var userInfo OAuthUserInfo
	if err := json.Unmarshal(plaintext, &userInfo); err != nil {
		return nil, fmt.Errorf("%w: parse user info: %v", ErrInvalidResponse, err)
	}

	return &userInfo, nil
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
