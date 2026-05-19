package openservice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modern-magic-go/openservice-sdk/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefundRequestPayload_UsesClientMIDAndOmitsEmptyOptionalFields(t *testing.T) {
	req := RefundRequest{
		OutTradeNo:   "T202604100001",
		OutRefundNo:  "R202604100001",
		TotalAmount:  100,
		RefundAmount: 20,
		NotifyURL:    "https://merchant.example.com/payment/refund/notify",
	}

	payload := req.payload("1900001")

	require.Equal(t, map[string]any{
		"mid":          "1900001",
		"outTradeNo":   "T202604100001",
		"outRefundNo":  "R202604100001",
		"totalAmount":  int64(100),
		"refundAmount": int64(20),
		"notifyUrl":    "https://merchant.example.com/payment/refund/notify",
	}, payload)
	assert.NotContains(t, payload, "currency")
	assert.NotContains(t, payload, "openid")
	assert.NotContains(t, payload, "subOpenid")
}

func TestRefundRequestPayload_PrefersRequestMIDAndIncludesOptionalFields(t *testing.T) {
	req := RefundRequest{
		BaseRequest:  BaseRequest{MID: "1900999"},
		OutTradeNo:   "T202604100001",
		OutRefundNo:  "R202604100001",
		TotalAmount:  100,
		RefundAmount: 20,
		Currency:     "CNY",
		NotifyURL:    "https://merchant.example.com/payment/refund/notify",
		OpenID:       "openid-1",
		SubOpenID:    "sub-openid-1",
	}

	payload := req.payload("1900001")

	assert.Equal(t, "1900999", payload["mid"])
	assert.Equal(t, "CNY", payload["currency"])
	assert.Equal(t, "openid-1", payload["openid"])
	assert.Equal(t, "sub-openid-1", payload["subOpenid"])
}

func TestRefundPayloadSigning_IgnoresEmptyOptionalFields(t *testing.T) {
	signer := NewSigner("test_merchant_secret_123456")
	payload := RefundRequest{
		OutTradeNo:   "T202604100001",
		OutRefundNo:  "R202604100001",
		TotalAmount:  100,
		RefundAmount: 20,
		NotifyURL:    "https://merchant.example.com/payment/refund/notify",
	}.payload("1900001")
	payload["nonce_str"] = "abc1234567890xyz"
	payload["timestamp"] = "1712736000"

	got := signer.BuildSignString(payload)

	assert.Equal(t,
		"mid=1900001&nonce_str=abc1234567890xyz&notifyUrl=https://merchant.example.com/payment/refund/notify&outRefundNo=R202604100001&outTradeNo=T202604100001&refundAmount=20&timestamp=1712736000&totalAmount=100&key=test_merchant_secret_123456",
		got,
	)
	assert.NotContains(t, got, "currency=")
	assert.NotContains(t, got, "openid=")
	assert.NotContains(t, got, "subOpenid=")
}

func TestDecodeResult_RefundEnvelopeUsesEnumStatus(t *testing.T) {
	var data RefundData
	body := []byte(`{"code":0,"message":"success","data":{"refund":{"mid":"1900001","appid":"wx1234567890","openid":"oUpF8uMuAJO_M2pxb1Q9zNjWeS6o","outTradeNo":"T202604100001","refundNo":"RF202604100001","outRefundNo":"R202604100001","channelRefundNo":"5000001234202404100000000000","currency":"CNY","totalAmount":100,"refundAmount":20,"transTime":"2026-04-10 10:05:00","transStatus":"Pending","transMessage":"退款处理中"}}}`)

	require.NoError(t, decodeResult(body, &data))
	assert.Equal(t, "1900001", data.Refund.MID)
	assert.Equal(t, "RF202604100001", data.Refund.RefundNo)
	assert.Equal(t, enums.TransStatusPending, data.Refund.TransStatus)
	assert.True(t, data.Refund.TransStatus.IsValid())
}

func TestDecodeResult_RefundReturnsMissingDataWhenEnvelopeDataIsNull(t *testing.T) {
	var data RefundData
	err := decodeResult([]byte(`{"code":0,"message":"success","data":null}`), &data)
	require.ErrorIs(t, err, ErrMissingData)
}

func TestDecodeResult_RefundReturnsResponseCodeError(t *testing.T) {
	var data RefundData
	err := decodeResult([]byte(`{"code":1,"message":"signature verify failed","data":null}`), &data)
	require.ErrorIs(t, err, ErrResponseCodeNonZero)
	require.Contains(t, err.Error(), "signature verify failed")
}

func TestClientRefund_PostsSignedPayloadAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/payment/refund", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "1900001", payload["mid"])
		assert.Equal(t, "T202604100001", payload["outTradeNo"])
		assert.Equal(t, "R202604100001", payload["outRefundNo"])
		assert.Equal(t, "https://merchant.example.com/payment/refund/notify", payload["notifyUrl"])
		assert.NotEmpty(t, payload["nonce_str"])
		assert.NotEmpty(t, payload["timestamp"])
		assert.NotEmpty(t, payload["sign"])
		assert.True(t, VerifySign(payload, "test_merchant_secret_123456"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"refund":{"mid":"1900001","outTradeNo":"T202604100001","refundNo":"RF202604100001","outRefundNo":"R202604100001","transStatus":"Pending","transMessage":"退款处理中"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL: server.URL,
		MID:     "1900001",
		Secret:  "test_merchant_secret_123456",
		Timeout: time.Second,
	})
	require.NoError(t, err)

	resp, err := client.Refund(context.Background(), RefundRequest{
		OutTradeNo:   "T202604100001",
		OutRefundNo:  "R202604100001",
		TotalAmount:  100,
		RefundAmount: 20,
		NotifyURL:    "https://merchant.example.com/payment/refund/notify",
	})
	require.NoError(t, err)
	assert.Equal(t, enums.TransStatusPending, resp.Refund.TransStatus)
	assert.Equal(t, "RF202604100001", resp.Refund.RefundNo)
}
