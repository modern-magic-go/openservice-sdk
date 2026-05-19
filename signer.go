package openservice

import (
	"net/url"
	"strings"

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

func normalizeSignValue(value any) string {
	return signing.NormalizeValue(value)
}
