package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"yuyue-auth/config"
	"yuyue-auth/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.IsInstalled() {
		if err := config.InitDatabase(); err != nil {
			log.Fatalf("Failed to init database: %v", err)
		}
		if err := config.InitRedis(); err != nil {
			log.Fatalf("Failed to init redis: %v", err)
		}
		log.Println("Database and Redis connected")
	} else {
		log.Println("System not installed, visit /install to setup")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.MaxMultipartMemory = 32 << 20

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	routes.Setup(r)

	r.Static("/uploads", "./uploads")

	frontendDist := "./frontend/.output/public"
	r.StaticFS("/_nuxt", http.Dir(filepath.Join(frontendDist, "_nuxt")))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/") {
			c.JSON(404, gin.H{"code": 404, "message": "Not found"})
			return
		}

		filePath := filepath.Join(frontendDist, path)
		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {
			c.File(filePath)
			return
		}

		htmlPath := filepath.Join(frontendDist, path, "index.html")
		if _, err := os.Stat(htmlPath); err == nil {
			c.File(htmlPath)
			return
		}

		c.File(filepath.Join(frontendDist, "200.html"))
	})

	port := 3132
	if config.Conf != nil && config.Conf.Port > 0 {
		port = config.Conf.Port
	}

	log.Printf("YuYue Auth System started on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
