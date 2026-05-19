package openservice

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/modern-magic-go/openservice-sdk/enums"
)

type NotificationKind string

const (
	NotificationKindPayment NotificationKind = "payment"
	NotificationKindRefund  NotificationKind = "refund"
)

// PaymentNotification 表示 OpenService 支付回调通知。
type PaymentNotification struct {
	MID            string            `json:"mid,omitempty"`
	AppID          string            `json:"appid,omitempty"`
	OpenID         string            `json:"openid,omitempty"`
	Subject        string            `json:"subject,omitempty"`
	Attach         string            `json:"attach,omitempty"`
	TradeNo        string            `json:"tradeNo,omitempty"`
	OutTradeNo     string            `json:"outTradeNo,omitempty"`
	ChannelTradeNo string            `json:"channelTradeNo,omitempty"`
	Amount         int64             `json:"amount,omitempty"`
	Currency       string            `json:"currency,omitempty"`
	TransTime      string            `json:"transTime,omitempty"`
	TransStatus    enums.TransStatus `json:"transStatus,omitempty"`
	TransMessage   string            `json:"transMessage,omitempty"`
	NonceStr       string            `json:"nonceStr,omitempty"`
	Timestamp      int64             `json:"timestamp,omitempty"`
	Sign           string            `json:"sign,omitempty"`
	Raw            map[string]string `json:"raw,omitempty"`
}

// RefundNotification 表示 OpenService 退款回调通知。
type RefundNotification struct {
	MID             string            `json:"mid,omitempty"`
	AppID           string            `json:"appid,omitempty"`
	OpenID          string            `json:"openid,omitempty"`
	OutTradeNo      string            `json:"outTradeNo,omitempty"`
	RefundNo        string            `json:"refundNo,omitempty"`
	OutRefundNo     string            `json:"outRefundNo,omitempty"`
	ChannelRefundNo string            `json:"channelRefundNo,omitempty"`
	Currency        string            `json:"currency,omitempty"`
	TotalAmount     int64             `json:"totalAmount,omitempty"`
	RefundAmount    int64             `json:"refundAmount,omitempty"`
	TransTime       string            `json:"transTime,omitempty"`
	TransStatus     enums.TransStatus `json:"transStatus,omitempty"`
	TransMessage    string            `json:"transMessage,omitempty"`
	NonceStr        string            `json:"nonceStr,omitempty"`
	Timestamp       int64             `json:"timestamp,omitempty"`
	Sign            string            `json:"sign,omitempty"`
	Raw             map[string]string `json:"raw,omitempty"`
}

// NotificationParseResult 表示回调通知解析结果。
type NotificationParseResult struct {
	Kind    NotificationKind     `json:"kind"`
	Payment *PaymentNotification `json:"payment,omitempty"`
	Refund  *RefundNotification  `json:"refund,omitempty"`
	Raw     map[string]string    `json:"raw,omitempty"`
}

func verifyPaymentNotification(input any, secret string) error {
	raw, err := notificationParams(input)
	if err != nil {
		return err
	}
	if strings.TrimSpace(secret) == "" {
		return ErrMissingSecret
	}
	if !NewSigner(secret).VerifySign(notificationSignParams(raw)) {
		return fmt.Errorf("%w: notification signature verify failed", ErrInvalidRequest)
	}
	return nil
}

func parsePaymentNotificationInput(input any) (*NotificationParseResult, error) {
	raw, err := notificationParams(input)
	if err != nil {
		return nil, err
	}
	if isRefundNotification(raw) {
		refund, err := parseRefundNotification(raw)
		if err != nil {
			return nil, err
		}
		return &NotificationParseResult{Kind: NotificationKindRefund, Refund: refund, Raw: raw}, nil
	}
	payment, err := parsePaymentNotification(raw)
	if err != nil {
		return nil, err
	}
	return &NotificationParseResult{Kind: NotificationKindPayment, Payment: payment, Raw: raw}, nil
}

func notificationParams(input any) (map[string]string, error) {
	switch params := input.(type) {
	case url.Values:
		return valuesToStringMap(params), nil
	case map[string]string:
		return cloneStringMap(params), nil
	case map[string]any:
		return anyMapToStringMap(params), nil
	default:
		return nil, fmt.Errorf("%w: unsupported notification params", ErrInvalidRequest)
	}
}

func valuesToStringMap(values url.Values) map[string]string {
	params := make(map[string]string, len(values))
	for key := range values {
		params[key] = values.Get(key)
	}
	return params
}

func cloneStringMap(input map[string]string) map[string]string {
	params := make(map[string]string, len(input))
	for key, value := range input {
		params[key] = value
	}
	return params
}

func anyMapToStringMap(input map[string]any) map[string]string {
	params := make(map[string]string, len(input))
	for key, value := range input {
		params[key] = normalizeSignValue(value)
	}
	return params
}

func notificationSignParams(raw map[string]string) map[string]any {
	params := make(map[string]any, len(raw))
	for key, value := range raw {
		params[key] = value
	}
	return params
}

func isRefundNotification(raw map[string]string) bool {
	return raw["refundNo"] != "" || raw["outRefundNo"] != "" || raw["channelRefundNo"] != "" || raw["refundAmount"] != ""
}

func parsePaymentNotification(raw map[string]string) (*PaymentNotification, error) {
	if raw["outTradeNo"] == "" {
		return nil, fmt.Errorf("%w: outTradeNo is required", ErrInvalidRequest)
	}
	amount, err := parseNotificationInt64(raw, "amount")
	if err != nil {
		return nil, err
	}
	timestamp, err := parseNotificationInt64(raw, "timestamp")
	if err != nil {
		return nil, err
	}
	return &PaymentNotification{
		MID:            raw["mid"],
		AppID:          raw["appid"],
		OpenID:         raw["openid"],
		Subject:        raw["subject"],
		Attach:         raw["attach"],
		TradeNo:        raw["tradeNo"],
		OutTradeNo:     raw["outTradeNo"],
		ChannelTradeNo: raw["channelTradeNo"],
		Amount:         amount,
		Currency:       raw["currency"],
		TransTime:      raw["transTime"],
		TransStatus:    enums.TransStatus(raw["transStatus"]),
		TransMessage:   raw["transMessage"],
		NonceStr:       raw["nonceStr"],
		Timestamp:      timestamp,
		Sign:           raw["sign"],
		Raw:            raw,
	}, nil
}

func parseRefundNotification(raw map[string]string) (*RefundNotification, error) {
	if raw["outRefundNo"] == "" {
		return nil, fmt.Errorf("%w: outRefundNo is required", ErrInvalidRequest)
	}
	totalAmount, err := parseNotificationInt64(raw, "totalAmount")
	if err != nil {
		return nil, err
	}
	refundAmount, err := parseNotificationInt64(raw, "refundAmount")
	if err != nil {
		return nil, err
	}
	timestamp, err := parseNotificationInt64(raw, "timestamp")
	if err != nil {
		return nil, err
	}
	return &RefundNotification{
		MID:             raw["mid"],
		AppID:           raw["appid"],
		OpenID:          raw["openid"],
		OutTradeNo:      raw["outTradeNo"],
		RefundNo:        raw["refundNo"],
		OutRefundNo:     raw["outRefundNo"],
		ChannelRefundNo: raw["channelRefundNo"],
		Currency:        raw["currency"],
		TotalAmount:     totalAmount,
		RefundAmount:    refundAmount,
		TransTime:       raw["transTime"],
		TransStatus:     enums.TransStatus(raw["transStatus"]),
		TransMessage:    raw["transMessage"],
		NonceStr:        raw["nonceStr"],
		Timestamp:       timestamp,
		Sign:            raw["sign"],
		Raw:             raw,
	}, nil
}

func parseNotificationInt64(raw map[string]string, key string) (int64, error) {
	value := strings.TrimSpace(raw[key])
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be int64", ErrInvalidRequest, key)
	}
	return parsed, nil
}
