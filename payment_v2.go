package openservice

import (
	"context"
	"net/http"
)

const (
	createOrderV2Path = "/api/v2/payment/create"
	queryOrderV2Path  = "/api/v2/payment/query"
)

// PaymentV2 是 v2 支付产品门面（Header HMAC-SHA256 签名）。
// 通过 client.PaymentV2() 获取。
type PaymentV2 struct {
	client *Client
}

// CreateOrder 调用 v2 创建订单接口。
func (p *PaymentV2) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data CreateOrderResponse
	if err := p.client.postJSONWithHeaderAuth(ctx, createOrderV2Path, req.payload(), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// QueryOrder 调用 v2 查询订单接口。
func (p *PaymentV2) QueryOrder(ctx context.Context, req QueryOrderV2Request) (*Order, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data Order
	if err := p.client.postJSONWithHeaderAuth(ctx, queryOrderV2Path, req.payload(), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// VerifyNotification 验签 v2 回调通知（仅 Header HMAC-SHA256，不兼容旧版 MD5）。
func (p *PaymentV2) VerifyNotification(r *http.Request) error {
	if p == nil || p.client == nil {
		return ErrInvalidConfig
	}
	return verifyHeaderNotification(p.client, r)
}

// ParseNotification 解析 v2 回调通知（仅解析，不验签。需先调用 VerifyNotification 验签）。
func (p *PaymentV2) ParseNotification(r *http.Request) (*NotificationParseResult, error) {
	return parseNotificationV2JSONBody(r)
}
