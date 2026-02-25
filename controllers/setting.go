package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type SettingsUpdateRequest struct {
	SiteTitle       string `json:"site_title"`
	SiteDescription string `json:"site_description"`
	SiteKeywords    string `json:"site_keywords"`
	FooterCopyright string `json:"footer_copyright"`
	PoliceRecord    string `json:"police_record"`
	SiteURL         string `json:"site_url"`
}

func GetSettings(c *gin.Context) {
	var settings []models.Setting
	config.DB.Where("`group` = ?", "system").Find(&settings)

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	utils.Success(c, result)
}

func UpdateSettings(c *gin.Context) {
	var req SettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	settingsMap := map[string]string{
		"site_title":       req.SiteTitle,
		"site_description": req.SiteDescription,
		"site_keywords":    req.SiteKeywords,
		"footer_copyright": req.FooterCopyright,
		"police_record":    req.PoliceRecord,
		"site_url":         req.SiteURL,
	}

	for key, value := range settingsMap {
		var setting models.Setting
		result := config.DB.Where("`key` = ?", key).First(&setting)
		if result.Error != nil {
			setting = models.Setting{
				Group: "system",
				Key:   key,
				Value: value,
			}
			config.DB.Create(&setting)
		} else {
			config.DB.Model(&setting).Update("value", value)
		}
	}

	utils.SuccessMsg(c, "保存成功")
}

func UploadIcon(c *gin.Context) {
	uploadFile(c, "site_icon", "icon")
}

func UploadFavicon(c *gin.Context) {
	uploadFile(c, "site_favicon", "favicon")
}

func uploadFile(c *gin.Context, settingKey, fileType string) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorBad(c, "请选择文件")
		return
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		utils.ErrorBad(c, "文件大小不能超过2MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".ico": true, ".webp": true}
	if !allowedExts[ext] {
		utils.ErrorBad(c, "仅支持 PNG/JPG/ICO/WEBP 格式")
		return
	}

	uploadDir := "uploads/icons"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("%s_%s%s", fileType, utils.GenerateRandomString(8), ext)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(header, filePath); err != nil {
		utils.ErrorServer(c, "保存文件失败")
		return
	}

	urlPath := "/" + filePath
	models.SetSetting(settingKey, urlPath)

	utils.Success(c, gin.H{
		"url": urlPath,
	})
}

func GetSiteInfo(c *gin.Context) {
	var settings []models.Setting
	config.DB.Where("`group` = ?", "system").Find(&settings)

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	utils.Success(c, result)
}

func GetUnauthPageHTML(c *gin.Context) {
	html := models.GetSetting("unauth_page_html")
	utils.Success(c, gin.H{"html": html})
}

func UpdateUnauthPageHTML(c *gin.Context) {
	var req struct {
		HTML string `json:"html"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	var setting models.Setting
	result := config.DB.Where("`key` = ?", "unauth_page_html").First(&setting)
	if result.Error != nil {
		setting = models.Setting{Group: "system", Key: "unauth_page_html", Value: req.HTML}
		config.DB.Create(&setting)
	} else {
		config.DB.Model(&setting).Update("value", req.HTML)
	}

	utils.SuccessMsg(c, "保存成功")
}

func GetDashboardStats(c *gin.Context) {
	var programCount, licenseCount, activeLicenseCount, orderCount int64

	config.DB.Model(&models.Program{}).Count(&programCount)
	config.DB.Model(&models.License{}).Count(&licenseCount)
	config.DB.Model(&models.License{}).Where("status = 1").Count(&activeLicenseCount)
	config.DB.Model(&models.Order{}).Count(&orderCount)

	var totalIncome float64
	config.DB.Model(&models.Order{}).Where("payment_status = 1").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)

	utils.Success(c, gin.H{
		"program_count":        programCount,
		"license_count":        licenseCount,
		"active_license_count": activeLicenseCount,
		"order_count":          orderCount,
		"total_income":         totalIncome,
	})
}
