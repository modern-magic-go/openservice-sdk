package openservice

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOAuthURL(t *testing.T) {
	cfg := Config{
		BaseURL: "https://openservice.example.com",
		MID:     "10001",
		Secret:  "test-secret",
		Timeout: 15 * time.Second,
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tests := []struct {
		name     string
		req      OAuthRequest
		wantErr  bool
		checkURL func(t *testing.T, urlStr string)
	}{
		{
			name: "basic oauth url",
			req: OAuthRequest{
				Scope:       OAuthScopeUserInfo,
				RedirectURL: "https://example.com/auth/callback",
			},
			wantErr: false,
			checkURL: func(t *testing.T, urlStr string) {
				u, err := url.Parse(urlStr)
				if err != nil {
					t.Fatalf("parse url failed: %v", err)
				}
				if u.Path != "/api/login/oauth" {
					t.Errorf("unexpected path: %s", u.Path)
				}
				if u.Query().Get("mch_id") != "10001" {
					t.Errorf("unexpected mch_id: %s", u.Query().Get("mch_id"))
				}
				if u.Query().Get("scope") != "snsapi_userinfo" {
					t.Errorf("unexpected scope: %s", u.Query().Get("scope"))
				}
				if u.Query().Get("redirect_url") != "https%3A%2F%2Fexample.com%2Fauth%2Fcallback" {
					t.Errorf("unexpected redirect_url: %s", u.Query().Get("redirect_url"))
				}
			},
		},
		{
			name: "custom mch_id",
			req: OAuthRequest{
				MID:         "custom-mch",
				Scope:       OAuthScopeBase,
				RedirectURL: "https://example.com/callback",
			},
			wantErr: false,
			checkURL: func(t *testing.T, urlStr string) {
				u, err := url.Parse(urlStr)
				if err != nil {
					t.Fatalf("parse url failed: %v", err)
				}
				if u.Query().Get("mch_id") != "custom-mch" {
					t.Errorf("unexpected mch_id: %s", u.Query().Get("mch_id"))
				}
				if u.Query().Get("scope") != "snsapi_base" {
					t.Errorf("unexpected scope: %s", u.Query().Get("scope"))
				}
			},
		},
		{
			name:    "empty redirect_url",
			req:     OAuthRequest{Scope: OAuthScopeUserInfo},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr, err := client.OAuthURL(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("OAuthURL failed: %v", err)
			}
			if tt.checkURL != nil {
				tt.checkURL(t, urlStr)
			}
			if !strings.HasPrefix(urlStr, cfg.BaseURL) {
				t.Errorf("url should start with baseURL: %s", urlStr)
			}
		})
	}
}

func TestDecryptTicket(t *testing.T) {
	key := "test-key-16byte"
	iv := "test-iv-16-byte"

	tests := []struct {
		name    string
		ticket  string
		key     string
		iv      string
		wantErr bool
	}{
		{
			name:    "empty ticket",
			ticket:  "",
			key:     key,
			iv:      iv,
			wantErr: true,
		},
		{
			name:    "invalid base64",
			ticket:  "not-valid-base64!!!",
			key:     key,
			iv:      iv,
			wantErr: true,
		},
		{
			name:    "empty key",
			ticket:  "valid-ticket",
			key:     "",
			iv:      iv,
			wantErr: true,
		},
		{
			name:    "empty iv",
			ticket:  "valid-ticket",
			key:     key,
			iv:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfo, err := DecryptTicket(tt.ticket, tt.key, tt.iv)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("DecryptTicket failed: %v", err)
			}
			if userInfo == nil {
				t.Error("userInfo should not be nil")
			}
		})
	}
}

func TestDecryptTicket_InvalidKey(t *testing.T) {
	_, err := DecryptTicket("valid-ticket", "", "valid-iv")
	if err == nil {
		t.Error("expected error for empty aes_key")
	}

	_, err = DecryptTicket("valid-ticket", "valid-key", "")
	if err == nil {
		t.Error("expected error for empty aes_iv")
	}
}

func TestJSSDKSignatureRequestPayload(t *testing.T) {
	req := JSSDKSignatureRequest{
		URL:       "https://example.com/page",
		MID:       "custom-mid",
		NonceStr:  "test-nonce",
		Timestamp: 1712736000,
	}
	payload := req.payload("default-mid")

	if payload["url"] != "https://example.com/page" {
		t.Errorf("unexpected url: %v", payload["url"])
	}
	if payload["mid"] != "custom-mid" {
		t.Errorf("unexpected mid: %v", payload["mid"])
	}
	if payload["nonce_str"] != "test-nonce" {
		t.Errorf("unexpected nonce_str: %v", payload["nonce_str"])
	}
	if payload["timestamp"] != int64(1712736000) {
		t.Errorf("unexpected timestamp: %v", payload["timestamp"])
	}
}

func TestOAuthScopeConstants(t *testing.T) {
	if OAuthScopeBase != "snsapi_base" {
		t.Errorf("OAuthScopeBase should be snsapi_base, got %s", OAuthScopeBase)
	}
	if OAuthScopeUserInfo != "snsapi_userinfo" {
		t.Errorf("OAuthScopeUserInfo should be snsapi_userinfo, got %s", OAuthScopeUserInfo)
	}
}

func TestDecryptTicket_FullCycle(t *testing.T) {
	key := "test-key-16-byte"
	iv := "test-iv-16-byte!"

	userInfo := OAuthUserInfo{
		AppID:    "wx1234567890abcdef",
		OpenID:   "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
		UnionID:  "o6_bmasdasdsad6_2sgVt7hMZOPfL",
		Nick:     "用户昵称",
		Avatar:   "https://example.com/avatar.jpg",
		Sex:      1,
		Country:  "中国",
		Province: "广东",
		City:     "深圳",
	}

	plaintext, err := json.Marshal(userInfo)
	if err != nil {
		t.Fatalf("marshal userInfo failed: %v", err)
	}

	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("create cipher failed: %v", err)
	}

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	ticket := base64.StdEncoding.EncodeToString(ciphertext)

	decrypted, err := DecryptTicket(ticket, key, iv)
	if err != nil {
		t.Fatalf("DecryptTicket failed: %v", err)
	}

	if decrypted.AppID != userInfo.AppID {
		t.Errorf("AppID mismatch: got %s, want %s", decrypted.AppID, userInfo.AppID)
	}
	if decrypted.OpenID != userInfo.OpenID {
		t.Errorf("OpenID mismatch: got %s, want %s", decrypted.OpenID, userInfo.OpenID)
	}
	if decrypted.UnionID != userInfo.UnionID {
		t.Errorf("UnionID mismatch: got %s, want %s", decrypted.UnionID, userInfo.UnionID)
	}
	if decrypted.Nick != userInfo.Nick {
		t.Errorf("Nick mismatch: got %s, want %s", decrypted.Nick, userInfo.Nick)
	}
}

func TestDecryptTicket_UrlSafeBase64(t *testing.T) {
	key := "test-key-16-byte"
	iv := "test-iv-16-byte!"

	userInfo := OAuthUserInfo{
		AppID:  "wx1234567890abcdef",
		OpenID: "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
	}

	plaintext, err := json.Marshal(userInfo)
	if err != nil {
		t.Fatalf("marshal userInfo failed: %v", err)
	}

	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("create cipher failed: %v", err)
	}

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	ticket := base64.URLEncoding.EncodeToString(ciphertext)

	decrypted, err := DecryptTicket(ticket, key, iv)
	if err != nil {
		t.Fatalf("DecryptTicket failed: %v", err)
	}

	if decrypted.OpenID != userInfo.OpenID {
		t.Errorf("OpenID mismatch: got %s, want %s", decrypted.OpenID, userInfo.OpenID)
	}
}

func TestDecryptTicket_MissingPadding(t *testing.T) {
	key := "test-key-16-byte"
	iv := "test-iv-16-byte!"

	userInfo := OAuthUserInfo{
		AppID:  "wx1234567890abcdef",
		OpenID: "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
	}

	plaintext, err := json.Marshal(userInfo)
	if err != nil {
		t.Fatalf("marshal userInfo failed: %v", err)
	}

	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("create cipher failed: %v", err)
	}

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	ticket := base64.StdEncoding.EncodeToString(ciphertext)
	for len(ticket)%4 != 0 {
		ticket += "="
	}
	ticket = strings.TrimRight(ticket, "=")

	decrypted, err := DecryptTicket(ticket, key, iv)
	if err != nil {
		t.Fatalf("DecryptTicket failed: %v", err)
	}

	if decrypted.OpenID != userInfo.OpenID {
		t.Errorf("OpenID mismatch: got %s, want %s", decrypted.OpenID, userInfo.OpenID)
	}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}
