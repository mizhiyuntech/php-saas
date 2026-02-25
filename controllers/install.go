package controllers

import (
	"context"
	"fmt"
	"time"

	"yuyue-auth/config"
	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type InstallRequest struct {
	MySQLHost     string `json:"mysql_host" binding:"required"`
	MySQLPort     int    `json:"mysql_port" binding:"required"`
	MySQLUser     string `json:"mysql_user" binding:"required"`
	MySQLPassword string `json:"mysql_password"`
	MySQLDatabase string `json:"mysql_database" binding:"required"`
	RedisHost     string `json:"redis_host" binding:"required"`
	RedisPort     int    `json:"redis_port" binding:"required"`
	RedisPassword string `json:"redis_password"`
	AdminUsername string `json:"admin_username" binding:"required"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
	SiteTitle     string `json:"site_title"`
}

func GetInstallStatus(c *gin.Context) {
	utils.Success(c, gin.H{
		"installed": config.IsInstalled(),
	})
}

func Install(c *gin.Context) {
	if config.IsInstalled() {
		utils.Error(c, 400, "系统已安装，请勿重复安装")
		return
	}

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBad(c, "参数错误: "+err.Error())
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		req.MySQLUser, req.MySQLPassword, req.MySQLHost, req.MySQLPort)

	testDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.Error(c, 500, "数据库连接失败: "+err.Error())
		return
	}

	testDB.Exec("CREATE DATABASE IF NOT EXISTS `" + req.MySQLDatabase + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")

	sqlDB, _ := testDB.DB()
	sqlDB.Close()

	cfg := &config.AppConfig{
		Installed: true,
		Port:      3132,
		JWTSecret: utils.GenerateRandomString(32),
		MySQL: config.MySQLConfig{
			Host:     req.MySQLHost,
			Port:     req.MySQLPort,
			User:     req.MySQLUser,
			Password: req.MySQLPassword,
			Database: req.MySQLDatabase,
		},
		Redis: config.RedisConfig{
			Host:     req.RedisHost,
			Port:     req.RedisPort,
			Password: req.RedisPassword,
			DB:       0,
		},
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.ErrorServer(c, "保存配置失败: "+err.Error())
		return
	}

	if err := config.InitDatabase(); err != nil {
		utils.ErrorServer(c, "初始化数据库失败: "+err.Error())
		return
	}

	if err := config.InitRedis(); err != nil {
		utils.ErrorServer(c, "初始化Redis失败: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := config.RDB.Ping(ctx).Err(); err != nil {
		utils.ErrorServer(c, "Redis连接失败: "+err.Error())
		return
	}

	if err := models.AutoMigrate(); err != nil {
		utils.ErrorServer(c, "数据库迁移失败: "+err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorServer(c, "密码加密失败")
		return
	}

	admin := models.Admin{
		Username: req.AdminUsername,
		Password: string(hashedPassword),
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		utils.ErrorServer(c, "创建管理员失败: "+err.Error())
		return
	}

	models.InitDefaultSettings()

	if req.SiteTitle != "" {
		models.SetSetting("site_title", req.SiteTitle)
	}

	utils.SuccessMsg(c, "安装成功")
}
