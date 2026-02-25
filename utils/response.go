package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": msg,
	})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})
}

func ErrorBad(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    400,
		"message": msg,
	})
}

func ErrorUnauth(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": msg,
	})
}

func ErrorServer(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    500,
		"message": msg,
	})
}

func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func GenerateLicenseKey() string {
	seg := make([]string, 4)
	for i := range seg {
		b := make([]byte, 4)
		rand.Read(b)
		seg[i] = strings.ToUpper(hex.EncodeToString(b))
	}
	return strings.Join(seg, "-")
}

func GenerateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("YY%s%s",
		now.Format("20060102150405"),
		GenerateRandomString(6),
	)
}

func PageQuery(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return page, pageSize
}
