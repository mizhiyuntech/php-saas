package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"yuyue-auth/config"
	"yuyue-auth/models"
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
		methods = append(methods, gin.H{
			"type":  cfg.PaymentType,
			"label": labels[cfg.PaymentType],
		})
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

	var payURL string

	switch req.PaymentMethod {
	case "epay":
		payURL = buildEpayURL(pc, order, siteURL)
	case "alipay":
		payURL = fmt.Sprintf("%s/purchase/paying?order_no=%s", siteURL, orderNo)
	case "wechat":
		payURL = fmt.Sprintf("%s/purchase/paying?order_no=%s", siteURL, orderNo)
	}

	utils.Success(c, gin.H{
		"order_no": orderNo,
		"amount":   order.Amount,
		"pay_url":  payURL,
	})
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
		"return_url":   siteURL + "/purchase/result?order_no=" + order.OrderNo,
		"name":         fmt.Sprintf("授权套餐 - 订单%s", order.OrderNo),
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

func HandlePaymentCallback(c *gin.Context) {
	paymentType := c.Param("type")
	action := c.Param("action")

	switch paymentType {
	case "epay":
		handleEpayCallback(c, action)
	default:
		handleGenericCallback(c, paymentType, action)
	}
}

func handleEpayCallback(c *gin.Context, action string) {
	orderNo := c.Query("out_trade_no")
	if orderNo == "" {
		orderNo = c.PostForm("out_trade_no")
	}
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		tradeNo = c.PostForm("trade_no")
	}
	tradeStatus := c.Query("trade_status")
	if tradeStatus == "" {
		tradeStatus = c.PostForm("trade_status")
	}

	if orderNo == "" {
		c.String(200, "fail")
		return
	}

	var pc models.PaymentConfig
	config.DB.Where("payment_type = ?", "epay").First(&pc)
	var cfg struct {
		APIKey string `json:"api_key"`
	}
	json.Unmarshal([]byte(pc.Config), &cfg)

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

	expectedSign := epaySign(allParams, cfg.APIKey)
	receivedSign := allParams["sign"]
	if receivedSign != "" && receivedSign != expectedSign {
		c.String(200, "sign error")
		return
	}

	if tradeStatus == "TRADE_SUCCESS" {
		completeOrder(orderNo, tradeNo)
	}

	if action == "notify" {
		c.String(200, "success")
	} else {
		siteURL := models.GetSetting("site_url")
		if siteURL == "" {
			siteURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
		c.Redirect(302, siteURL+"/purchase/result?order_no="+orderNo)
	}
}

func handleGenericCallback(c *gin.Context, paymentType, action string) {
	orderNo := c.Query("out_trade_no")
	if orderNo == "" {
		orderNo = c.PostForm("out_trade_no")
	}
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		tradeNo = c.PostForm("trade_no")
	}

	if orderNo != "" {
		completeOrder(orderNo, tradeNo)
	}

	if action == "notify" {
		c.String(200, "success")
	} else {
		siteURL := models.GetSetting("site_url")
		if siteURL == "" {
			siteURL = fmt.Sprintf("http://%s", c.Request.Host)
		}
		c.Redirect(302, siteURL+"/purchase/result?order_no="+orderNo)
	}
}

func completeOrder(orderNo, tradeNo string) {
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
		go sendOrderEmail(order.BuyerEmail, orderNo, program.Name, license.LicenseKey, order.Amount)
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
		utils.Success(c, gin.H{
			"status":  "pirated",
			"message": record.Message,
		})
		return
	}

	var license models.License
	result := config.DB.Where("device_info LIKE ?", "%"+req.Domain+"%").First(&license)
	if result.Error == nil && license.Status == 1 {
		utils.Success(c, gin.H{
			"status":       "authorized",
			"program_name": "",
			"license_key":  license.LicenseKey[:8] + "****",
			"expires_at":   license.ExpiresAt,
		})

		var program models.Program
		if config.DB.First(&program, license.ProgramID).Error == nil {
			c.JSON(200, gin.H{
				"code": 0, "message": "success",
				"data": gin.H{
					"status":       "authorized",
					"program_name": program.Name,
					"license_key":  license.LicenseKey[:8] + "****",
					"expires_at":   license.ExpiresAt,
				},
			})
		}
		return
	}

	utils.Success(c, gin.H{
		"status":  "unknown",
		"message": "未查询到该域名的授权信息",
	})
}

func sendOrderEmail(to, orderNo, programName, licenseKey string, amount float64) {
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

	subject := fmt.Sprintf("%s - 订单支付成功通知", siteTitle)
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
