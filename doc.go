// Package openservice provides the OpenService SDK client for payment, login,
// JSSDK, notification parsing, and signature integration.
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
//	    AESKey:  "ticket-aes-key",
//	    AESIv:   "ticket-aes-iv",
//	    Timeout: 15 * time.Second,
//	})
//	if err != nil {
//	    return err
//	}
//
//	// 支付门面：公众号和小程序都使用统一下单；openid 由业务项目提前取得。
//	paymentResp, err := client.Payment().Prepay(ctx, openservice.PrepayRequest{
//	    Subject:    "会员充值",
//	    OutTradeNo: "T202604100001",
//	    Amount:     100,
//	    OpenID:     "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
//	    NotifyURL:  "https://api.example.com/pay/notify",
//	})
//	if err != nil {
//	    return err
//	}
//	// paymentResp.Prepay 可直接传给前端拉起微信支付。
//	_ = paymentResp
//
//	// 公众号门面：OAuth、JSSDK 签名和 ticket 解密按公众号产品语义聚合。
//	officialAccount := client.OfficialAccount()
//	oauthURL, err := officialAccount.OAuthURL(ctx, openservice.OAuthRequest{
//	    Scope:       openservice.OAuthScopeUserInfo,
//	    RedirectURL: "https://example.com/wechat/callback",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = oauthURL
//
//	jssdkSignature, err := officialAccount.JSSDKSignature(ctx, openservice.JSSDKSignatureRequest{
//	    URL: "https://example.com/page",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = jssdkSignature
//
//	userInfo, err := officialAccount.DecryptTicket("ticket-from-callback")
//	if err != nil {
//	    return err
//	}
//	_ = userInfo
//
//	// 支付 / 退款回调通知解析：HTTP handler 仍负责返回纯文本 SUCCESS。
//	parsed, err := openservice.ParsePaymentNotification(r.PostForm, "merchant-secret")
//	if err != nil {
//	    return err
//	}
//	switch parsed.Kind {
//	case openservice.NotificationKindPayment:
//	    _ = parsed.Payment.OutTradeNo
//	case openservice.NotificationKindRefund:
//	    _ = parsed.Refund.OutRefundNo
//	}
//
//	// 小程序身份能力也在小程序门面下。
//	miniProgram := client.MiniProgram()
//	loginResp, err := miniProgram.Login(ctx, openservice.MiniAppLoginRequest{
//	    Code: "0811A11xxxxx",
//	})
//	if err != nil {
//	    return err
//	}
//	// loginResp.OpenID, loginResp.UnionID 可用于业务登录。
//
//	decryptedData, err := miniProgram.DecryptData(ctx, openservice.DecryptRequest{
//	    AppID:  loginResp.AppID,
//	    OpenID: loginResp.OpenID,
//	    Data:   "xxxxx", // 小程序授权回调返回的 encryptedData
//	    IV:     "yyyyy", // 小程序授权回调返回的 iv
//	})
//	if err != nil {
//	    return err
//	}
//	// decryptedData 为解密后的 map，可根据需要断言获取具体字段
//
//	_ = decryptedData
package openservice
