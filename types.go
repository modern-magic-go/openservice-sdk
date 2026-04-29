package openservice

import (
	"strings"

	"github.com/modern-magic-go/openservice/enums"
)

// Result 定义 OpenService 统一响应包装。
type Result[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// PaymentParams 表示前端拉起支付参数集合。
type PaymentParams struct {
	AppID     string `json:"appId,omitempty"`
	TimeStamp string `json:"timeStamp,omitempty"`
	NonceStr  string `json:"nonceStr,omitempty"`
	Package   string `json:"package,omitempty"`
	SignType  string `json:"signType,omitempty"`
	PaySign   string `json:"paySign,omitempty"`
}

// Order 表示 OpenService 返回的订单投影。
type Order struct {
	MID            string             `json:"mid,omitempty"`
	AppID          string             `json:"appid,omitempty"`
	OpenID         string             `json:"openid,omitempty"`
	Subject        string             `json:"subject,omitempty"`
	Attach         string             `json:"attach,omitempty"`
	TradeNo        string             `json:"tradeNo,omitempty"`
	OutTradeNo     string             `json:"outTradeNo,omitempty"`
	ChannelTradeNo string             `json:"channelTradeNo,omitempty"`
	Amount         int64              `json:"amount,omitempty"`
	TotalAmount    int64              `json:"totalAmount,omitempty"`
	RefundAmount   int64              `json:"refundAmount,omitempty"`
	Currency       string             `json:"currency,omitempty"`
	TransTime      string             `json:"transTime,omitempty"`
	TransStatus    enums.TransStatus  `json:"transStatus,omitempty"`
	TransMessage   string             `json:"transMessage,omitempty"`
	NotifyStatus   enums.NotifyStatus `json:"notifyStatus,omitempty"`
	NotifyURL      string             `json:"notifyUrl,omitempty"`
	PaidAt         string             `json:"paidAt,omitempty"`
	CreatedAt      string             `json:"createdAt,omitempty"`
	UpdatedAt      string             `json:"updatedAt,omitempty"`
	Raw            map[string]any     `json:"raw,omitempty"`
}

// Refund 表示 OpenService 返回的退款投影。
type Refund struct {
	MID             string             `json:"mid,omitempty"`
	AppID           string             `json:"appid,omitempty"`
	OpenID          string             `json:"openid,omitempty"`
	TradeNo         string             `json:"tradeNo,omitempty"`
	OutTradeNo      string             `json:"outTradeNo,omitempty"`
	RefundNo        string             `json:"refundNo,omitempty"`
	OutRefundNo     string             `json:"outRefundNo,omitempty"`
	ChannelRefundNo string             `json:"channelRefundNo,omitempty"`
	Amount          int64              `json:"amount,omitempty"`
	TotalAmount     int64              `json:"totalAmount,omitempty"`
	RefundAmount    int64              `json:"refundAmount,omitempty"`
	Currency        string             `json:"currency,omitempty"`
	TransTime       string             `json:"transTime,omitempty"`
	TransStatus     enums.TransStatus  `json:"transStatus,omitempty"`
	TransMessage    string             `json:"transMessage,omitempty"`
	NotifyStatus    enums.NotifyStatus `json:"notifyStatus,omitempty"`
	NotifyURL       string             `json:"notifyUrl,omitempty"`
	RefundedAt      string             `json:"refundedAt,omitempty"`
	CreatedAt       string             `json:"createdAt,omitempty"`
	UpdatedAt       string             `json:"updatedAt,omitempty"`
	Raw             map[string]any     `json:"raw,omitempty"`
}

// PrepayData 表示统一下单响应数据。
type PrepayData struct {
	Prepay PaymentParams `json:"prepay"`
	Order  Order         `json:"order"`
}

// ScanPayData 表示付款码支付响应数据。
type ScanPayData struct {
	Order Order `json:"order"`
}

// QueryOrderData 表示订单查询响应数据。
type QueryOrderData struct {
	Order
}

// RefundData 表示退款发起响应数据。
type RefundData struct {
	Refund Refund `json:"refund"`
}

// BaseRequest 表示所有受保护接口共享字段。
type BaseRequest struct {
	MID       string `json:"mid,omitempty"`
	NonceStr  string `json:"nonce_str,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Sign      string `json:"sign,omitempty"`
}

// PrepayRequest 表示统一下单请求。
type PrepayRequest struct {
	BaseRequest
	Subject    string `json:"subject,omitempty"`
	Attach     string `json:"attach,omitempty"`
	OutTradeNo string `json:"outTradeNo"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency,omitempty"`
	OpenID     string `json:"openid,omitempty"`
	NotifyURL  string `json:"notifyUrl,omitempty"`
}

// ScanPayRequest 表示付款码支付请求。
type ScanPayRequest struct {
	BaseRequest
	Subject    string `json:"subject,omitempty"`
	DeviceInfo string `json:"deviceInfo,omitempty"`
	Attach     string `json:"attach,omitempty"`
	OutTradeNo string `json:"outTradeNo"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency,omitempty"`
	AuthCode   string `json:"authCode"`
	NotifyURL  string `json:"notifyUrl,omitempty"`
}

// QueryOrderRequest 表示订单查询请求。
type QueryOrderRequest struct {
	BaseRequest
	OutTradeNo string `json:"outTradeNo"`
}

// RefundRequest 表示退款发起请求。
type RefundRequest struct {
	BaseRequest
	OutTradeNo   string `json:"outTradeNo"`
	OutRefundNo  string `json:"outRefundNo"`
	TotalAmount  int64  `json:"totalAmount"`
	RefundAmount int64  `json:"refundAmount"`
	Currency     string `json:"currency,omitempty"`
	NotifyURL    string `json:"notifyUrl,omitempty"`
	OpenID       string `json:"openid,omitempty"`
	SubOpenID    string `json:"subOpenid,omitempty"`
}

func resolveString(primary, fallback string) string {
	if trimmed := strings.TrimSpace(primary); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}

func (r PrepayRequest) payload(mid string) map[string]any {
	payload := map[string]any{
		"mid":        resolveString(r.MID, mid),
		"subject":    r.Subject,
		"outTradeNo": r.OutTradeNo,
		"amount":     r.Amount,
		"openid":     r.OpenID,
		"notifyUrl":  r.NotifyURL,
	}
	if r.Attach != "" {
		payload["attach"] = r.Attach
	}
	if r.Currency != "" {
		payload["currency"] = r.Currency
	}
	return payload
}

func (r ScanPayRequest) payload(mid string) map[string]any {
	payload := map[string]any{
		"mid":        resolveString(r.MID, mid),
		"subject":    r.Subject,
		"outTradeNo": r.OutTradeNo,
		"amount":     r.Amount,
		"authCode":   r.AuthCode,
		"notifyUrl":  r.NotifyURL,
	}
	if r.DeviceInfo != "" {
		payload["deviceInfo"] = r.DeviceInfo
	}
	if r.Attach != "" {
		payload["attach"] = r.Attach
	}
	if r.Currency != "" {
		payload["currency"] = r.Currency
	}
	return payload
}

func (r QueryOrderRequest) payload(mid string) map[string]any {
	return map[string]any{
		"mid":        resolveString(r.MID, mid),
		"outTradeNo": r.OutTradeNo,
	}
}

func (r RefundRequest) payload(mid string) map[string]any {
	payload := map[string]any{
		"mid":          resolveString(r.MID, mid),
		"outTradeNo":   r.OutTradeNo,
		"outRefundNo":  r.OutRefundNo,
		"totalAmount":  r.TotalAmount,
		"refundAmount": r.RefundAmount,
		"notifyUrl":    r.NotifyURL,
	}
	if r.Currency != "" {
		payload["currency"] = r.Currency
	}
	if r.OpenID != "" {
		payload["openid"] = r.OpenID
	}
	if r.SubOpenID != "" {
		payload["subOpenid"] = r.SubOpenID
	}
	return payload
}

// MiniAppLoginRequest 表示小程序/小游戏登录请求。
type MiniAppLoginRequest struct {
	Code     string `json:"code"`
	MID      string `json:"mid,omitempty"`
	NonceStr string `json:"nonce_str,omitempty"`
	Sign     string `json:"sign,omitempty"`
}

// MiniAppLoginData 表示小程序/小游戏登录响应数据。
type MiniAppLoginData struct {
	AppID   string `json:"appid"`             // 微信应用 AppID
	OpenID  string `json:"openid"`            // 用户唯一标识
	UnionID string `json:"unionid,omitempty"` // 用户统一标识（若存在）
}

// DecryptRequest 表示小程序数据解密请求。
type DecryptRequest struct {
	AppID    string `json:"appid"`
	OpenID   string `json:"openid"`
	Data     string `json:"data"`
	IV       string `json:"iv"`
	MID      string `json:"mid,omitempty"`
	NonceStr string `json:"nonce_str,omitempty"`
	Sign     string `json:"sign,omitempty"`
}

// DecryptData 表示小程序数据解密响应数据。
// 解密后的数据以 JSON 对象形式返回，具体字段取决于微信接口
type DecryptData map[string]any

// payload 实现请求参数构造
func (r MiniAppLoginRequest) payload(mid string) map[string]any {
	return map[string]any{
		"code":   r.Code,
		"mch_id": resolveString(r.MID, mid),
	}
}

func (r DecryptRequest) payload(mid string) map[string]any {
	payload := map[string]any{
		"appid":  r.AppID,
		"openid": r.OpenID,
		"data":   r.Data,
		"iv":     r.IV,
	}
	payload["mch_id"] = resolveString(r.MID, mid)
	return payload
}
