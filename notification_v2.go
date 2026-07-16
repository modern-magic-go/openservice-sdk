package openservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// VerifyNotificationV2 验签 v2 回调通知，自动检测签名模式。
// header 模式（X-Auth-Signature 存在）→ HMAC-SHA256 验签；
// form 模式 → 回退到现有 MD5 验签。
func (p *Payment) VerifyNotificationV2(r *http.Request) error {
	if p == nil || p.client == nil {
		return ErrInvalidConfig
	}
	if r == nil {
		return fmt.Errorf("%w: request is nil", ErrInvalidRequest)
	}

	if r.Header.Get("X-Auth-Signature") != "" {
		return verifyHeaderNotification(p.client, r)
	}

	// form 模式：回退到现有 v1 验签
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("%w: parse form: %v", ErrInvalidRequest, err)
	}
	return verifyPaymentNotification(r.PostForm, p.client.config.Secret)
}

// ParseNotificationV2 验签并解析 v2 回调通知。
// 内部先调 VerifyNotificationV2 验签，通过后再解析 body。
func (p *Payment) ParseNotificationV2(r *http.Request) (*NotificationParseResult, error) {
	if err := p.VerifyNotificationV2(r); err != nil {
		return nil, err
	}

	// 验签已通过，直接解析 body
	if r.Header.Get("X-Auth-Signature") != "" {
		return parseNotificationV2JSONBody(r)
	}
	return parsePaymentNotificationInput(r.PostForm)
}

// verifyHeaderNotification 验证 header 模式（HMAC-SHA256）回调签名。
func verifyHeaderNotification(client *Client, r *http.Request) error {
	mid := r.Header.Get("X-Auth-Mid")
	timestamp := r.Header.Get("X-Auth-Timestamp")
	nonce := r.Header.Get("X-Auth-Nonce")
	signature := r.Header.Get("X-Auth-Signature")

	if mid == "" || timestamp == "" || nonce == "" || signature == "" {
		return fmt.Errorf("%w: missing X-Auth-* headers", ErrInvalidRequest)
	}

	body, err := readAndRestoreBody(r)
	if err != nil {
		return fmt.Errorf("%w: read body: %v", ErrInvalidRequest, err)
	}

	signer := NewSigner(client.config.Secret)
	if !signer.VerifyHeader(r.Method, r.URL.Path, body, timestamp, nonce, signature, client.config.Secret) {
		return ErrInvalidSign
	}

	return nil
}

// parseNotificationV2JSONBody 解析 JSON body 回调（不验签）。
func parseNotificationV2JSONBody(r *http.Request) (*NotificationParseResult, error) {
	body, err := readAndRestoreBody(r)
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", ErrInvalidRequest, err)
	}

	// 先解析为 map 保留原始数据
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: json decode: %v", ErrInvalidResponse, err)
	}

	rawStrings := make(map[string]string, len(raw))
	for k, v := range raw {
		rawStrings[k] = normalizeSignValue(v)
	}

	if isRefundNotification(rawStrings) {
		refund, err := parseRefundNotification(rawStrings)
		if err != nil {
			return nil, err
		}
		refund.MID = r.Header.Get("X-Auth-Mid")
		return &NotificationParseResult{Kind: NotificationKindRefund, Refund: refund, Raw: rawStrings}, nil
	}

	payment, err := parsePaymentNotification(rawStrings)
	if err != nil {
		return nil, err
	}
	// mid 仅通过 Header 传递，body 中不含；注入到解析结果方便业务使用
	payment.MID = r.Header.Get("X-Auth-Mid")
	return &NotificationParseResult{Kind: NotificationKindPayment, Payment: payment, Raw: rawStrings}, nil
}

// readAndRestoreBody 读取 r.Body 的全部内容，并用 bytes.NewReader 恢复，使 body 可被重复读取。
func readAndRestoreBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
