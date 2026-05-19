package openservice

import (
	"context"
	"fmt"
	"net/url"
)

// OfficialAccount 是微信公众号产品门面。
type OfficialAccount struct {
	client *Client
}

// OAuthURL 生成微信公众号 OAuth 授权跳转地址。
func (o *OfficialAccount) OAuthURL(ctx context.Context, req OAuthRequest) (string, error) {
	if err := ensureContext(ctx); err != nil {
		return "", err
	}
	if o == nil || o.client == nil {
		return "", ErrInvalidConfig
	}
	if o.client.config.BaseURL == "" {
		return "", ErrInvalidConfig
	}
	mchID := resolveString(req.MID, o.client.config.MID)
	if mchID == "" {
		return "", ErrMissingMID
	}
	if req.RedirectURL == "" {
		return "", fmt.Errorf("%w: redirect_url is required", ErrInvalidRequest)
	}
	scope := req.Scope
	if scope == "" {
		scope = OAuthScopeBase
	}
	redirectURLEncoded := url.QueryEscape(req.RedirectURL)
	params := url.Values{}
	params.Set("mch_id", mchID)
	params.Set("scope", string(scope))
	params.Set("redirect_url", redirectURLEncoded)
	return o.client.config.BaseURL + oauthPath + "?" + params.Encode(), nil
}

// JSSDKSignature 获取微信公众号 JSSDK 签名参数。
func (o *OfficialAccount) JSSDKSignature(ctx context.Context, req JSSDKSignatureRequest) (*JSSDKSignatureData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if o == nil || o.client == nil {
		return nil, ErrInvalidConfig
	}
	if req.URL == "" {
		return nil, fmt.Errorf("%w: url is required", ErrInvalidRequest)
	}
	mid := resolveString(req.MID, o.client.config.MID)
	if mid == "" {
		return nil, ErrMissingMID
	}
	var data JSSDKSignatureData
	queryParams := req.payload(mid)
	if err := o.client.getJSON(ctx, jssdkSignPath, queryParams, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// DecryptTicket 解密微信公众号授权回调 ticket。
func (o *OfficialAccount) DecryptTicket(ticket string) (*OAuthUserInfo, error) {
	if o == nil || o.client == nil {
		return nil, ErrInvalidConfig
	}
	cfg := o.client.Config()
	return decryptTicket(ticket, cfg.AESKey, cfg.AESIv)
}
