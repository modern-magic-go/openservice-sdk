package openservice

import (
	"context"
)

// MiniProgram 是微信小程序产品门面。
type MiniProgram struct {
	client *Client
}

// Login 调用小程序登录接口。
// 通过 wx.login() 获取的 code 换取用户的 openid 和 session_key。
func (m *MiniProgram) Login(ctx context.Context, req MiniAppLoginRequest) (*MiniAppLoginData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if m == nil || m.client == nil {
		return nil, ErrInvalidConfig
	}
	var data MiniAppLoginData
	if err := m.client.getJSON(ctx, miniAppLoginPath, req.payload(m.client.config.MID), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// DecryptData 调用小程序数据解密接口。
func (m *MiniProgram) DecryptData(ctx context.Context, req DecryptRequest) (DecryptData, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if m == nil || m.client == nil {
		return nil, ErrInvalidConfig
	}
	var data DecryptData
	if err := m.client.postJSON(ctx, miniAppDecryptPath, req.payload(m.client.config.MID), &data); err != nil {
		return nil, err
	}
	return data, nil
}
