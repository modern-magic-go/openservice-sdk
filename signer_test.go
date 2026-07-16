package openservice

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/modern-magic-go/openservice-sdk/internal/signing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSignString(t *testing.T) {
	signer := NewSigner("secret-key")
	got := signer.BuildSignString(map[string]any{
		"b":      2,
		"a":      "1",
		"sign":   "skip-me",
		"empty":  "",
		"nil":    nil,
		"flag":   false,
		"number": int64(7),
	})

	assert.Equal(t, "a=1&b=2&flag=false&number=7&key=secret-key", got)
}

func TestSignUppercaseMD5(t *testing.T) {
	signer := NewSigner("secret-key")
	params := map[string]any{"a": "1", "b": "2"}

	got := signer.Sign(params)
	manual := md5.Sum([]byte("a=1&b=2&key=secret-key"))
	require.Equal(t, strings.ToUpper(hex.EncodeToString(manual[:])), got)
}

func TestSignValues(t *testing.T) {
	signer := NewSigner("secret-key")
	values := url.Values{}
	values.Set("b", "2")
	values.Set("a", "1")

	got := signer.SignValues(values)
	require.Equal(t, signer.Sign(map[string]any{"a": "1", "b": "2"}), got)
}

func TestSignIgnoresSignatureField(t *testing.T) {
	signer := NewSigner("secret-key")
	got := signer.BuildSignString(map[string]any{
		"signature": "skip",
		"sign":      "skip",
		"a":         "1",
	})

	assert.Equal(t, "a=1&key=secret-key", got)
}

func TestSignerVerifySignIgnoresSignatureField(t *testing.T) {
	signer := NewSigner("secret-key")
	params := map[string]any{
		"signature": "skip",
		"a":         "1",
	}
	params["sign"] = signer.Sign(params)

	assert.True(t, signer.VerifySign(params))
}

func TestSignerVerifySign_DoesNotMutateInput(t *testing.T) {
	signer := NewSigner("secret-key")
	params := map[string]any{
		"a":         "1",
		"signature": "keep-me",
	}
	params["sign"] = signer.Sign(params)

	signBefore := params["sign"]
	signatureBefore := params["signature"]

	assert.True(t, signer.VerifySign(params))
	assert.Equal(t, signBefore, params["sign"])
	assert.Equal(t, signatureBefore, params["signature"])
}

func TestQueryOrderData_UnmarshalFlatResponse(t *testing.T) {
	var data QueryOrderData
	body := []byte(`{"amount":117,"appid":"wx0ac8965ec03033dd","attach":"","channelTradeNo":"4200003106202604118816609115","currency":"CNY","mid":"208241210173031774","openid":"oa1uS66WIiSDxe6z4AImJhsBFwuw","outTradeNo":"10008","subject":"渠道H5餐饮测试专用(仅供测试)外卖订单","tradeNo":"400260411083900003","transMessage":"支付成功","transStatus":"Success","transTime":"2026-04-11 08:39:13"}`)

	require.NoError(t, json.Unmarshal(body, &data))
	assert.Equal(t, "Success", string(data.Order.TransStatus))
	assert.Equal(t, "10008", data.Order.OutTradeNo)
	assert.Equal(t, "400260411083900003", data.Order.TradeNo)
}

// ==================== v2 HMAC-SHA256 ====================

func TestSHA256Hex(t *testing.T) {
	got := signing.SHA256Hex([]byte("hello"))
	manual := sha256.Sum256([]byte("hello"))
	require.Equal(t, hex.EncodeToString(manual[:]), got)
}

func TestHMACSHA256(t *testing.T) {
	got := signing.HMACSHA256("secret", "message")
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("message"))
	expected := strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
	require.Equal(t, expected, got)
}

func TestSignHeader_GeneratesHeaders(t *testing.T) {
	signer := NewSigner("test-secret")
	body := []byte(`{"outTradeNo":"ORDER001","amount":100}`)

	headers := signer.SignHeader("POST", "/api/v2/payment/create", body, "M001")

	assert.NotEmpty(t, headers["X-Auth-Mid"])
	assert.NotEmpty(t, headers["X-Auth-Timestamp"])
	assert.NotEmpty(t, headers["X-Auth-Nonce"])
	assert.NotEmpty(t, headers["X-Auth-Signature"])
	assert.Equal(t, "M001", headers["X-Auth-Mid"])
	assert.Len(t, headers["X-Auth-Nonce"], 32) // 16 bytes → 32 hex chars
}

func TestVerifyHeader_CorrectSignature(t *testing.T) {
	signer := NewSigner("test-secret")
	method := "POST"
	path := "/api/v2/payment/create"
	body := []byte(`{"outTradeNo":"ORDER001","amount":100}`)
	mid := "M001"

	headers := signer.SignHeader(method, path, body, mid)
	ts := headers["X-Auth-Timestamp"]
	nonce := headers["X-Auth-Nonce"]
	signature := headers["X-Auth-Signature"]

	ok := signer.VerifyHeader(method, path, body, ts, nonce, signature, "test-secret")
	assert.True(t, ok, "signature should verify with correct secret")
}

func TestVerifyHeader_WrongSecret(t *testing.T) {
	signer := NewSigner("test-secret")
	method := "POST"
	path := "/api/v2/payment/create"
	body := []byte(`{"outTradeNo":"ORDER001","amount":100}`)
	mid := "M001"

	headers := signer.SignHeader(method, path, body, mid)
	ts := headers["X-Auth-Timestamp"]
	nonce := headers["X-Auth-Nonce"]
	signature := headers["X-Auth-Signature"]

	ok := signer.VerifyHeader(method, path, body, ts, nonce, signature, "wrong-secret")
	assert.False(t, ok, "signature should fail with wrong secret")
}

func TestVerifyHeader_TamperedBody(t *testing.T) {
	signer := NewSigner("test-secret")
	method := "POST"
	path := "/api/v2/payment/create"
	body := []byte(`{"outTradeNo":"ORDER001","amount":100}`)
	mid := "M001"

	headers := signer.SignHeader(method, path, body, mid)
	ts := headers["X-Auth-Timestamp"]
	nonce := headers["X-Auth-Nonce"]
	signature := headers["X-Auth-Signature"]

	// 篡改 body
	tamperedBody := []byte(`{"outTradeNo":"ORDER001","amount":999}`)

	ok := signer.VerifyHeader(method, path, tamperedBody, ts, nonce, signature, "test-secret")
	assert.False(t, ok, "signature should fail when body is tampered")
}

func TestVerifyHeader_DifferentPath(t *testing.T) {
	signer := NewSigner("test-secret")
	method := "POST"
	path := "/api/v2/payment/create"
	body := []byte(`{"outTradeNo":"ORDER001","amount":100}`)
	mid := "M001"

	headers := signer.SignHeader(method, path, body, mid)
	ts := headers["X-Auth-Timestamp"]
	nonce := headers["X-Auth-Nonce"]
	signature := headers["X-Auth-Signature"]

	// 用不同 path 验证
	ok := signer.VerifyHeader(method, "/api/v2/payment/query", body, ts, nonce, signature, "test-secret")
	assert.False(t, ok, "signature should fail with different path")
}
