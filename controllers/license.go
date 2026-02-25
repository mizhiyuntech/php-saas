package controllers

import (
	"strconv"
	"time"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type CreateLicenseRequest struct {
	ProgramID uint   `json:"program_id" binding:"required"`
	Count     int    `json:"count" binding:"required,min=1,max=100"`
	Duration  int    `json:"duration"`
	Remark    string `json:"remark"`
}

type UpdateLicenseRequest struct {
	Status   *int   `json:"status"`
	Duration int    `json:"duration"`
	Remark   string `json:"remark"`
}

func ListLicenses(c *gin.Context) {
	page, pageSize := utils.PageQuery(c)
	var total int64
	var licenses []models.License

	query := config.DB.Model(&models.License{}).Preload("Program")

	if programID := c.Query("program_id"); programID != "" {
		query = query.Where("program_id = ?", programID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("license_key LIKE ? OR remark LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&licenses)

	utils.Success(c, gin.H{
		"list":      licenses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func CreateLicenses(c *gin.Context) {
	var req CreateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误: "+err.Error())
		return
	}

	var program models.Program
	if err := config.DB.First(&program, req.ProgramID).Error; err != nil {
		utils.Error(c, 404, "程序不存在")
		return
	}

	var licenses []models.License
	for i := 0; i < req.Count; i++ {
		license := models.License{
			ProgramID:  req.ProgramID,
			LicenseKey: utils.GenerateLicenseKey(),
			Status:     0,
			Duration:   req.Duration,
			Remark:     req.Remark,
		}
		licenses = append(licenses, license)
	}

	if err := config.DB.Create(&licenses).Error; err != nil {
		utils.ErrorServer(c, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"count":    len(licenses),
		"licenses": licenses,
	})
}

func UpdateLicense(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	var license models.License
	if err := config.DB.First(&license, id).Error; err != nil {
		utils.Error(c, 404, "授权码不存在")
		return
	}

	var req UpdateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Duration > 0 {
		updates["duration"] = req.Duration
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	config.DB.Model(&license).Updates(updates)
	utils.Success(c, license)
}

func DeleteLicense(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	if err := config.DB.Delete(&models.License{}, id).Error; err != nil {
		utils.ErrorServer(c, "删除失败")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}

func VerifyLicense(c *gin.Context) {
	type VerifyRequest struct {
		LicenseKey string `json:"license_key" binding:"required"`
		ProgramID  uint   `json:"program_id" binding:"required"`
		DeviceInfo string `json:"device_info"`
		Timestamp  int64  `json:"_ts"`
		Signature  string `json:"_sg"`
		TokenHash  string `json:"_th"`
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	nowMs := time.Now().UnixMilli()
	if req.Timestamp > 0 {
		diff := nowMs - req.Timestamp
		if diff < 0 {
			diff = -diff
		}
		if diff > 300000 {
			utils.Error(c, 403, "请求已过期")
			return
		}
	}

	var license models.License
	if err := config.DB.Where("license_key = ? AND program_id = ?", req.LicenseKey, req.ProgramID).First(&license).Error; err != nil {
		utils.Error(c, 404, "授权码无效")
		return
	}

	if license.Status == 3 {
		utils.Error(c, 403, "授权码已禁用")
		return
	}

	if license.Status == 2 {
		utils.Error(c, 403, "授权码已过期")
		return
	}

	if license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now()) {
		config.DB.Model(&license).Update("status", 2)
		utils.Error(c, 403, "授权码已过期")
		return
	}

	if license.Status == 0 {
		now := time.Now()
		updates := map[string]interface{}{
			"status":       1,
			"activated_at": now,
			"device_info":  req.DeviceInfo,
		}
		if license.Duration > 0 {
			expiresAt := now.AddDate(0, 0, license.Duration)
			updates["expires_at"] = expiresAt
		}
		config.DB.Model(&license).Updates(updates)
		config.DB.First(&license, license.ID)
	}

	serverTs := time.Now().UnixMilli()
	responseToken := utils.GenerateRandomString(16)

	utils.Success(c, gin.H{
		"valid":       true,
		"license_key": license.LicenseKey,
		"expires_at":  license.ExpiresAt,
		"status":      license.Status,
		"server_ts":   serverTs,
		"token":       responseToken,
	})
}
