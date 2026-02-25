package controllers

import (
	"encoding/json"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	FromName string `json:"from_name"`
	SSL      bool   `json:"ssl"`
}

func GetSMTPConfig(c *gin.Context) {
	val := models.GetSetting("smtp_config")
	if val == "" {
		utils.Success(c, SMTPConfig{Port: 465, SSL: true})
		return
	}

	var cfg SMTPConfig
	json.Unmarshal([]byte(val), &cfg)
	utils.Success(c, cfg)
}

func UpdateSMTPConfig(c *gin.Context) {
	var cfg SMTPConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	data, _ := json.Marshal(cfg)

	var setting models.Setting
	result := config.DB.Where("`key` = ?", "smtp_config").First(&setting)
	if result.Error != nil {
		setting = models.Setting{Group: "system", Key: "smtp_config", Value: string(data)}
		config.DB.Create(&setting)
	} else {
		config.DB.Model(&setting).Update("value", string(data))
	}

	utils.SuccessMsg(c, "保存成功")
}

func TestSMTPConfig(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "请输入收件邮箱")
		return
	}

	val := models.GetSetting("smtp_config")
	if val == "" {
		utils.Error(c, 400, "请先配置SMTP")
		return
	}

	var cfg SMTPConfig
	json.Unmarshal([]byte(val), &cfg)

	siteTitle := models.GetSetting("site_title")
	if siteTitle == "" {
		siteTitle = "鱼跃授权"
	}

	err := utils.SendMail(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.FromName, cfg.SSL,
		req.To, siteTitle+" - 测试邮件", "这是一封测试邮件，如果您收到此邮件说明SMTP配置正确。")

	if err != nil {
		utils.Error(c, 500, "发送失败: "+err.Error())
		return
	}

	utils.SuccessMsg(c, "发送成功")
}
