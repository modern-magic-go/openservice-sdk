package gateway

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
)

type Signer interface {
	Sign(params map[string]any) string
}

type Logger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

type ErrorSet struct {
	InvalidConfig       error
	InvalidRequest      error
	InvalidResponse     error
	HTTPTransport       error
	UnexpectedStatus    error
	ResponseCodeNonZero error
	MissingData         error
}

type ResultEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type HTTPGateway struct {
	baseURL    string
	httpClient *http.Client
	signer     Signer
	logger     Logger
	errs       ErrorSet
}

func New(baseURL string, httpClient *http.Client, signer Signer, logger Logger, errs ErrorSet) *HTTPGateway {
	return &HTTPGateway{baseURL: baseURL, httpClient: httpClient, signer: signer, logger: logger, errs: errs}
}

func (g *HTTPGateway) PostJSON(ctx context.Context, path string, payload map[string]any, out any) error {
	if g == nil {
		return g.errs.InvalidConfig
	}
	if path == "" {
		return g.errs.InvalidRequest
	}

	requestPayload := make(map[string]any, len(payload)+3)
	for key, value := range payload {
		requestPayload[key] = value
	}
	requestPayload["nonce_str"] = generateNonceStr(16)
	requestPayload["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	requestPayload["sign"] = g.signer.Sign(requestPayload)

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("%w: marshal request body: %v", g.errs.InvalidRequest, err)
	}

	fullURL := g.baseURL + path
	if g.logger != nil {
		g.logger.Infow("openservice-sdk request", "channel", "openservice-sdk", "method", "POST", "url", fullURL, "payload", string(body))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: create request: %v", g.errs.InvalidRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		if g.logger != nil {
			g.logger.Errorw("openservice-sdk request failed", "channel", "openservice-sdk", "url", fullURL, "error", err)
		}
		return fmt.Errorf("%w: %v", g.errs.HTTPTransport, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status=%d", g.errs.UnexpectedStatus, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response body: %v", g.errs.InvalidResponse, err)
	}

	if g.logger != nil {
		g.logger.Infow("openservice-sdk response", "channel", "openservice-sdk", "url", fullURL, "status", resp.StatusCode, "body", string(bodyBytes))
	}

	return g.decodeResult(bodyBytes, out)
}

func (g *HTTPGateway) GetJSON(ctx context.Context, path string, queryParams map[string]any, out any) error {
	if g == nil {
		return g.errs.InvalidConfig
	}
	if path == "" {
		return g.errs.InvalidRequest
	}

	queryParams["nonce_str"] = generateNonceStr(16)
	queryParams["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	queryParams["sign"] = g.signer.Sign(queryParams)

	params := url.Values{}
	for key, value := range queryParams {
		params.Add(key, fmt.Sprintf("%v", value))
	}

	fullURL := g.baseURL + path + "?" + params.Encode()
	if g.logger != nil {
		g.logger.Infow("openservice-sdk request", "channel", "openservice-sdk", "method", "GET", "url", fullURL, "params", queryParams)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("%w: create request: %v", g.errs.InvalidRequest, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", g.errs.HTTPTransport, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status=%d", g.errs.UnexpectedStatus, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response body: %v", g.errs.InvalidResponse, err)
	}

	return g.decodeResult(bodyBytes, out)
}

func (g *HTTPGateway) decodeResult(body []byte, out any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return g.errs.InvalidResponse
	}

	var envelope ResultEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("%w: %v", g.errs.InvalidResponse, err)
	}
	if envelope.Code != 0 {
		if envelope.Message == "" {
			return fmt.Errorf("%w: code=%d", g.errs.ResponseCodeNonZero, envelope.Code)
		}
		return fmt.Errorf("%w: code=%d message=%s", g.errs.ResponseCodeNonZero, envelope.Code, envelope.Message)
	}
	if len(bytes.TrimSpace(envelope.Data)) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) {
		return g.errs.MissingData
	}
	if out == nil {
		return g.errs.InvalidResponse
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("%w: %v", g.errs.InvalidResponse, err)
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
