package signing

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

func BuildSignString(params map[string]any, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "sign" || key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		value := NormalizeValue(params[key])
		if value == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
	}
	builder.WriteString("&key=")
	builder.WriteString(secret)

	return builder.String()
}

func Sign(params map[string]any, secret string) string {
	if strings.TrimSpace(secret) == "" {
		return ""
	}
	return UpperMD5(BuildSignString(params, secret))
}

func Verify(params map[string]any, secret string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	sign := ""
	if value, ok := params["sign"].(string); ok {
		sign = value
	}
	return Sign(params, secret) == sign
}

func NormalizeValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case []any:
		if len(typed) == 0 {
			return ""
		}
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	case map[string]any:
		if len(typed) == 0 {
			return ""
		}
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	default:
		if typed == nil {
			return ""
		}
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	}
}

func UpperMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// HMACSHA256 对 message 做 HMAC-SHA256，返回大写 hex 字符串。
func HMACSHA256(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

// SHA256Hex 对 data 做 SHA256，返回小写 hex 字符串。
func SHA256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
