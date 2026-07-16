// Package openservice provides the OpenService SDK client for OpenService
// payment, WeChat Mini Program, Official Account, notification parsing, and
// signature integration.
//
// The package only depends on the standard library and is intended to be wired by
// the host application. The host is responsible for loading Config, providing any
// custom HTTP client or transport, and handling logging/lifecycle concerns.
//
// Client setup:
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
//	_ = client.Config()
//	_ = client.HTTPClient()
//	_ = client.Signer()
//
// Custom HTTP client and logger can be injected when constructing the client:
//
//	client, err = openservice.NewClient(cfg,
//	    openservice.WithHTTPClient(httpClient),
//	    openservice.WithLogger(logger),
//	)
//	if err != nil {
//	    return err
//	}
//
// Payment facade:
//
// The Payment facade is the single entry for payment APIs. Official Account and
// Mini Program payment both use Prepay; openid is obtained by the business
// application before calling the SDK. Business code only passes business fields;
// mid, nonce_str, timestamp, and sign are completed by the signed gateway.
//
//	payment := client.Payment()
//	prepayResp, err := payment.Prepay(ctx, openservice.PrepayRequest{
//	    Subject:    "会员充值",
//	    OutTradeNo: "T202604100001",
//	    Amount:     100,
//	    OpenID:     "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
//	    NotifyURL:  "https://api.example.com/pay/notify",
//	})
//	if err != nil {
//	    return err
//	}
//	// prepayResp.Prepay can be passed to the frontend to invoke WeChat payment.
//	_ = prepayResp.Prepay
//	_ = prepayResp.Order
//
//	scanResp, err := payment.ScanPay(ctx, openservice.ScanPayRequest{
//	    Subject:    "线下收款",
//	    OutTradeNo: "T202604100002",
//	    Amount:     100,
//	    AuthCode:   "付款码",
//	    NotifyURL:  "https://api.example.com/pay/notify",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = scanResp.Order
//
//	orderResp, err := payment.QueryOrder(ctx, openservice.QueryOrderRequest{
//	    OutTradeNo: "T202604100001",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = orderResp.TransStatus
//
//	refundCreated, err := payment.Refund(ctx, openservice.RefundRequest{
//	    OutTradeNo:   "T202604100001",
//	    OutRefundNo:  "R202604100001",
//	    TotalAmount:  100,
//	    RefundAmount: 20,
//	    NotifyURL:    "https://api.example.com/refund/notify",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = refundCreated.Refund
//
//	refundResp, err := payment.QueryRefund(ctx, openservice.QueryRefundRequest{
//	    OutRefundNo: "R202604100001",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = refundResp.TransStatus
//
//	unionResp, err := payment.GetPaidUnionID(ctx, openservice.GetPaidUnionIDRequest{
//	    OutTradeNo: "T202604100001",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = unionResp.UnionID
//
// V2 Payment facade (Alipay APP pay):
//
// The v2 payment API uses Header-based HMAC-SHA256 signature (X-Auth-* headers)
// instead of body-embedded MD5. Use PaymentV2() for v2 APIs.
//
//	paymentV2 := client.PaymentV2()
//
//	// Create an Alipay APP payment order.
//	createResp, err := paymentV2.CreateOrder(ctx, openservice.CreateOrderRequest{
//	    Provider:   "alipay",
//	    Method:     "app",
//	    OutTradeNo: "ORDER20260716001",
//	    Amount:     100, // 1.00 CNY
//	    Subject:    "测试商品",
//	    NotifyUrl:  "https://api.example.com/notify",
//	})
//	if err != nil {
//	    return err
//	}
//	// createResp.PayParams.OrderString is the signed query string for Alipay SDK.
//	_ = createResp.PayParams.OrderString
//
//	// Query an order by outTradeNo.
//	order, err := paymentV2.QueryOrder(ctx, openservice.QueryOrderV2Request{
//	    OutTradeNo: "ORDER20260716001",
//	})
//	if err != nil {
//	    return err
//	}
//	_ = order.TransStatus
//	_ = order.Amount
//
// V2 Payment notifications:
//
// PaymentV2 notification methods verify and parse callbacks using
// Header HMAC-SHA256 only (no form/MD5 fallback).
//
//	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
//	    if err := paymentV2.VerifyNotification(r); err != nil {
//	        w.WriteHeader(403)
//	        return
//	    }
//	    parsed, err := paymentV2.ParseNotification(r)
//	    if err != nil {
//	        w.WriteHeader(400)
//	        return
//	    }
//	    // parsed.Payment.OutTradeNo / TradeNo / TransStatus / Amount ...
//	    w.Write([]byte("SUCCESS"))
//	})
//
// Payment notifications (v1, form-urlencoded):
//
// Payment and refund notifications are form-urlencoded callbacks sent by
// OpenService. The SDK verifies and parses them, but the HTTP handler still owns
// idempotency, business persistence, and returning plain text SUCCESS.
//
//	if err := payment.VerifyNotification(r.PostForm); err != nil {
//	    return err
//	}
//	parsed, err := payment.ParseNotification(r.PostForm)
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
// Official Account facade:
//
// OAuth URL generation, JSSDK signature retrieval, and encrypted ticket decoding
// are grouped by Official Account product semantics.
//
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
//	accessToken, err := officialAccount.AccessToken(ctx)
//	if err != nil {
//	    return err
//	}
//	_ = accessToken.AccessToken
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
// Mini Program facade:
//
//	miniProgram := client.MiniProgram()
//	miniAccessToken, err := miniProgram.AccessToken(ctx)
//	if err != nil {
//	    return err
//	}
//	_ = miniAccessToken.AccessToken
//
//	loginResp, err := miniProgram.Login(ctx, openservice.MiniAppLoginRequest{
//	    Code: "0811A11xxxxx",
//	})
//	if err != nil {
//	    return err
//	}
//	// loginResp.OpenID and loginResp.UnionID can be used by business login.
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
//	_ = decryptedData
//
// Signature helpers:
//
// The gateway signs protected OpenService requests automatically. Signer helpers
// are exposed for local signature debugging and callback-like integrations.
//
//	signer := openservice.NewSigner("merchant-secret")
//	params := map[string]any{"mid": "1900001", "outTradeNo": "T202604100001"}
//	signString := signer.BuildSignString(params)
//	params["sign"] = signer.Sign(params)
//	_ = signString
//	_ = signer.VerifySign(params)
package openservice
