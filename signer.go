package openservice

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type Signer struct {
	secret string
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: strings.TrimSpace(secret)}
}

func (s *Signer) BuildSignString(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := params[k]
		val := normalizeSignValue(v)
		if val == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(val)
	}
	sb.WriteString("&key=")
	sb.WriteString(s.secret)

	return sb.String()
}

func (s *Signer) Sign(params map[string]any) string {
	if s == nil || s.secret == "" {
		return ""
	}
	return upperMD5(s.BuildSignString(params))
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

func VerifySign(params map[string]any, secret string) bool {
	sign := ""
	if v, ok := params["sign"].(string); ok {
		sign = v
	}
	verifyParams := make(map[string]any, len(params))
	for key, value := range params {
		if key == "sign" || key == "signature" {
			continue
		}
		verifyParams[key] = value
	}

	keys := make([]string, 0, len(verifyParams))
	for k := range verifyParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := verifyParams[k]
		val := normalizeSignValue(v)
		if val == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(val)
	}
	sb.WriteString("&key=")
	sb.WriteString(secret)

	hash := md5.Sum([]byte(sb.String()))
	real := strings.ToUpper(hex.EncodeToString(hash[:]))
	return real == sign
}

func normalizeSignValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case []any:
		if len(v) == 0 {
			return ""
		}
		b, _ := json.Marshal(v)
		return string(b)
	case map[string]any:
		if len(v) == 0 {
			return ""
		}
		b, _ := json.Marshal(v)
		return string(b)
	default:
		if v == nil {
			return ""
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func upperMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
