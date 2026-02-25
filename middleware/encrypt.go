package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"yuyue-auth/models"
	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

func APIEncrypt() gin.HandlerFunc {
	return func(c *gin.Context) {
		encryptMode := models.GetSetting("api_encrypt_mode")
		encryptKey := models.GetSetting("api_encrypt_key")

		if encryptMode == "" || encryptMode == "none" || encryptKey == "" {
			c.Next()
			return
		}

		isCoreAPI := isCoreEndpoint(c.Request.URL.Path)

		if encryptMode == "partial" && !isCoreAPI {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "application/encrypted" || c.GetHeader("X-Encrypted") == "1" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "读取请求失败"})
				c.Abort()
				return
			}

			var encReq struct {
				Data string `json:"data"`
			}
			if err := json.Unmarshal(body, &encReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求格式错误"})
				c.Abort()
				return
			}

			plaintext, err := utils.AES256Decrypt(encReq.Data, encryptKey)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "解密失败"})
				c.Abort()
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBufferString(plaintext))
			c.Request.ContentLength = int64(len(plaintext))
			c.Request.Header.Set("Content-Type", "application/json")
		}

		writer := &encryptResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}, key: encryptKey}
		c.Writer = writer

		c.Next()

		if writer.body.Len() > 0 {
			encrypted, err := utils.AES256Encrypt(writer.body.String(), encryptKey)
			if err != nil {
				c.Writer = writer.ResponseWriter
				c.JSON(500, gin.H{"code": 500, "message": "加密响应失败"})
				return
			}

			resp := gin.H{"encrypted": true, "data": encrypted}
			respBytes, _ := json.Marshal(resp)

			writer.ResponseWriter.Header().Set("Content-Type", "application/json")
			writer.ResponseWriter.Header().Set("X-Encrypted", "1")
			writer.ResponseWriter.WriteHeader(http.StatusOK)
			writer.ResponseWriter.Write(respBytes)
		}
	}
}

func isCoreEndpoint(path string) bool {
	coreEndpoints := []string{
		"/api/license/verify",
		"/api/piracy/check",
		"/api/auth/login",
	}
	for _, ep := range coreEndpoints {
		if path == ep {
			return true
		}
	}
	return false
}

type encryptResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	key  string
}

func (w *encryptResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *encryptResponseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}
