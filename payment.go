package openservice

import "context"

// Payment 是支付产品门面。
type Payment struct {
	client *Client
}

// Prepay 调用统一下单接口。
func (p *Payment) Prepay(ctx context.Context, req PrepayRequest) (*PrepayData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data PrepayData
	if err := p.client.postJSON(ctx, prepayPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ScanPay 调用付款码支付接口。
func (p *Payment) ScanPay(ctx context.Context, req ScanPayRequest) (*ScanPayData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data ScanPayData
	if err := p.client.postJSON(ctx, scanPayPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// QueryOrder 调用订单查询接口。
func (p *Payment) QueryOrder(ctx context.Context, req QueryOrderRequest) (*QueryOrderData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data QueryOrderData
	if err := p.client.postJSON(ctx, queryOrderPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// QueryRefund 调用退款查询接口。
func (p *Payment) QueryRefund(ctx context.Context, req QueryRefundRequest) (*QueryRefundData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data QueryRefundData
	if err := p.client.postJSON(ctx, queryRefundPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Refund 调用退款接口。
func (p *Payment) Refund(ctx context.Context, req RefundRequest) (*RefundData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data RefundData
	if err := p.client.postJSON(ctx, refundPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetPaidUnionID 调用支付后 UnionID 查询接口。
func (p *Payment) GetPaidUnionID(ctx context.Context, req GetPaidUnionIDRequest) (*GetPaidUnionIDData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if p == nil || p.client == nil {
		return nil, ErrInvalidConfig
	}
	var data GetPaidUnionIDData
	if err := p.client.postJSON(ctx, getPaidUnionIDPath, req.payload(p.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}
