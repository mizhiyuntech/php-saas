package controllers

import (
	"yuyue-auth/config"
	"yuyue-auth/middleware"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	var admin models.Admin
	if err := config.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		utils.Error(c, 401, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		utils.Error(c, 401, "用户名或密码错误")
		return
	}

	token, err := middleware.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		utils.ErrorServer(c, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"token":    token,
		"username": admin.Username,
	})
}

func GetProfile(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	username, _ := c.Get("username")

	utils.Success(c, gin.H{
		"id":       adminID,
		"username": username,
	})
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	var admin models.Admin
	if err := config.DB.First(&admin, adminID).Error; err != nil {
		utils.ErrorServer(c, "用户不存在")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		utils.Error(c, 400, "原密码错误")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorServer(c, "密码加密失败")
		return
	}

	config.DB.Model(&admin).Update("password", string(hashedPassword))
	utils.SuccessMsg(c, "密码修改成功")
}
