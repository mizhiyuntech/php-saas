package controllers

import (
	"strconv"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type ProgramRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Language    string `json:"language"`
	Status      *int   `json:"status"`
}

func ListPrograms(c *gin.Context) {
	page, pageSize := utils.PageQuery(c)
	var total int64
	var programs []models.Program

	query := config.DB.Model(&models.Program{})

	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&programs)

	utils.Success(c, gin.H{
		"list":      programs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	var program models.Program
	if err := config.DB.First(&program, id).Error; err != nil {
		utils.Error(c, 404, "程序不存在")
		return
	}

	utils.Success(c, program)
}

func CreateProgram(c *gin.Context) {
	var req ProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误: "+err.Error())
		return
	}

	program := models.Program{
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		Language:    req.Language,
		SecretKey:   utils.GenerateRandomString(32),
		Status:      1,
	}

	if err := config.DB.Create(&program).Error; err != nil {
		utils.ErrorServer(c, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, program)
}

func UpdateProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	var program models.Program
	if err := config.DB.First(&program, id).Error; err != nil {
		utils.Error(c, 404, "程序不存在")
		return
	}

	var req ProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"version":     req.Version,
		"language":    req.Language,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	config.DB.Model(&program).Updates(updates)
	utils.Success(c, program)
}

func DeleteProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	if err := config.DB.Delete(&models.Program{}, id).Error; err != nil {
		utils.ErrorServer(c, "删除失败")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}

func GetAllPrograms(c *gin.Context) {
	var programs []models.Program
	config.DB.Where("status = ?", 1).Select("id, name, language").Find(&programs)
	utils.Success(c, programs)
}
