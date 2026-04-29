package openservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modern-magic-go/openservice/internal/gateway"
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
	logger     Logger
	gateway    *gateway.HTTPGateway
}

type Logger interface {
	Sugar() SugaredLogger
}

type SugaredLogger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

// Option 定义客户端可选装配项。
type Option func(*clientOptions)

type clientOptions struct {
	httpClient *http.Client
	logger     Logger
}

// WithHTTPClient 注入自定义 HTTP 客户端。
func WithHTTPClient(httpClient *http.Client) Option {
	return func(opts *clientOptions) {
		opts.httpClient = httpClient
	}
}

func WithLogger(logger Logger) Option {
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
		gateway: gateway.New(
			cfg.BaseURL,
			httpClient,
			NewSigner(cfg.Secret),
			loggerAdapter{logger: options.logger},
			gateway.ErrorSet{
				InvalidConfig:       ErrInvalidConfig,
				InvalidRequest:      ErrInvalidRequest,
				InvalidResponse:     ErrInvalidResponse,
				HTTPTransport:       ErrHTTPTransport,
				UnexpectedStatus:    ErrUnexpectedStatus,
				ResponseCodeNonZero: ErrResponseCodeNonZero,
				MissingData:         ErrMissingData,
			},
		),
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
	if c == nil || c.gateway == nil {
		return ErrInvalidConfig
	}
	return c.gateway.PostJSON(ctx, path, payload, out)
}

func (c *Client) getJSON(ctx context.Context, path string, queryParams map[string]any, out any) error {
	if c == nil || c.gateway == nil {
		return ErrInvalidConfig
	}
	return c.gateway.GetJSON(ctx, path, queryParams, out)
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

type loggerAdapter struct{ logger Logger }

func (l loggerAdapter) Infow(msg string, keysAndValues ...any) {
	if l.logger == nil {
		return
	}
	l.logger.Sugar().Infow(msg, keysAndValues...)
}

func (l loggerAdapter) Errorw(msg string, keysAndValues ...any) {
	if l.logger == nil {
		return
	}
	l.logger.Sugar().Errorw(msg, keysAndValues...)
}
