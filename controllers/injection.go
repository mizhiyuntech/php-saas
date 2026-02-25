package controllers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"yuyue-auth/utils"

	"github.com/gin-gonic/gin"
)

var supportedLanguages = []gin.H{
	{"value": "php", "label": "PHP"},
	{"value": "python", "label": "Python"},
	{"value": "nodejs", "label": "Node.js"},
	{"value": "go", "label": "Go"},
	{"value": "java", "label": "Java"},
	{"value": "csharp", "label": "C#"},
}

func GetLanguages(c *gin.Context) {
	utils.Success(c, supportedLanguages)
}

func InjectAuthorization(c *gin.Context) {
	language := c.PostForm("language")
	if language == "" {
		utils.ErrorBad(c, "请选择开发语言")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorBad(c, "请上传ZIP文件")
		return
	}
	defer file.Close()

	if header.Size > 30*1024*1024 {
		utils.ErrorBad(c, "文件大小不能超过30MB")
		return
	}

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		utils.ErrorBad(c, "仅支持ZIP格式文件")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorServer(c, "读取文件失败")
		return
	}

	reader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		utils.ErrorServer(c, "解析ZIP文件失败")
		return
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, f := range reader.File {
		w, err := writer.Create(f.Name)
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		io.Copy(w, rc)
		rc.Close()
	}

	authFiles := generateAuthFiles(language)
	for name, content := range authFiles {
		w, err := writer.Create(filepath.Join("_auth", name))
		if err != nil {
			continue
		}
		w.Write([]byte(content))
	}

	writer.Close()

	outputName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)) + "_authorized.zip"

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", outputName))
	c.Data(200, "application/zip", buf.Bytes())
}

func generateAuthFiles(language string) map[string]string {
	files := make(map[string]string)

	switch language {
	case "php":
		files["check_license.php"] = `<?php
/**
 * 鱼跃授权 - 授权验证模块
 * 将此文件引入到您的项目入口文件中
 */

function yuyue_check_license($license_key, $api_url, $program_id) {
    $data = json_encode([
        'license_key' => $license_key,
        'program_id'  => (int)$program_id,
        'device_info' => php_uname()
    ]);

    $ch = curl_init($api_url . '/api/license/verify');
    curl_setopt_array($ch, [
        CURLOPT_POST           => true,
        CURLOPT_POSTFIELDS     => $data,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_HTTPHEADER     => ['Content-Type: application/json'],
        CURLOPT_TIMEOUT        => 10,
    ]);

    $response = curl_exec($ch);
    curl_close($ch);

    if (!$response) {
        return false;
    }

    $result = json_decode($response, true);
    return isset($result['code']) && $result['code'] === 0;
}

// 使用示例:
// $authorized = yuyue_check_license('YOUR-LICENSE-KEY', 'https://your-domain.com', 1);
// if (!$authorized) { die('授权验证失败'); }
`
		files["config.php"] = `<?php
// 鱼跃授权配置文件
return [
    'api_url'     => 'https://your-domain.com',
    'program_id'  => 1,
    'license_key' => 'YOUR-LICENSE-KEY',
];
`

	case "python":
		files["check_license.py"] = `"""
鱼跃授权 - 授权验证模块
将此模块导入到您的项目中使用
"""
import json
import platform
import urllib.request


def check_license(license_key: str, api_url: str, program_id: int) -> bool:
    data = json.dumps({
        "license_key": license_key,
        "program_id": program_id,
        "device_info": platform.node()
    }).encode("utf-8")

    req = urllib.request.Request(
        f"{api_url}/api/license/verify",
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST"
    )

    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            result = json.loads(resp.read().decode())
            return result.get("code") == 0
    except Exception:
        return False


# 使用示例:
# if not check_license("YOUR-LICENSE-KEY", "https://your-domain.com", 1):
#     raise SystemExit("授权验证失败")
`
		files["config.py"] = `# 鱼跃授权配置文件
AUTH_CONFIG = {
    "api_url": "https://your-domain.com",
    "program_id": 1,
    "license_key": "YOUR-LICENSE-KEY",
}
`

	case "nodejs":
		files["check_license.js"] = `/**
 * 鱼跃授权 - 授权验证模块
 * 在项目中引入此模块
 */
const https = require('https');
const http = require('http');
const os = require('os');

function checkLicense(licenseKey, apiUrl, programId) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify({
      license_key: licenseKey,
      program_id: programId,
      device_info: os.hostname()
    });

    const url = new URL(apiUrl + '/api/license/verify');
    const client = url.protocol === 'https:' ? https : http;

    const req = client.request({
      hostname: url.hostname,
      port: url.port,
      path: url.pathname,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      },
      timeout: 10000
    }, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          const result = JSON.parse(body);
          resolve(result.code === 0);
        } catch (e) {
          resolve(false);
        }
      });
    });

    req.on('error', () => resolve(false));
    req.write(data);
    req.end();
  });
}

module.exports = { checkLicense };

// 使用示例:
// const { checkLicense } = require('./_auth/check_license');
// const authorized = await checkLicense('YOUR-LICENSE-KEY', 'https://your-domain.com', 1);
`
		files["config.js"] = `// 鱼跃授权配置文件
module.exports = {
  apiUrl: 'https://your-domain.com',
  programId: 1,
  licenseKey: 'YOUR-LICENSE-KEY',
};
`

	case "go":
		files["check_license.go"] = `package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// CheckLicense 验证授权 - 鱼跃授权验证模块
func CheckLicense(licenseKey, apiURL string, programID int) error {
	hostname, _ := os.Hostname()
	payload := map[string]interface{}{
		"license_key": licenseKey,
		"program_id":  programID,
		"device_info": hostname,
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(apiURL+"/api/license/verify", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("授权验证请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("授权验证响应解析失败: %w", err)
	}

	if code, ok := result["code"].(float64); !ok || code != 0 {
		msg, _ := result["message"].(string)
		return fmt.Errorf("授权验证失败: %s", msg)
	}

	return nil
}
`
		files["config.go"] = `package auth

// 鱼跃授权配置
var (
	APIUrl     = "https://your-domain.com"
	ProgramID  = 1
	LicenseKey = "YOUR-LICENSE-KEY"
)
`

	case "java":
		files["CheckLicense.java"] = `package auth;

import java.io.*;
import java.net.*;
import java.nio.charset.StandardCharsets;

/**
 * 鱼跃授权 - 授权验证模块
 */
public class CheckLicense {

    public static boolean verify(String licenseKey, String apiUrl, int programId) {
        try {
            String hostname = InetAddress.getLocalHost().getHostName();
            String jsonPayload = String.format(
                "{\"license_key\":\"%s\",\"program_id\":%d,\"device_info\":\"%s\"}",
                licenseKey, programId, hostname
            );

            URL url = new URL(apiUrl + "/api/license/verify");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Content-Type", "application/json");
            conn.setDoOutput(true);
            conn.setConnectTimeout(10000);
            conn.setReadTimeout(10000);

            try (OutputStream os = conn.getOutputStream()) {
                os.write(jsonPayload.getBytes(StandardCharsets.UTF_8));
            }

            try (BufferedReader br = new BufferedReader(
                    new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8))) {
                StringBuilder response = new StringBuilder();
                String line;
                while ((line = br.readLine()) != null) {
                    response.append(line);
                }
                return response.toString().contains("\"code\":0");
            }
        } catch (Exception e) {
            return false;
        }
    }
}
`

	case "csharp":
		files["CheckLicense.cs"] = `using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Auth
{
    /// <summary>
    /// 鱼跃授权 - 授权验证模块
    /// </summary>
    public class CheckLicense
    {
        private static readonly HttpClient _client = new HttpClient { Timeout = TimeSpan.FromSeconds(10) };

        public static async Task<bool> VerifyAsync(string licenseKey, string apiUrl, int programId)
        {
            try
            {
                var payload = new
                {
                    license_key = licenseKey,
                    program_id = programId,
                    device_info = Environment.MachineName
                };

                var json = JsonSerializer.Serialize(payload);
                var content = new StringContent(json, Encoding.UTF8, "application/json");
                var response = await _client.PostAsync($"{apiUrl}/api/license/verify", content);
                var result = await response.Content.ReadAsStringAsync();

                using var doc = JsonDocument.Parse(result);
                return doc.RootElement.GetProperty("code").GetInt32() == 0;
            }
            catch
            {
                return false;
            }
        }
    }
}
`
	}

	files["README.md"] = `# 鱼跃授权 - 授权验证模块

## 使用说明

1. 将 _auth 目录复制到您的项目中
2. 修改配置文件中的 api_url、program_id 和 license_key
3. 在项目入口处调用授权验证函数
4. 验证通过后程序正常运行，否则将提示授权失败

## 配置项

- api_url: 授权系统API地址
- program_id: 程序ID（在授权系统中创建程序后获取）
- license_key: 授权码（产品序列码）
`

	return files
}
