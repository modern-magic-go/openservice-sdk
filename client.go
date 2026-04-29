package openservice

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	appLogger "github.com/modern-magic-go/logger"
)

const (
	prepayPath         = "/api/payment/prepay"
	scanPayPath        = "/api/payment/scanPay"
	queryOrderPath     = "/api/payment/queryOrder"
	refundPath         = "/api/payment/refund"
	miniAppLoginPath   = "/login/miniapp"
	miniGameLoginPath  = "/login/minigame"
	miniAppDecryptPath = "/api/decrypted/miniapp"
)

// Client 是 OpenService 模块主入口。
type Client struct {
	config     Config
	httpClient *http.Client
	signer     *Signer
	logger     *appLogger.Logger
}

// Option 定义客户端可选装配项。
type Option func(*clientOptions)

type clientOptions struct {
	httpClient *http.Client
	logger     *appLogger.Logger
}

// WithHTTPClient 注入自定义 HTTP 客户端。
func WithHTTPClient(httpClient *http.Client) Option {
	return func(opts *clientOptions) {
		opts.httpClient = httpClient
	}
}

func WithLogger(logger *appLogger.Logger) Option {
	return func(opts *clientOptions) {
		opts.logger = logger
	}
}

// NewClient 创建 OpenService 客户端。
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	cfg = normalizeConfig(cfg)

	options := &clientOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	httpClient := options.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	} else if httpClient.Timeout <= 0 {
		httpClient.Timeout = cfg.Timeout
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		signer:     NewSigner(cfg.Secret),
		logger:     options.logger,
	}, nil
}

// Config 返回客户端配置副本。
func (c *Client) Config() Config {
	if c == nil {
		return Config{}
	}
	return c.config
}

// HTTPClient 返回底层 HTTP 客户端。
func (c *Client) HTTPClient() *http.Client {
	if c == nil {
		return nil
	}
	return c.httpClient
}

// Signer 返回客户端绑定的签名器。
func (c *Client) Signer() *Signer {
	if c == nil {
		return nil
	}
	return c.signer
}

// Prepay 调用统一下单接口。
func (c *Client) Prepay(ctx context.Context, req PrepayRequest) (*PrepayData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data PrepayData
	if err := c.postJSON(ctx, prepayPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ScanPay 调用付款码支付接口。
func (c *Client) ScanPay(ctx context.Context, req ScanPayRequest) (*ScanPayData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data ScanPayData
	if err := c.postJSON(ctx, scanPayPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// QueryOrder 调用订单查询接口。
func (c *Client) QueryOrder(ctx context.Context, req QueryOrderRequest) (*QueryOrderData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data QueryOrderData
	if err := c.postJSON(ctx, queryOrderPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Refund 调用退款接口。
func (c *Client) Refund(ctx context.Context, req RefundRequest) (*RefundData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data RefundData
	if err := c.postJSON(ctx, refundPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// MiniAppLogin 调用小程序登录接口。
// 通过 wx.login() 获取的 code 换取用户的 openid 和 session_key。
// session_key 不会返回给调用方，仅存储在服务端用于后续数据解密。
func (c *Client) MiniAppLogin(ctx context.Context, req MiniAppLoginRequest) (*MiniAppLoginData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data MiniAppLoginData
	if err := c.getJSON(ctx, miniAppLoginPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// MiniGameLogin 调用小游戏登录接口。
// 与小程序登录接口完全相同，共用同一个 handler。
func (c *Client) MiniGameLogin(ctx context.Context, req MiniAppLoginRequest) (*MiniAppLoginData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data MiniAppLoginData
	if err := c.postJSON(ctx, miniGameLoginPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// DecryptMiniAppData 调用小程序数据解密接口。
// 使用存储在服务端的 session_key 解密微信返回的加密数据（如用户手机号、运动数据等）。
func (c *Client) DecryptMiniAppData(ctx context.Context, req DecryptRequest) (DecryptData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	var data DecryptData
	if err := c.postJSON(ctx, miniAppDecryptPath, req.payload(c.config.MID), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Client) postJSON(ctx context.Context, path string, payload map[string]any, out any) error {
	if c == nil {
		return ErrInvalidConfig
	}
	if path == "" {
		return ErrInvalidRequest
	}

	requestPayload := make(map[string]any, len(payload)+3)
	for key, value := range payload {
		requestPayload[key] = value
	}
	requestPayload["nonce_str"] = generateNonceStr(16)
	requestPayload["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	requestPayload["sign"] = c.signer.Sign(requestPayload)

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("%w: marshal request body: %v", ErrInvalidRequest, err)
	}

	fullURL := c.config.BaseURL + path

	if c.logger != nil {
		c.logger.Sugar().Infow("openservice request",
			"channel", "openservice",
			"method", "POST",
			"url", fullURL,
			"payload", string(body),
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: create request: %v", ErrInvalidRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.logger != nil {
			c.logger.Sugar().Errorw("openservice request failed",
				"channel", "openservice",
				"url", fullURL,
				"error", err,
			)
		}
		return fmt.Errorf("%w: %v", ErrHTTPTransport, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status=%d", ErrUnexpectedStatus, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response body: %v", ErrInvalidResponse, err)
	}

	if c.logger != nil {
		c.logger.Sugar().Infow("openservice response",
			"channel", "openservice",
			"url", fullURL,
			"status", resp.StatusCode,
			"body", string(bodyBytes),
		)
	}

	if err := decodeResult(bodyBytes, out); err != nil {
		return err
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path string, queryParams map[string]any, out any) error {
	if c == nil {
		return ErrInvalidConfig
	}
	if path == "" {
		return ErrInvalidRequest
	}

	// 添加签名所需字段（保留 code 参与签名）
	queryParams["nonce_str"] = generateNonceStr(16)
	queryParams["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	sign := c.signer.Sign(queryParams)
	queryParams["sign"] = sign

	// 构建 query string
	params := url.Values{}
	for key, value := range queryParams {
		params.Add(key, fmt.Sprintf("%v", value))
	}

	fullURL := c.config.BaseURL + path + "?" + params.Encode()
	if c.logger != nil {
		c.logger.Sugar().Infow("openservice request",
			"channel", "openservice",
			"method", "GET",
			"url", fullURL,
			"params", queryParams,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("%w: create request: %v", ErrInvalidRequest, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrHTTPTransport, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status=%d", ErrUnexpectedStatus, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response body: %v", ErrInvalidResponse, err)
	}

	if err := decodeResult(bodyBytes, out); err != nil {
		return err
	}
	return nil
}

func decodeResult(body []byte, out any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return ErrInvalidResponse
	}

	var envelope Result[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	if envelope.Code != 0 {
		if envelope.Message == "" {
			return fmt.Errorf("%w: code=%d", ErrResponseCodeNonZero, envelope.Code)
		}
		return fmt.Errorf("%w: code=%d message=%s", ErrResponseCodeNonZero, envelope.Code, envelope.Message)
	}
	if len(bytes.TrimSpace(envelope.Data)) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) {
		return ErrMissingData
	}
	if out == nil {
		return ErrInvalidResponse
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return nil
}

func ensureContext(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidRequest
	}
	return nil
}

func generateNonceStr(length int) string {
	if length <= 0 {
		length = 16
	}
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buf := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		buf[i] = charset[n.Int64()]
	}
	return string(buf)
}
