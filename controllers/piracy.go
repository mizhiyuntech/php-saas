package controllers

import (
	"strconv"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type PiracyRequest struct {
	Domain  string `json:"domain" binding:"required"`
	IP      string `json:"ip"`
	Message string `json:"message" binding:"required"`
	Status  *int   `json:"status"`
}

func ListPiracy(c *gin.Context) {
	page, pageSize := utils.PageQuery(c)
	var total int64
	var records []models.PiracyRecord

	query := config.DB.Model(&models.PiracyRecord{})

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("domain LIKE ? OR ip LIKE ? OR message LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)

	utils.Success(c, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func CreatePiracy(c *gin.Context) {
	var req PiracyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误: "+err.Error())
		return
	}

	record := models.PiracyRecord{
		Domain:  req.Domain,
		IP:      req.IP,
		Message: req.Message,
		Status:  1,
	}

	if err := config.DB.Create(&record).Error; err != nil {
		utils.ErrorServer(c, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, record)
}

func UpdatePiracy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	var record models.PiracyRecord
	if err := config.DB.First(&record, id).Error; err != nil {
		utils.Error(c, 404, "记录不存在")
		return
	}

	var req PiracyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"domain":  req.Domain,
		"ip":      req.IP,
		"message": req.Message,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	config.DB.Model(&record).Updates(updates)
	utils.Success(c, record)
}

func DeletePiracy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	if err := config.DB.Delete(&models.PiracyRecord{}, id).Error; err != nil {
		utils.ErrorServer(c, "删除失败")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}

func CheckPiracy(c *gin.Context) {
	type CheckRequest struct {
		Domain string `json:"domain"`
		IP     string `json:"ip"`
	}

	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Success(c, gin.H{"pirated": false})
		return
	}

	var record models.PiracyRecord
	query := config.DB.Where("status = 1")

	if req.Domain != "" && req.IP != "" {
		query = query.Where("domain = ? OR ip = ?", req.Domain, req.IP)
	} else if req.Domain != "" {
		query = query.Where("domain = ?", req.Domain)
	} else if req.IP != "" {
		query = query.Where("ip = ?", req.IP)
	} else {
		utils.Success(c, gin.H{"pirated": false})
		return
	}

	if err := query.First(&record).Error; err != nil {
		utils.Success(c, gin.H{"pirated": false})
		return
	}

	piracyHTML := models.GetSetting("piracy_page_html")
	if piracyHTML == "" {
		piracyHTML = defaultPiracyHTML()
	}

	utils.Success(c, gin.H{
		"pirated": true,
		"message": record.Message,
		"html":    piracyHTML,
	})
}

func GetPiracyPageHTML(c *gin.Context) {
	html := models.GetSetting("piracy_page_html")
	utils.Success(c, gin.H{"html": html})
}

func UpdatePiracyPageHTML(c *gin.Context) {
	var req struct {
		HTML string `json:"html"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	var setting models.Setting
	result := config.DB.Where("`key` = ?", "piracy_page_html").First(&setting)
	if result.Error != nil {
		setting = models.Setting{Group: "system", Key: "piracy_page_html", Value: req.HTML}
		config.DB.Create(&setting)
	} else {
		config.DB.Model(&setting).Update("value", req.HTML)
	}

	utils.SuccessMsg(c, "保存成功")
}

func defaultPiracyHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>警告</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#fef2f2}
.box{width:100%;max-width:480px;padding:40px;background:#fff;border-radius:8px;border:1px solid #fecaca;text-align:center}
h2{font-size:22px;font-weight:700;color:#dc2626;margin-bottom:12px}
.msg{font-size:15px;color:#991b1b;line-height:1.6;margin-bottom:20px;padding:16px;background:#fef2f2;border-radius:6px}
.sub{font-size:13px;color:#9ca3af}
</style>
</head>
<body>
<div class="box">
<h2>盗版警告</h2>
<div class="msg">{{MESSAGE}}</div>
<p class="sub">此程序为未授权的盗版副本，请立即停止使用并联系正版授权方</p>
</div>
</body>
</html>`
}
