package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	wechat "github.com/go-pay/gopay/wechat/v3"
)

type AlipayConfig struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"alipay_public_key"`
	IsProd     bool   `json:"is_prod"`
}

type WechatConfig struct {
	AppID      string `json:"app_id"`
	MchID      string `json:"mch_id"`
	APIv3Key   string `json:"api_v3_key"`
	SerialNo   string `json:"serial_no"`
	PrivateKey string `json:"private_key"`
}

func CreateAlipayPagePay(order models.Order, siteURL string) (string, error) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'alipay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", fmt.Errorf("支付宝未配置")
	}

	var cfg AlipayConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return "", fmt.Errorf("支付宝配置不完整")
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return "", fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	client.SetNotifyUrl(siteURL + "/api/payment/callback/alipay/notify")
	client.SetReturnUrl(siteURL + "/purchase/result?order_no=" + order.OrderNo)

	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("total_amount", fmt.Sprintf("%.2f", order.Amount))
	bm.Set("subject", fmt.Sprintf("程序授权 - %s", order.OrderNo))
	bm.Set("product_code", "FAST_INSTANT_TRADE_PAY")

	payUrl, err := client.TradePagePay(context.Background(), bm)
	if err != nil {
		return "", fmt.Errorf("创建支付宝订单失败: %w", err)
	}

	return payUrl, nil
}

func CreateAlipayWapPay(order models.Order, siteURL string) (string, error) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'alipay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", fmt.Errorf("支付宝未配置")
	}

	var cfg AlipayConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return "", fmt.Errorf("支付宝配置不完整")
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return "", fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	client.SetNotifyUrl(siteURL + "/api/payment/callback/alipay/notify")
	client.SetReturnUrl(siteURL + "/purchase/result?order_no=" + order.OrderNo)

	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("total_amount", fmt.Sprintf("%.2f", order.Amount))
	bm.Set("subject", fmt.Sprintf("程序授权 - %s", order.OrderNo))
	bm.Set("product_code", "QUICK_WAP_WAY")
	bm.Set("quit_url", siteURL+"/purchase")

	payUrl, err := client.TradeWapPay(context.Background(), bm)
	if err != nil {
		return "", fmt.Errorf("创建支付宝H5订单失败: %w", err)
	}

	return payUrl, nil
}

func CreateAlipayTradePrecreate(order models.Order, siteURL string) (string, error) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'alipay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", fmt.Errorf("支付宝未配置")
	}

	var cfg AlipayConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return "", fmt.Errorf("支付宝配置不完整")
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return "", fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	client.SetNotifyUrl(siteURL + "/api/payment/callback/alipay/notify")

	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("total_amount", fmt.Sprintf("%.2f", order.Amount))
	bm.Set("subject", fmt.Sprintf("程序授权 - %s", order.OrderNo))

	resp, err := client.TradePrecreate(context.Background(), bm)
	if err != nil {
		return "", fmt.Errorf("创建支付宝预下单失败: %w", err)
	}

	if resp.Response.QrCode == "" {
		return "", fmt.Errorf("获取二维码失败")
	}

	return resp.Response.QrCode, nil
}

func VerifyAlipayNotify(notifyBody map[string][]string) (string, bool) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'alipay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", false
	}

	var cfg AlipayConfig
	json.Unmarshal([]byte(pc.Config), &cfg)

	bm := gopay.BodyMap{}
	for k, v := range notifyBody {
		if len(v) > 0 {
			bm.Set(k, v[0])
		}
	}

	ok, _ := alipay.VerifySign(cfg.PublicKey, bm)

	orderNo := ""
	if v, ok2 := notifyBody["out_trade_no"]; ok2 && len(v) > 0 {
		orderNo = v[0]
	}

	tradeStatus := ""
	if v, ok2 := notifyBody["trade_status"]; ok2 && len(v) > 0 {
		tradeStatus = v[0]
	}

	if ok && (tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED") {
		return orderNo, true
	}

	return orderNo, false
}

func CreateWechatNativePay(order models.Order, siteURL string) (string, error) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'wechat' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", fmt.Errorf("微信支付未配置")
	}

	var cfg WechatConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.MchID == "" || cfg.PrivateKey == "" {
		return "", fmt.Errorf("微信支付配置不完整")
	}

	client, err := wechat.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	expire := time.Now().Add(30 * time.Minute).Format(time.RFC3339)

	bm := gopay.BodyMap{}
	bm.Set("appid", cfg.AppID)
	bm.Set("mchid", cfg.MchID)
	bm.Set("description", fmt.Sprintf("程序授权 - %s", order.OrderNo))
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("time_expire", expire)
	bm.Set("notify_url", siteURL+"/api/payment/callback/wechat/notify")
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", int(order.Amount*100))
		bm.Set("currency", "CNY")
	})

	resp, err := client.V3TransactionNative(context.Background(), bm)
	if err != nil {
		return "", fmt.Errorf("创建微信Native订单失败: %w", err)
	}

	if resp.Response == nil || resp.Response.CodeUrl == "" {
		return "", fmt.Errorf("获取微信支付二维码失败: %s", resp.Error)
	}

	return resp.Response.CodeUrl, nil
}

func CreateWechatH5Pay(order models.Order, siteURL, clientIP string) (string, error) {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'wechat' AND enabled = ?", true).First(&pc).Error; err != nil {
		return "", fmt.Errorf("微信支付未配置")
	}

	var cfg WechatConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.MchID == "" || cfg.PrivateKey == "" {
		return "", fmt.Errorf("微信支付配置不完整")
	}

	client, err := wechat.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	expire := time.Now().Add(30 * time.Minute).Format(time.RFC3339)

	bm := gopay.BodyMap{}
	bm.Set("appid", cfg.AppID)
	bm.Set("mchid", cfg.MchID)
	bm.Set("description", fmt.Sprintf("程序授权 - %s", order.OrderNo))
	bm.Set("out_trade_no", order.OrderNo)
	bm.Set("time_expire", expire)
	bm.Set("notify_url", siteURL+"/api/payment/callback/wechat/notify")
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", int(order.Amount*100))
		bm.Set("currency", "CNY")
	})
	bm.SetBodyMap("scene_info", func(bm gopay.BodyMap) {
		bm.Set("payer_client_ip", clientIP)
		bm.SetBodyMap("h5_info", func(bm gopay.BodyMap) {
			bm.Set("type", "Wap")
		})
	})

	resp, err := client.V3TransactionH5(context.Background(), bm)
	if err != nil {
		return "", fmt.Errorf("创建微信H5订单失败: %w", err)
	}

	if resp.Response == nil || resp.Response.H5Url == "" {
		return "", fmt.Errorf("获取微信H5支付链接失败: %s", resp.Error)
	}

	return resp.Response.H5Url, nil
}

func QueryAlipayTradeStatus(orderNo string) bool {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'alipay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return false
	}
	var cfg AlipayConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.AppID == "" {
		return false
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return false
	}

	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", orderNo)

	resp, err := client.TradeQuery(context.Background(), bm)
	if err != nil {
		return false
	}

	return resp.Response.TradeStatus == "TRADE_SUCCESS" || resp.Response.TradeStatus == "TRADE_FINISHED"
}

func QueryWechatTradeStatus(orderNo string) bool {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'wechat' AND enabled = ?", true).First(&pc).Error; err != nil {
		return false
	}
	var cfg WechatConfig
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.MchID == "" {
		return false
	}

	client, err := wechat.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return false
	}

	resp, err := client.V3TransactionQueryOrder(context.Background(), wechat.OutTradeNo, orderNo)
	if err != nil {
		return false
	}

	return resp.Response.TradeState == "SUCCESS"
}

func QueryEpayTradeStatus(orderNo string) bool {
	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = 'epay' AND enabled = ?", true).First(&pc).Error; err != nil {
		return false
	}
	var cfg struct {
		APIURL     string `json:"api_url"`
		MerchantID string `json:"merchant_id"`
		APIKey     string `json:"api_key"`
	}
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.APIURL == "" {
		return false
	}

	queryURL := fmt.Sprintf("%s/api.php?act=order&pid=%s&out_trade_no=%s",
		strings.TrimRight(cfg.APIURL, "/"), cfg.MerchantID, orderNo)

	resp, err := http.Get(queryURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Code   int    `json:"code"`
		Status int    `json:"status"`
		State  string `json:"trade_status"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Status == 1 || result.State == "TRADE_SUCCESS"
}

func CompleteOrder(orderNo, tradeNo string) {
	var order models.Order
	if err := config.DB.Where("order_no = ? AND payment_status = 0", orderNo).First(&order).Error; err != nil {
		return
	}

	var pkg models.Package
	if order.PackageID != nil {
		config.DB.First(&pkg, *order.PackageID)
	}

	var program models.Program
	config.DB.First(&program, order.ProgramID)

	license := models.License{
		ProgramID:  order.ProgramID,
		LicenseKey: utils.GenerateLicenseKey(),
		Status:     0,
		Duration:   pkg.Duration,
		Remark:     fmt.Sprintf("订单购买: %s", orderNo),
	}
	config.DB.Create(&license)

	now := time.Now()
	config.DB.Model(&order).Updates(map[string]interface{}{
		"payment_status": 1,
		"trade_no":       tradeNo,
		"license_id":     license.ID,
		"paid_at":        now,
	})

	if order.BuyerEmail != "" {
		go sendOrderNotification(order.BuyerEmail, orderNo, program.Name, license.LicenseKey, order.Amount)
	}
}

func sendOrderNotification(to, orderNo, programName, licenseKey string, amount float64) {
	smtpVal := models.GetSetting("smtp_config")
	if smtpVal == "" {
		return
	}
	var cfg struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		FromName string `json:"from_name"`
		SSL      bool   `json:"ssl"`
	}
	json.Unmarshal([]byte(smtpVal), &cfg)
	if cfg.Host == "" || cfg.User == "" {
		return
	}

	siteTitle := models.GetSetting("site_title")
	if siteTitle == "" {
		siteTitle = "鱼跃授权"
	}

	subject := fmt.Sprintf("%s - 订单支付成功", siteTitle)
	body := fmt.Sprintf(`<div style="max-width:500px;margin:0 auto;font-family:system-ui,sans-serif;color:#333">
<h2 style="color:#16a34a;text-align:center">支付成功</h2>
<div style="background:#f8fafc;padding:20px;border-radius:8px;margin:16px 0">
<p><b>订单号：</b>%s</p>
<p><b>程序：</b>%s</p>
<p><b>金额：</b>%.2f 元</p>
<p style="margin-top:16px"><b>授权码：</b></p>
<p style="font-size:18px;font-weight:bold;color:#2563eb;letter-spacing:1px">%s</p>
</div>
<p style="font-size:13px;color:#9ca3af;text-align:center">请妥善保存您的授权码 - %s</p>
</div>`, orderNo, programName, amount, licenseKey, siteTitle)

	utils.SendMail(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.FromName, cfg.SSL, to, subject, body)
}
