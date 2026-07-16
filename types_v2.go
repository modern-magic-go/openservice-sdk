package openservice

// ==================== v2 支付 API 类型 ====================

// CreateOrderRequest 表示 v2 创建订单请求。
type CreateOrderRequest struct {
	Provider   string         `json:"provider"`            // 支付渠道："alipay"
	Method     string         `json:"method"`              // 支付方式："app"
	OutTradeNo string         `json:"outTradeNo"`          // 商户订单号
	Amount     int64          `json:"amount"`              // 金额，单位：分
	Subject    string         `json:"subject"`             // 订单标题
	Currency   string         `json:"currency,omitempty"`  // 币种，默认 "CNY"
	NotifyUrl  string         `json:"notifyUrl,omitempty"` // 回调地址
	Attach     string         `json:"attach,omitempty"`    // 附加数据
	Extra      map[string]any `json:"extra,omitempty"`     // 渠道扩展参数
}

// payload 构建 v2 请求体（不含 mid——mid 通过 Header 传递）。
func (r CreateOrderRequest) payload() map[string]any {
	p := map[string]any{
		"provider":   r.Provider,
		"method":     r.Method,
		"outTradeNo": r.OutTradeNo,
		"amount":     r.Amount,
		"subject":    r.Subject,
	}
	if r.Currency != "" {
		p["currency"] = r.Currency
	}
	if r.NotifyUrl != "" {
		p["notifyUrl"] = r.NotifyUrl
	}
	if r.Attach != "" {
		p["attach"] = r.Attach
	}
	if r.Extra != nil {
		p["extra"] = r.Extra
	}
	return p
}

// AppPayParams 表示支付宝 APP 支付拉起参数。
type AppPayParams struct {
	OrderString string `json:"orderString"`
}

// CreateOrderResponse 表示 v2 创建订单响应。
// 包含 Order 的全部字段 + 渠道专属 payParams。
type CreateOrderResponse struct {
	Order
	PayParams AppPayParams `json:"payParams"`
}

// QueryOrderV2Request 表示 v2 查询订单请求。
type QueryOrderV2Request struct {
	TradeNo    string `json:"tradeNo,omitempty"`    // 平台单号
	OutTradeNo string `json:"outTradeNo,omitempty"` // 商户订单号
}

// payload 构建 v2 查询请求体。
func (r QueryOrderV2Request) payload() map[string]any {
	p := map[string]any{}
	if r.TradeNo != "" {
		p["tradeNo"] = r.TradeNo
	}
	if r.OutTradeNo != "" {
		p["outTradeNo"] = r.OutTradeNo
	}
	return p
}
