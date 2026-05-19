package openservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modern-magic-go/openservice-sdk/internal/gateway"
)

const (
	prepayPath         = "/api/payment/prepay"
	scanPayPath        = "/api/payment/scanPay"
	queryOrderPath     = "/api/payment/queryOrder"
	queryRefundPath    = "/api/payment/queryRefund"
	refundPath         = "/api/payment/refund"
	getPaidUnionIDPath = "/api/payment/getPaidUnionid"
	miniAppLoginPath   = "/api/login/miniapp"
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

// Payment 返回支付产品门面。
func (c *Client) Payment() *Payment {
	return &Payment{client: c}
}

// MiniProgram 返回微信小程序产品门面。
func (c *Client) MiniProgram() *MiniProgram {
	return &MiniProgram{client: c}
}

// OfficialAccount 返回微信公众号产品门面。
func (c *Client) OfficialAccount() *OfficialAccount {
	return &OfficialAccount{client: c}
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
