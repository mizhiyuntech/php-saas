package controllers

import (
	"encoding/json"
	"fmt"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type PaymentConfigRequest struct {
	PaymentType string          `json:"payment_type" binding:"required"`
	Enabled     bool            `json:"enabled"`
	Config      json.RawMessage `json:"config"`
}

func GetPaymentConfigs(c *gin.Context) {
	var configs []models.PaymentConfig
	config.DB.Find(&configs)

	siteURL := models.GetSetting("site_url")
	if siteURL == "" {
		siteURL = fmt.Sprintf("http://%s", c.Request.Host)
	}

	result := make([]gin.H, 0)
	for _, cfg := range configs {
		item := gin.H{
			"id":           cfg.ID,
			"payment_type": cfg.PaymentType,
			"enabled":      cfg.Enabled,
			"config":       json.RawMessage(cfg.Config),
			"created_at":   cfg.CreatedAt,
			"updated_at":   cfg.UpdatedAt,
		}

		switch cfg.PaymentType {
		case "wechat":
			item["return_url"] = siteURL + "/api/payment/callback/wechat/return"
			item["notify_url"] = siteURL + "/api/payment/callback/wechat/notify"
			item["auth_dir"] = siteURL + "/api/payment/wechat/"
		case "alipay":
			item["return_url"] = siteURL + "/api/payment/callback/alipay/return"
			item["notify_url"] = siteURL + "/api/payment/callback/alipay/notify"
		case "epay":
			item["return_url"] = siteURL + "/api/payment/callback/epay/return"
			item["notify_url"] = siteURL + "/api/payment/callback/epay/notify"
		}

		result = append(result, item)
	}

	if len(result) == 0 {
		defaultTypes := []string{"wechat", "alipay", "epay"}
		for _, pt := range defaultTypes {
			pc := models.PaymentConfig{
				PaymentType: pt,
				Enabled:     false,
				Config:      "{}",
			}
			config.DB.Create(&pc)
		}
		config.DB.Find(&configs)
		for _, cfg := range configs {
			result = append(result, gin.H{
				"id":           cfg.ID,
				"payment_type": cfg.PaymentType,
				"enabled":      cfg.Enabled,
				"config":       json.RawMessage(cfg.Config),
			})
		}
	}

	utils.Success(c, result)
}

func UpdatePaymentConfig(c *gin.Context) {
	var req PaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	var pc models.PaymentConfig
	result := config.DB.Where("payment_type = ?", req.PaymentType).First(&pc)
	if result.Error != nil {
		pc = models.PaymentConfig{
			PaymentType: req.PaymentType,
		}
	}

	pc.Enabled = req.Enabled
	pc.Config = string(req.Config)

	if result.Error != nil {
		config.DB.Create(&pc)
	} else {
		config.DB.Save(&pc)
	}

	utils.SuccessMsg(c, "保存成功")
}

func ListOrders(c *gin.Context) {
	page, pageSize := utils.PageQuery(c)
	var total int64
	var orders []models.Order

	query := config.DB.Model(&models.Order{}).Preload("Program").Preload("License")

	if orderNo := c.Query("order_no"); orderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+orderNo+"%")
	}
	if method := c.Query("payment_method"); method != "" {
		query = query.Where("payment_method = ?", method)
	}
	if status := c.Query("payment_status"); status != "" {
		query = query.Where("payment_status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders)

	var totalAmount float64
	config.DB.Model(&models.Order{}).Where("payment_status = 1").Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)

	utils.Success(c, gin.H{
		"list":         orders,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_amount": totalAmount,
	})
}

func PaymentCallback(c *gin.Context) {
	paymentType := c.Param("type")
	action := c.Param("action")

	_ = paymentType
	_ = action

	utils.SuccessMsg(c, "ok")
}

func GetOrderStats(c *gin.Context) {
	var todayAmount, weekAmount, monthAmount, totalAmount float64

	config.DB.Model(&models.Order{}).
		Where("payment_status = 1 AND DATE(paid_at) = CURDATE()").
		Select("COALESCE(SUM(amount), 0)").Scan(&todayAmount)

	config.DB.Model(&models.Order{}).
		Where("payment_status = 1 AND paid_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").
		Select("COALESCE(SUM(amount), 0)").Scan(&weekAmount)

	config.DB.Model(&models.Order{}).
		Where("payment_status = 1 AND paid_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)").
		Select("COALESCE(SUM(amount), 0)").Scan(&monthAmount)

	config.DB.Model(&models.Order{}).
		Where("payment_status = 1").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)

	utils.Success(c, gin.H{
		"today": todayAmount,
		"week":  weekAmount,
		"month": monthAmount,
		"total": totalAmount,
	})
}
