package controllers

import (
	"strconv"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

type PackageRequest struct {
	ProgramID   uint    `json:"program_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	Price       float64 `json:"price" binding:"required"`
	SortOrder   int     `json:"sort_order"`
	Status      *int    `json:"status"`
}

func ListPackages(c *gin.Context) {
	page, pageSize := utils.PageQuery(c)
	var total int64
	var packages []models.Package

	query := config.DB.Model(&models.Package{}).Preload("Program")

	if programID := c.Query("program_id"); programID != "" {
		query = query.Where("program_id = ?", programID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	query.Order("sort_order ASC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&packages)

	utils.Success(c, gin.H{
		"list":      packages,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func CreatePackage(c *gin.Context) {
	var req PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误: "+err.Error())
		return
	}

	pkg := models.Package{
		ProgramID:   req.ProgramID,
		Name:        req.Name,
		Description: req.Description,
		Duration:    req.Duration,
		Price:       req.Price,
		SortOrder:   req.SortOrder,
		Status:      1,
	}

	if err := config.DB.Create(&pkg).Error; err != nil {
		utils.ErrorServer(c, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, pkg)
}

func UpdatePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	var pkg models.Package
	if err := config.DB.First(&pkg, id).Error; err != nil {
		utils.Error(c, 404, "套餐不存在")
		return
	}

	var req PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"program_id":  req.ProgramID,
		"name":        req.Name,
		"description": req.Description,
		"duration":    req.Duration,
		"price":       req.Price,
		"sort_order":  req.SortOrder,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	config.DB.Model(&pkg).Updates(updates)
	utils.Success(c, pkg)
}

func DeletePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBad(c, "无效的ID")
		return
	}

	if err := config.DB.Delete(&models.Package{}, id).Error; err != nil {
		utils.ErrorServer(c, "删除失败")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}
