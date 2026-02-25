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
	if _, err := os.Stat(frontendDist); err == nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			if strings.HasPrefix(path, "/api/") {
				c.JSON(404, gin.H{"code": 404, "message": "Not found"})
				return
			}

			filePath := filepath.Join(frontendDist, path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}

			c.File(filepath.Join(frontendDist, "index.html"))
		})
	} else {
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, gin.H{"code": 404, "message": "Not found"})
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fallbackPage()))
		})
	}

	port := 3132
	if config.Conf != nil && config.Conf.Port > 0 {
		port = config.Conf.Port
	}

	log.Printf("YuYue Auth System started on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}

func fallbackPage() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>YuYue Auth</title>
<style>
body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f5f5f5}
.c{text-align:center;padding:40px;background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,.1)}
h1{color:#333;margin-bottom:16px}
p{color:#666}
code{background:#f0f0f0;padding:2px 8px;border-radius:4px}
</style>
</head>
<body>
<div class="c">
<h1>YuYue Auth System</h1>
<p>Frontend not built. Run: cd frontend && npm install && npm run generate</p>
</div>
</body>
</html>`
}
