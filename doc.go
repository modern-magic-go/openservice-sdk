// Package openservice provides a reusable client for the OpenService payment channel.
//
// The package only depends on the standard library and is intended to be wired by
// the host application. The host is responsible for loading Config, providing any
// custom HTTP client or transport, and handling logging/lifecycle concerns.
//
// Typical usage:
//
//	client, err := openservice.NewClient(openservice.Config{
//	    BaseURL: "https://openservice.example.com",
//	    MID:     "1900001",
//	    Secret:  "merchant-secret",
//	    Timeout: 15 * time.Second,
//	})
//	if err != nil {
//	    return err
//	}
//
//	// 支付相关接口
//	resp, err := client.Prepay(ctx, openservice.PrepayRequest{
//	    Subject:    "会员充值",
//	    OutTradeNo: "T202604100001",
//	    Amount:     100,
//	    OpenID:     "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
//	    NotifyURL:  "https://api.example.com/pay/notify",
//	})
//	if err != nil {
//	    return err
//	}
//
//	_ = resp
//
//	// 小程序登录接口
//	loginResp, err := client.MiniAppLogin(ctx, openservice.MiniAppLoginRequest{
//	    Code: "0811A11xxxxx",
//	})
//	if err != nil {
//	    return err
//	}
//	// loginResp.OpenID, loginResp.UnionID 可用于业务登录
//
//	// 小程序数据解密接口
//	decryptedData, err := client.DecryptMiniAppData(ctx, openservice.DecryptRequest{
//	    AppID:  "wx1234567890abcdef",
//	    OpenID: "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
//	    Data:   "xxxxx", // Base64 加密数据
//	    IV:     "yyyyy", // Base64 初始向量
//	})
//	if err != nil {
//	    return err
//	}
//	// decryptedData 为解密后的 map，可根据需要断言获取具体字段
//
//	_ = decryptedData
package openservice
