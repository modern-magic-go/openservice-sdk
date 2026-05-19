package openservice

import (
	"net/url"
	"testing"
	"time"

	"github.com/modern-magic-go/openservice-sdk/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentNotification_VerifyAndParsePayment(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "https://openservice.example.com",
		MID:     "1900001",
		Secret:  "merchant-secret",
		Timeout: time.Second,
	})
	require.NoError(t, err)

	params := url.Values{}
	params.Set("mid", "1900001")
	params.Set("outTradeNo", "T202604100001")
	params.Set("tradeNo", "P202604100001")
	params.Set("amount", "100")
	params.Set("transStatus", "Success")
	params.Set("nonceStr", "abc1234567890xyz")
	params.Set("timestamp", "1712736000")
	params.Set("sign", NewSigner("merchant-secret").SignValues(params))

	require.NoError(t, client.Payment().VerifyNotification(params))
	parsed, err := client.Payment().ParseNotification(params)
	require.NoError(t, err)

	assert.Equal(t, NotificationKindPayment, parsed.Kind)
	require.NotNil(t, parsed.Payment)
	assert.Equal(t, "T202604100001", parsed.Payment.OutTradeNo)
	assert.Equal(t, int64(100), parsed.Payment.Amount)
	assert.Equal(t, enums.TransStatusSuccess, parsed.Payment.TransStatus)
}

func TestPaymentNotification_VerifyAndParseRefund(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "https://openservice.example.com",
		MID:     "1900001",
		Secret:  "merchant-secret",
		Timeout: time.Second,
	})
	require.NoError(t, err)

	params := url.Values{}
	params.Set("mid", "1900001")
	params.Set("outTradeNo", "T202604100001")
	params.Set("refundNo", "RF202604100001")
	params.Set("outRefundNo", "R202604100001")
	params.Set("refundAmount", "20")
	params.Set("transStatus", "Success")
	params.Set("nonceStr", "abc1234567890xyz")
	params.Set("timestamp", "1712736000")
	params.Set("sign", NewSigner("merchant-secret").SignValues(params))

	require.NoError(t, client.Payment().VerifyNotification(params))
	parsed, err := client.Payment().ParseNotification(params)
	require.NoError(t, err)

	assert.Equal(t, NotificationKindRefund, parsed.Kind)
	require.NotNil(t, parsed.Refund)
	assert.Equal(t, "R202604100001", parsed.Refund.OutRefundNo)
	assert.Equal(t, int64(20), parsed.Refund.RefundAmount)
	assert.Equal(t, enums.TransStatusSuccess, parsed.Refund.TransStatus)
}

func TestPaymentNotification_VerifyRejectsInvalidSign(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "https://openservice.example.com",
		MID:     "1900001",
		Secret:  "merchant-secret",
		Timeout: time.Second,
	})
	require.NoError(t, err)

	params := map[string]string{
		"mid":        "1900001",
		"outTradeNo": "T202604100001",
		"amount":     "100",
		"sign":       "BAD_SIGN",
	}

	err = client.Payment().VerifyNotification(params)
	require.ErrorIs(t, err, ErrInvalidRequest)
}
