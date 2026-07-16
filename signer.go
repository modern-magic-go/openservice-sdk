package openservice

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/modern-magic-go/openservice-sdk/internal/signing"
)

type Signer struct {
	secret string
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: strings.TrimSpace(secret)}
}

func (s *Signer) BuildSignString(params map[string]any) string {
	return signing.BuildSignString(params, s.secret)
}

func (s *Signer) Sign(params map[string]any) string {
	if s == nil {
		return ""
	}
	return signing.Sign(params, s.secret)
}

func (s *Signer) SignValues(values url.Values) string {
	params := make(map[string]any, len(values))
	for key, raw := range values {
		if len(raw) == 0 {
			continue
		}
		if len(raw) == 1 {
			params[key] = raw[0]
			continue
		}
		params[key] = raw
	}
	return s.Sign(params)
}

func (s *Signer) VerifySign(params map[string]any) bool {
	if s == nil {
		return false
	}
	return signing.Verify(params, s.secret)
}

// SignHeader 生成 Header 签名（HMAC-SHA256）。
// 返回的 map 可直接设为 HTTP Header：X-Auth-Mid, X-Auth-Timestamp, X-Auth-Nonce, X-Auth-Signature。
// 签名原文格式：METHOD\nPATH\nTIMESTAMP\nNONCE\nSHA256(BODY)\n
func (s *Signer) SignHeader(method, path string, body []byte, mid string) map[string]string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(16)

	bodyHash := signing.SHA256Hex(body)

	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, path, ts, nonce, bodyHash)
	signature := signing.HMACSHA256(s.secret, signStr)

	return map[string]string{
		"X-Auth-Mid":       mid,
		"X-Auth-Timestamp": ts,
		"X-Auth-Nonce":     nonce,
		"X-Auth-Signature": signature,
	}
}

// VerifyHeader 验证 v2 Header 签名。
// timestamp / nonce / signature 从 HTTP Header 中提取，secret 由调用方查 DB 后传入。
func (s *Signer) VerifyHeader(method, path string, body []byte, timestamp, nonce, signature, secret string) bool {
	bodyHash := signing.SHA256Hex(body)
	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, path, timestamp, nonce, bodyHash)

	expected := signing.HMACSHA256(secret, signStr)
	return expected == signature
}

// generateNonce 生成指定字节数的随机 hex 字符串。
func generateNonce(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func normalizeSignValue(value any) string {
	return signing.NormalizeValue(value)
}
