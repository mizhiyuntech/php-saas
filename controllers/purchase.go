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
}
