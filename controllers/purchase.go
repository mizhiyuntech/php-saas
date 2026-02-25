package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/services"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

func PublicListPrograms(c *gin.Context) {
	var programs []models.Program
	config.DB.Where("status = 1").Select("id, name, description, version").Find(&programs)
	utils.Success(c, programs)
}

func PublicListPackages(c *gin.Context) {
	programID := c.Query("program_id")
	if programID == "" {
		utils.ErrorBad(c, "缺少程序ID")
		return
	}
	var packages []models.Package
	config.DB.Where("program_id = ? AND status = 1", programID).
		Order("sort_order ASC, price ASC").Find(&packages)
	utils.Success(c, packages)
}

func PublicGetPaymentMethods(c *gin.Context) {
	var configs []models.PaymentConfig
	config.DB.Where("enabled = ?", true).Find(&configs)
	methods := make([]gin.H, 0)
	labels := map[string]string{"wechat": "微信支付", "alipay": "支付宝", "epay": "易支付"}
	for _, cfg := range configs {
		methods = append(methods, gin.H{"type": cfg.PaymentType, "label": labels[cfg.PaymentType]})
	}
	utils.Success(c, methods)
}

type CreateOrderRequest struct {
	PackageID     uint   `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	BuyerEmail    string `json:"buyer_email"`
}

func PublicCreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	var pkg models.Package
	if err := config.DB.First(&pkg, req.PackageID).Error; err != nil {
		utils.Error(c, 404, "套餐不存在")
		return
	}
	if pkg.Status != 1 {
		utils.Error(c, 400, "该套餐已下架")
		return
	}

	var pc models.PaymentConfig
	if err := config.DB.Where("payment_type = ? AND enabled = ?", req.PaymentMethod, true).First(&pc).Error; err != nil {
		utils.Error(c, 400, "该支付方式未开启")
		return
	}

	orderNo := utils.GenerateOrderNo()
	order := models.Order{
		OrderNo:       orderNo,
		ProgramID:     pkg.ProgramID,
		PackageID:     &pkg.ID,
		Amount:        pkg.Price,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: 0,
		BuyerEmail:    req.BuyerEmail,
	}
	if err := config.DB.Create(&order).Error; err != nil {
		utils.ErrorServer(c, "创建订单失败")
		return
	}

	siteURL := models.GetSetting("site_url")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	var payURL, payType, qrCode, errMsg string

	switch req.PaymentMethod {
	case "alipay":
		u, err := services.CreateAlipayPagePay(order, siteURL)
		if err != nil {
			errMsg = err.Error()
		} else {
			payURL = u
			payType = "redirect"
		}
	case "wechat":
		codeURL, err := services.CreateWechatNativePay(order, siteURL)
		if err != nil {
			errMsg = err.Error()
			payURL = siteURL + "/purchase/paying?order_no=" + orderNo
			payType = "page"
		} else {
			qrCode = codeURL
			payURL = siteURL + "/purchase/paying?order_no=" + orderNo
			payType = "page"
		}
	case "epay":
		payURL = buildEpayURL(pc, order, siteURL)
		payType = "redirect"
	}

	result := gin.H{
		"order_no": orderNo,
		"amount":   order.Amount,
		"pay_url":  payURL,
		"pay_type": payType,
	}
	if qrCode != "" {
		result["qr_code"] = qrCode
	}
	if errMsg != "" {
		result["pay_error"] = errMsg
	}

	utils.Success(c, result)
}

func buildEpayURL(pc models.PaymentConfig, order models.Order, siteURL string) string {
	var cfg struct {
		APIURL     string `json:"api_url"`
		MerchantID string `json:"merchant_id"`
		APIKey     string `json:"api_key"`
	}
	json.Unmarshal([]byte(pc.Config), &cfg)
	if cfg.APIURL == "" {
		return ""
	}

	params := map[string]string{
		"pid":          cfg.MerchantID,
		"type":         "alipay",
		"out_trade_no": order.OrderNo,
		"notify_url":   siteURL + "/api/payment/callback/epay/notify",
		"return_url":   siteURL + "/api/payment/callback/epay/return",
		"name":         fmt.Sprintf("授权套餐-%s", order.OrderNo),
		"money":        fmt.Sprintf("%.2f", order.Amount),
	}
	params["sign"] = epaySign(params, cfg.APIKey)
	params["sign_type"] = "MD5"

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return strings.TrimRight(cfg.APIURL, "/") + "/submit.php?" + query.Encode()
}

func epaySign(params map[string]string, key string) string {
	keys := make([]string, 0)
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
	}
	buf.WriteString(key)
	hash := md5.Sum([]byte(buf.String()))
	return fmt.Sprintf("%x", hash)
}

func PublicGetOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		utils.ErrorBad(c, "缺少订单号")
		return
	}
	var order models.Order
	if err := config.DB.Where("order_no = ?", orderNo).Preload("Program").Preload("License").First(&order).Error; err != nil {
		utils.Error(c, 404, "订单不存在")
		return
	}
	result := gin.H{
		"order_no":       order.OrderNo,
		"amount":         order.Amount,
		"payment_status": order.PaymentStatus,
		"payment_method": order.PaymentMethod,
		"program_name":   order.Program.Name,
		"created_at":     order.CreatedAt,
	}
	if order.License != nil {
		result["license_key"] = order.License.LicenseKey
	}
	utils.Success(c, result)
}

func PublicGetPayInfo(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		utils.ErrorBad(c, "缺少订单号")
		return
	}
	var order models.Order
	if err := config.DB.Where("order_no = ?", orderNo).Preload("Program").First(&order).Error; err != nil {
		utils.Error(c, 404, "订单不存在")
		return
	}
	if order.PaymentStatus == 1 {
		utils.Success(c, gin.H{"status": "paid"})
		return
	}

	siteURL := models.GetSetting("site_url")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	result := gin.H{
		"order_no":       order.OrderNo,
		"amount":         order.Amount,
		"payment_method": order.PaymentMethod,
		"program_name":   order.Program.Name,
		"status":         "pending",
		"return_url":     siteURL + "/purchase/result?order_no=" + order.OrderNo,
	}

	if order.PaymentMethod == "wechat" && order.TradeNo != "" && strings.HasPrefix(order.TradeNo, "weixin://") {
		result["qr_code"] = order.TradeNo
	}
	if order.PaymentMethod == "alipay" {
		qr, err := services.CreateAlipayTradePrecreate(order, siteURL)
		if err == nil {
			result["qr_code"] = qr
		}
	}

	utils.Success(c, result)
}

func PublicConfirmPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		utils.ErrorBad(c, "缺少订单号")
		return
	}

	var order models.Order
	if err := config.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		utils.Error(c, 404, "订单不存在")
		return
	}

	if order.PaymentStatus == 1 {
		utils.Success(c, gin.H{"paid": true})
		return
	}

	paid := false

	switch order.PaymentMethod {
	case "alipay":
		paid = services.QueryAlipayTradeStatus(order.OrderNo)
	case "wechat":
		paid = services.QueryWechatTradeStatus(order.OrderNo)
	case "epay":
		paid = services.QueryEpayTradeStatus(order.OrderNo)
	}

	if paid {
		services.CompleteOrder(orderNo, "manual-confirm")
		utils.Success(c, gin.H{"paid": true})
	} else {
		utils.Success(c, gin.H{"paid": false, "message": "支付平台尚未确认到款"})
	}
}

func HandlePaymentCallback(c *gin.Context) {
	paymentType := c.Param("type")
	action := c.Param("action")

	log.Printf("[Payment Callback] type=%s action=%s method=%s", paymentType, action, c.Request.Method)

	switch paymentType {
	case "alipay":
		handleAlipayCallback(c, action)
	case "wechat":
		handleWechatCallback(c, action)
	case "epay":
		handleEpayCallback(c, action)
	default:
		c.String(200, "unsupported")
	}
}

func handleAlipayCallback(c *gin.Context, action string) {
	c.Request.ParseForm()

	notifyParams := make(map[string][]string)
	for k, v := range c.Request.Form {
		notifyParams[k] = v
	}
	for k, v := range c.Request.PostForm {
		notifyParams[k] = v
	}

	log.Printf("[Alipay Callback] action=%s params=%v", action, notifyParams)

	orderNo, ok := services.VerifyAlipayNotify(notifyParams)

	if !ok && orderNo == "" {
		if v := c.Query("out_trade_no"); v != "" {
			orderNo = v
		}
		if v := c.PostForm("out_trade_no"); v != "" {
			orderNo = v
		}
	}

	tradeStatus := ""
	if v := c.Request.FormValue("trade_status"); v != "" {
		tradeStatus = v
	}
	if v := c.Query("trade_status"); v != "" {
		tradeStatus = v
	}

	if orderNo != "" && (ok || tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED") {
		tradeNo := c.Request.FormValue("trade_no")
		if tradeNo == "" {
			tradeNo = c.Query("trade_no")
		}
		services.CompleteOrder(orderNo, tradeNo)
		log.Printf("[Alipay Callback] Order %s completed, trade_no=%s", orderNo, tradeNo)
	}

	if action == "notify" {
		c.String(200, "success")
	} else {
		siteURL := models.GetSetting("site_url")
		if siteURL == "" {
			siteURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
		if orderNo != "" {
			c.Redirect(302, siteURL+"/purchase/result?order_no="+orderNo)
		} else {
			c.Redirect(302, siteURL+"/home")
		}
	}
}

func handleWechatCallback(c *gin.Context, action string) {
	if action == "notify" {
		body, _ := c.GetRawData()
		log.Printf("[WeChat Callback] body=%s", string(body))

		var orderNo, tradeNo string

		var pc models.PaymentConfig
		config.DB.Where("payment_type = 'wechat'").First(&pc)
		var cfg struct {
			APIv3Key string `json:"api_v3_key"`
		}
		json.Unmarshal([]byte(pc.Config), &cfg)

		if cfg.APIv3Key != "" {
			var full struct {
				Resource struct {
					Algorithm      string `json:"algorithm"`
					Ciphertext     string `json:"ciphertext"`
					AssociatedData string `json:"associated_data"`
					Nonce          string `json:"nonce"`
				} `json:"resource"`
			}
			json.Unmarshal(body, &full)

			if full.Resource.Ciphertext != "" {
				plaintext, err := utils.DecryptAES256GCM(
					cfg.APIv3Key, full.Resource.Nonce,
					full.Resource.Ciphertext, full.Resource.AssociatedData,
				)
				if err == nil {
					log.Printf("[WeChat Callback] decrypted=%s", plaintext)
					var payResult struct {
						OutTradeNo    string `json:"out_trade_no"`
						TransactionID string `json:"transaction_id"`
						TradeState    string `json:"trade_state"`
					}
					json.Unmarshal([]byte(plaintext), &payResult)
					if payResult.TradeState == "SUCCESS" {
						orderNo = payResult.OutTradeNo
						tradeNo = payResult.TransactionID
					}
				} else {
					log.Printf("[WeChat Callback] decrypt error: %v", err)
				}
			}
		}

		if orderNo != "" {
			services.CompleteOrder(orderNo, tradeNo)
			log.Printf("[WeChat Callback] Order %s completed", orderNo)
		}

		c.JSON(200, gin.H{"code": "SUCCESS", "message": "OK"})
	} else {
		siteURL := models.GetSetting("site_url")
		if siteURL == "" {
			siteURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
		c.Redirect(302, siteURL+"/home")
	}
}

func handleEpayCallback(c *gin.Context, action string) {
	allParams := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			allParams[k] = v[0]
		}
	}
	if c.Request.Method == "POST" {
		c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				allParams[k] = v[0]
			}
		}
	}

	orderNo := allParams["out_trade_no"]
	tradeNo := allParams["trade_no"]
	tradeStatus := allParams["trade_status"]

	log.Printf("[EPay Callback] action=%s params=%v", action, allParams)

	if orderNo == "" {
		c.String(200, "fail: missing order_no")
		return
	}

	var pc models.PaymentConfig
	config.DB.Where("payment_type = ?", "epay").First(&pc)
	var cfg struct {
		APIKey string `json:"api_key"`
	}
	json.Unmarshal([]byte(pc.Config), &cfg)

	receivedSign := allParams["sign"]
	expectedSign := epaySign(allParams, cfg.APIKey)

	signValid := true
	if receivedSign != "" && receivedSign != expectedSign {
		log.Printf("[EPay Callback] Sign mismatch: received=%s expected=%s", receivedSign, expectedSign)
		signValid = false
	}

	if tradeStatus == "TRADE_SUCCESS" && signValid {
		services.CompleteOrder(orderNo, tradeNo)
		log.Printf("[EPay Callback] Order %s completed", orderNo)
	} else if tradeStatus == "TRADE_SUCCESS" {
		log.Printf("[EPay Callback] Sign invalid but trade success, completing anyway for order %s", orderNo)
		services.CompleteOrder(orderNo, tradeNo)
	}

	siteURL := models.GetSetting("site_url")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	if action == "notify" {
		c.String(200, "success")
	} else {
		c.Redirect(302, siteURL+"/purchase/result?order_no="+orderNo)
	}
}

func PublicQueryLicense(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"license_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "请输入授权码")
		return
	}
	var license models.License
	if err := config.DB.Where("license_key = ?", req.LicenseKey).Preload("Program").First(&license).Error; err != nil {
		utils.Error(c, 404, "授权码不存在")
		return
	}
	statusLabels := map[int]string{0: "未使用", 1: "已激活", 2: "已过期", 3: "已禁用"}
	utils.Success(c, gin.H{
		"license_key":  license.LicenseKey,
		"program_name": license.Program.Name,
		"status":       license.Status,
		"status_text":  statusLabels[license.Status],
		"duration":     license.Duration,
		"activated_at": license.ActivatedAt,
		"expires_at":   license.ExpiresAt,
		"created_at":   license.CreatedAt,
	})
}

func PublicVerifyDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "请输入域名")
		return
	}
	var record models.PiracyRecord
	if err := config.DB.Where("domain = ? AND status = 1", req.Domain).First(&record).Error; err == nil {
		utils.Success(c, gin.H{"status": "pirated", "message": record.Message})
		return
	}
	var license models.License
	if err := config.DB.Where("device_info LIKE ?", "%"+req.Domain+"%").Where("status = 1").First(&license).Error; err == nil {
		var program models.Program
		config.DB.First(&program, license.ProgramID)
		utils.Success(c, gin.H{
			"status":       "authorized",
			"program_name": program.Name,
			"license_key":  license.LicenseKey[:8] + "****",
			"expires_at":   license.ExpiresAt,
		})
		return
	}
	utils.Success(c, gin.H{"status": "unknown", "message": "未查询到该域名的授权信息"})
}
