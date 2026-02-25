package controllers

import (
	"archive/zip"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"yuyue-auth/models"
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

func xorEncode(data, key string) string {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return base64.StdEncoding.EncodeToString(result)
}

func fileHash(content string) string {
	h := md5.Sum([]byte(content))
	return hex.EncodeToString(h[:])
}

var sdkDir = "lib/.cache"

var entryFiles = map[string][]string{
	"php":    {"index.php", "app.php", "public/index.php", "main.php"},
	"python": {"main.py", "app.py", "manage.py", "run.py", "wsgi.py"},
	"nodejs": {"index.js", "app.js", "server.js", "main.js", "src/index.js", "src/app.js"},
	"go":     {"main.go", "cmd/main.go", "cmd/server/main.go"},
	"java":   {"src/Main.java", "src/App.java", "src/main/java/Main.java", "src/main/java/App.java"},
	"csharp": {"Program.cs", "Startup.cs", "src/Program.cs"},
}

func sdkFileName(lang string) string {
	switch lang {
	case "php":
		return "bootstrap.php"
	case "python":
		return "bootstrap.py"
	case "nodejs":
		return "bootstrap.js"
	case "go":
		return "bootstrap.go"
	case "java":
		return "Bootstrap.java"
	case "csharp":
		return "Bootstrap.cs"
	}
	return "bootstrap"
}

func entryImportLine(lang, sdkPath string) string {
	switch lang {
	case "php":
		return fmt.Sprintf("<?php require_once __DIR__.'/%s/%s'; ?>", sdkPath, sdkFileName(lang))
	case "python":
		mod := strings.ReplaceAll(sdkPath, "/", ".")
		return fmt.Sprintf("from %s.bootstrap import *\n", mod)
	case "nodejs":
		return fmt.Sprintf("require('./%s/%s');\n", sdkPath, sdkFileName(lang))
	}
	return ""
}

func injectEntryFile(lang, originalContent, sdkPath string) string {
	switch lang {
	case "php":
		importLine := fmt.Sprintf("require_once __DIR__.'/%s/%s';", sdkPath, sdkFileName(lang))
		if strings.HasPrefix(strings.TrimSpace(originalContent), "<?php") {
			return strings.Replace(originalContent, "<?php", "<?php\n"+importLine, 1)
		}
		return "<?php\n" + importLine + "\n?>" + originalContent
	case "python":
		importLine := fmt.Sprintf("from %s.bootstrap import *", strings.ReplaceAll(sdkPath, "/", "."))
		return importLine + "\n" + originalContent
	case "nodejs":
		importLine := fmt.Sprintf("require('./%s/%s');", sdkPath, sdkFileName(lang))
		return importLine + "\n" + originalContent
	}
	return originalContent
}

func generateLicFile(licenseKey, programID, serverURL, secKey string) string {
	xe := func(d, k string) string {
		o := make([]byte, len(d))
		for i := 0; i < len(d); i++ {
			o[i] = d[i] ^ k[i%len(k)]
		}
		return base64.StdEncoding.EncodeToString(o)
	}
	hm := func(d, k string) string {
		h := hmacSHA256(d, k)
		return h
	}

	e1 := xe(licenseKey, secKey)
	e2 := xe(programID, secKey)
	e3 := xe(serverURL, secKey)
	chk := hm(licenseKey+programID+serverURL, secKey)

	return e1 + "|" + e2 + "|" + e3 + "|" + chk
}

func hmacSHA256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
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

	injectLicense := c.PostForm("inject_license") == "true"
	licenseKey := c.PostForm("license_key")
	programID := c.PostForm("program_id")
	serverURL := c.PostForm("server_url")

	unauthHTML := models.GetSetting("unauth_page_html")
	if unauthHTML == "" {
		unauthHTML = defaultUnauthHTML()
	}

	secKey := utils.GenerateRandomString(16)
	authFiles := generateAuthFiles(language, secKey, unauthHTML)

	if injectLicense && licenseKey != "" && programID != "" && serverURL != "" {
		authFiles[".lic"] = generateLicFile(licenseKey, programID, serverURL, secKey)
	}

	existingFiles := map[string]bool{}
	for _, f := range reader.File {
		existingFiles[f.Name] = true
	}
	entryTargets := entryFiles[language]

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, f := range reader.File {
		content, err := readZipFile(f)
		if err != nil {
			continue
		}

		shouldInject := false
		for _, ef := range entryTargets {
			if matchEntryFile(f.Name, ef) {
				shouldInject = true
				break
			}
		}

		if shouldInject && (language == "php" || language == "python" || language == "nodejs") {
			relSDK := relativeSDKPath(f.Name, sdkDir)
			content = injectEntryFile(language, content, relSDK)
		}

		w, err := writer.Create(f.Name)
		if err != nil {
			continue
		}
		w.Write([]byte(content))
	}

	for name, content := range authFiles {
		w, err := writer.Create(filepath.Join(sdkDir, name))
		if err != nil {
			continue
		}
		w.Write([]byte(content))
	}

	if language == "python" {
		initPath := filepath.Join(sdkDir, "__init__.py")
		if !existingFiles[initPath] {
			w, _ := writer.Create(initPath)
			w.Write([]byte(""))
		}
		parts := strings.Split(sdkDir, "/")
		for i := range parts {
			p := filepath.Join(strings.Join(parts[:i+1], "/"), "__init__.py")
			if !existingFiles[p] {
				w, _ := writer.Create(p)
				w.Write([]byte(""))
			}
		}
	}

	writer.Close()

	outputName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)) + "_authorized.zip"

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", outputName))
	c.Data(200, "application/zip", buf.Bytes())
}

func readZipFile(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	return string(data), err
}

func matchEntryFile(zipName, pattern string) bool {
	base := filepath.Base(zipName)
	patBase := filepath.Base(pattern)
	if base != patBase {
		return false
	}
	parts := strings.Split(filepath.ToSlash(zipName), "/")
	if len(parts) <= 2 {
		return true
	}
	patParts := strings.Split(filepath.ToSlash(pattern), "/")
	if len(patParts) > 1 {
		return strings.HasSuffix(filepath.ToSlash(zipName), pattern)
	}
	return true
}

func relativeSDKPath(entryFile, sdkPath string) string {
	entryDir := filepath.Dir(entryFile)
	parts := strings.Split(filepath.ToSlash(entryDir), "/")
	depth := 0
	for _, p := range parts {
		if p != "" && p != "." {
			depth++
		}
	}
	if depth == 0 {
		return sdkPath
	}
	prefix := strings.Repeat("../", depth)
	return prefix + sdkPath
}

func generateAuthFiles(language, secKey, unauthHTML string) map[string]string {
	files := make(map[string]string)

	encodedHTML := base64.StdEncoding.EncodeToString([]byte(unauthHTML))

	switch language {
	case "php":
		sdkCode := generatePHPSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode
		files["env.dat"] = xorEncode("YUYUE_AUTH_MARKER", secKey)

	case "python":
		sdkCode := generatePythonSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode
		files["__init__.py"] = ""

	case "nodejs":
		sdkCode := generateNodeSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode

	case "go":
		sdkCode := generateGoSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode

	case "java":
		sdkCode := generateJavaSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode

	case "csharp":
		sdkCode := generateCSharpSDK(secKey, encodedHTML)
		files[sdkFileName(language)] = sdkCode
	}

	return files
}

func generatePHPSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`<?php
defined('APP_START') or define('APP_START', microtime(true));
$_sys_k = '%s';
$_sys_d = __DIR__;
$_sys_f = $_sys_d . DIRECTORY_SEPARATOR . '.lic';

function _sys_xd($d, $k) {
    $r = base64_decode($d);
    $o = '';
    for ($i = 0; $i < strlen($r); $i++) {
        $o .= chr(ord($r[$i]) ^ ord($k[$i %% strlen($k)]));
    }
    return $o;
}

function _sys_hm($d, $k) {
    return hash_hmac('sha256', $d, $k);
}

function _sys_vf($key, $pid, $url, $k) {
    $ts = time();
    $sg = _sys_hm($key . $pid . $ts, $k);
    $payload = json_encode([
        'license_key' => $key,
        'program_id'  => (int)$pid,
        'device_info' => php_uname('n') . '|' . php_uname('m'),
        '_ts' => $ts,
        '_sg' => $sg
    ]);
    $ch = curl_init($url . '/api/license/verify');
    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => $payload,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
        CURLOPT_TIMEOUT => 10,
        CURLOPT_SSL_VERIFYPEER => false,
    ]);
    $resp = curl_exec($ch);
    $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    if (!$resp || $code !== 200) return false;
    $r = json_decode($resp, true);
    return isset($r['code']) && $r['code'] === 0;
}

$_sys_ok = false;

if (file_exists($_sys_f)) {
    $_sys_ld = @file_get_contents($_sys_f);
    if ($_sys_ld) {
        $_sys_parts = explode('|', $_sys_ld, 4);
        if (count($_sys_parts) === 4) {
            $_sys_lk = _sys_xd($_sys_parts[0], $_sys_k);
            $_sys_pid = _sys_xd($_sys_parts[1], $_sys_k);
            $_sys_url = _sys_xd($_sys_parts[2], $_sys_k);
            $_sys_chk = $_sys_parts[3];
            if (_sys_hm($_sys_lk . $_sys_pid . $_sys_url, $_sys_k) === $_sys_chk) {
                $_sys_ok = _sys_vf($_sys_lk, $_sys_pid, $_sys_url, $_sys_k);
            }
        }
    }
}

if ($_sys_ok) return;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['_sys_action']) && $_POST['_sys_action'] === 'activate') {
    $_sys_lk = trim($_POST['_sys_lk'] ?? '');
    $_sys_pid = trim($_POST['_sys_pid'] ?? '');
    $_sys_url = trim($_POST['_sys_url'] ?? '');
    if ($_sys_lk && $_sys_pid && $_sys_url) {
        if (_sys_vf($_sys_lk, $_sys_pid, $_sys_url, $_sys_k)) {
            $_sys_chk = _sys_hm($_sys_lk . $_sys_pid . $_sys_url, $_sys_k);
            $enc = implode('|', [
                base64_encode(_sys_xd_enc($_sys_lk, $_sys_k)),
                base64_encode(_sys_xd_enc($_sys_pid, $_sys_k)),
                base64_encode(_sys_xd_enc($_sys_url, $_sys_k)),
                $_sys_chk
            ]);
            // re-encode with xor for storage
            $_sys_store = implode('|', [
                _sys_xe($_sys_lk, $_sys_k),
                _sys_xe($_sys_pid, $_sys_k),
                _sys_xe($_sys_url, $_sys_k),
                $_sys_chk
            ]);
            @file_put_contents($_sys_f, $_sys_store);
            header('Location: ' . $_SERVER['REQUEST_URI']);
            exit;
        } else {
            $_sys_err = 'LICENSE_INVALID';
        }
    }
}

function _sys_xe($d, $k) {
    $o = '';
    for ($i = 0; $i < strlen($d); $i++) {
        $o .= chr(ord($d[$i]) ^ ord($k[$i %% strlen($k)]));
    }
    return base64_encode($o);
}

$_sys_html = base64_decode('%s');
if (isset($_sys_err)) {
    $_sys_html = str_replace('<!--ERR-->', '<p style="color:#dc2626;text-align:center;margin:8px 0">授权码无效，请检查后重试</p>', $_sys_html);
}
echo $_sys_html;
exit;
`, secKey, encodedHTML)
}

func generatePythonSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`import os, sys, json, base64, hashlib, hmac, platform, time
from urllib.request import Request, urlopen
from urllib.error import URLError

_K = '%s'
_D = os.path.dirname(os.path.abspath(__file__))
_F = os.path.join(_D, '.lic')

def _xd(d, k):
    r = base64.b64decode(d)
    return bytes([r[i] ^ ord(k[i %% len(k)]) for i in range(len(r))])

def _xe(d, k):
    r = bytes([ord(d[i]) ^ ord(k[i %% len(k)]) for i in range(len(d))])
    return base64.b64encode(r).decode()

def _hm(d, k):
    return hmac.new(k.encode(), d.encode(), hashlib.sha256).hexdigest()

def _vf(key, pid, url, k):
    try:
        ts = str(int(time.time()))
        sg = _hm(key + pid + ts, k)
        payload = json.dumps({
            'license_key': key, 'program_id': int(pid),
            'device_info': platform.node() + '|' + platform.machine(),
            '_ts': ts, '_sg': sg
        }).encode()
        req = Request(url + '/api/license/verify', data=payload,
                      headers={'Content-Type': 'application/json'}, method='POST')
        with urlopen(req, timeout=10) as resp:
            r = json.loads(resp.read())
            return r.get('code') == 0
    except Exception:
        return False

_ok = False
if os.path.exists(_F):
    try:
        with open(_F, 'r') as f:
            parts = f.read().strip().split('|', 3)
        if len(parts) == 4:
            lk = _xd(parts[0], _K).decode()
            pid = _xd(parts[1], _K).decode()
            url = _xd(parts[2], _K).decode()
            chk = parts[3]
            if _hm(lk + pid + url, _K) == chk:
                _ok = _vf(lk, pid, url, _K)
    except Exception:
        pass

if not _ok:
    print('This application requires a valid license.')
    print('Please configure the license in:', _F)
    print('Format: encoded_key|encoded_pid|encoded_url|hmac_checksum')
    sys.exit(1)
`, secKey)
}

func generateNodeSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`'use strict';
const _c = require('crypto'), _h = require('http'), _hs = require('https');
const _p = require('path'), _f = require('fs'), _o = require('os');
const _K = '%s';
const _D = __dirname;
const _LF = _p.join(_D, '.lic');

function _xd(d, k) {
  const r = Buffer.from(d, 'base64');
  return Buffer.from(r.map((b, i) => b ^ k.charCodeAt(i %% k.length))).toString();
}
function _xe(d, k) {
  const r = Buffer.from(d).map((b, i) => b ^ k.charCodeAt(i %% k.length));
  return Buffer.from(r).toString('base64');
}
function _hm(d, k) {
  return _c.createHmac('sha256', k).update(d).digest('hex');
}
function _vf(key, pid, url, k) {
  return new Promise((resolve) => {
    const ts = Math.floor(Date.now()/1000).toString();
    const sg = _hm(key + pid + ts, k);
    const data = JSON.stringify({
      license_key: key, program_id: parseInt(pid),
      device_info: _o.hostname() + '|' + _o.arch(),
      _ts: ts, _sg: sg
    });
    const u = new URL(url + '/api/license/verify');
    const client = u.protocol === 'https:' ? _hs : _h;
    const req = client.request({
      hostname: u.hostname, port: u.port, path: u.pathname,
      method: 'POST', headers: {'Content-Type':'application/json','Content-Length':Buffer.byteLength(data)},
      timeout: 10000
    }, (res) => {
      let body = '';
      res.on('data', c => body += c);
      res.on('end', () => {
        try { resolve(JSON.parse(body).code === 0); } catch(e) { resolve(false); }
      });
    });
    req.on('error', () => resolve(false));
    req.write(data); req.end();
  });
}

(async () => {
  let ok = false;
  if (_f.existsSync(_LF)) {
    try {
      const parts = _f.readFileSync(_LF, 'utf8').trim().split('|', 4);
      if (parts.length === 4) {
        const lk = _xd(parts[0], _K), pid = _xd(parts[1], _K), url = _xd(parts[2], _K);
        if (_hm(lk + pid + url, _K) === parts[3]) {
          ok = await _vf(lk, pid, url, _K);
        }
      }
    } catch(e) {}
  }
  if (!ok) {
    const html = Buffer.from('%s', 'base64').toString();
    try {
      const express = require.main && require.main.exports;
    } catch(e) {}
    console.error('License verification failed. Configure license in: ' + _LF);
    process.exit(1);
  }
})();
`, secKey, encodedHTML)
}

func generateGoSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`package bootstrap

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var _k = "%s"

func init() {
	dir := selfDir()
	licFile := filepath.Join(dir, ".lic")
	if data, err := os.ReadFile(licFile); err == nil {
		parts := strings.SplitN(string(data), "|", 4)
		if len(parts) == 4 {
			lk := xd(parts[0], _k)
			pid := xd(parts[1], _k)
			url := xd(parts[2], _k)
			chk := parts[3]
			if hm(lk+pid+url, _k) == chk {
				if vf(lk, pid, url, _k) {
					return
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "License verification failed.")
	fmt.Fprintln(os.Stderr, "Configure license in:", licFile)
	os.Exit(1)
}

func selfDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}

func xd(d, k string) string {
	r, _ := base64.StdEncoding.DecodeString(d)
	out := make([]byte, len(r))
	for i := range r { out[i] = r[i] ^ k[i%%len(k)] }
	return string(out)
}

func hm(d, k string) string {
	h := hmac.New(sha256.New, []byte(k))
	h.Write([]byte(d))
	return hex.EncodeToString(h.Sum(nil))
}

func vf(key, pid, url, k string) bool {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sg := hm(key+pid+ts, k)
	hostname, _ := os.Hostname()
	payload, _ := json.Marshal(map[string]interface{}{
		"license_key": key, "program_id": atoi(pid),
		"device_info": hostname + "|" + runtime.GOARCH,
		"_ts": ts, "_sg": sg,
	})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url+"/api/license/verify", "application/json", bytes.NewReader(payload))
	if err != nil { return false }
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	code, _ := result["code"].(float64)
	return code == 0
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
`, secKey)
}

func generateJavaSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`import java.io.*;
import java.net.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.security.*;
import java.util.Base64;
import javax.crypto.*;
import javax.crypto.spec.*;

public class Bootstrap {
    private static final String K = "%s";

    static {
        try {
            String dir = new File(Bootstrap.class.getProtectionDomain()
                .getCodeSource().getLocation().toURI()).getParent();
            File licFile = new File(dir, ".lic");
            boolean ok = false;
            if (licFile.exists()) {
                String content = new String(Files.readAllBytes(licFile.toPath()), StandardCharsets.UTF_8).trim();
                String[] parts = content.split("\\|", 4);
                if (parts.length == 4) {
                    String lk = xd(parts[0], K);
                    String pid = xd(parts[1], K);
                    String url = xd(parts[2], K);
                    if (hm(lk + pid + url, K).equals(parts[3])) {
                        ok = vf(lk, pid, url, K);
                    }
                }
            }
            if (!ok) {
                System.err.println("License verification failed. Configure: " + licFile.getPath());
                System.exit(1);
            }
        } catch (Exception e) {
            System.err.println("License check error: " + e.getMessage());
            System.exit(1);
        }
    }

    private static String xd(String d, String k) {
        byte[] r = Base64.getDecoder().decode(d);
        byte[] out = new byte[r.length];
        for (int i = 0; i < r.length; i++) out[i] = (byte)(r[i] ^ k.charAt(i %% k.length()));
        return new String(out, StandardCharsets.UTF_8);
    }

    private static String hm(String d, String k) throws Exception {
        Mac mac = Mac.getInstance("HmacSHA256");
        mac.init(new SecretKeySpec(k.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
        byte[] hash = mac.doFinal(d.getBytes(StandardCharsets.UTF_8));
        StringBuilder sb = new StringBuilder();
        for (byte b : hash) sb.append(String.format("%%02x", b));
        return sb.toString();
    }

    private static boolean vf(String key, String pid, String url, String k) {
        try {
            long ts = System.currentTimeMillis() / 1000;
            String sg = hm(key + pid + ts, k);
            String hostname = InetAddress.getLocalHost().getHostName();
            String json = String.format(
                "{\"license_key\":\"%%s\",\"program_id\":%%s,\"device_info\":\"%%s\",\"_ts\":\"%%d\",\"_sg\":\"%%s\"}",
                key, pid, hostname, ts, sg);
            HttpURLConnection conn = (HttpURLConnection) new URL(url + "/api/license/verify").openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Content-Type", "application/json");
            conn.setDoOutput(true);
            conn.setConnectTimeout(10000);
            conn.setReadTimeout(10000);
            conn.getOutputStream().write(json.getBytes(StandardCharsets.UTF_8));
            BufferedReader br = new BufferedReader(new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8));
            StringBuilder resp = new StringBuilder();
            String line;
            while ((line = br.readLine()) != null) resp.append(line);
            return resp.toString().contains("\"code\":0");
        } catch (Exception e) { return false; }
    }
}
`, secKey)
}

func generateCSharpSDK(secKey, encodedHTML string) string {
	return fmt.Sprintf(`using System;
using System.IO;
using System.Net.Http;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace App.Internal {
    internal static class Bootstrap {
        private static readonly string K = "%s";
        private static readonly HttpClient C = new HttpClient { Timeout = TimeSpan.FromSeconds(10) };

        [System.Runtime.CompilerServices.ModuleInitializer]
        internal static void Init() {
            string dir = AppDomain.CurrentDomain.BaseDirectory;
            string licFile = Path.Combine(dir, ".lic");
            bool ok = false;
            if (File.Exists(licFile)) {
                try {
                    string[] parts = File.ReadAllText(licFile).Trim().Split('|', 4);
                    if (parts.Length == 4) {
                        string lk = Xd(parts[0], K);
                        string pid = Xd(parts[1], K);
                        string url = Xd(parts[2], K);
                        if (Hm(lk + pid + url, K) == parts[3])
                            ok = Vf(lk, pid, url, K).GetAwaiter().GetResult();
                    }
                } catch {}
            }
            if (!ok) {
                Console.Error.WriteLine("License verification failed. Configure: " + licFile);
                Environment.Exit(1);
            }
        }

        static string Xd(string d, string k) {
            byte[] r = Convert.FromBase64String(d);
            byte[] o = new byte[r.Length];
            for (int i = 0; i < r.Length; i++) o[i] = (byte)(r[i] ^ k[i %% k.Length]);
            return Encoding.UTF8.GetString(o);
        }

        static string Hm(string d, string k) {
            using var hm = new HMACSHA256(Encoding.UTF8.GetBytes(k));
            byte[] hash = hm.ComputeHash(Encoding.UTF8.GetBytes(d));
            return BitConverter.ToString(hash).Replace("-", "").ToLower();
        }

        static async System.Threading.Tasks.Task<bool> Vf(string key, string pid, string url, string k) {
            try {
                long ts = DateTimeOffset.UtcNow.ToUnixTimeSeconds();
                string sg = Hm(key + pid + ts, k);
                var payload = new { license_key = key, program_id = int.Parse(pid),
                    device_info = Environment.MachineName, _ts = ts.ToString(), _sg = sg };
                var json = JsonSerializer.Serialize(payload);
                var resp = await C.PostAsync(url + "/api/license/verify",
                    new StringContent(json, Encoding.UTF8, "application/json"));
                var body = await resp.Content.ReadAsStringAsync();
                using var doc = JsonDocument.Parse(body);
                return doc.RootElement.GetProperty("code").GetInt32() == 0;
            } catch { return false; }
        }
    }
}
`, secKey)
}

func defaultUnauthHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>程序授权验证</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#f8fafc}
.box{width:100%;max-width:420px;padding:40px;background:#fff;border-radius:8px;box-shadow:0 1px 3px rgba(0,0,0,.1)}
h2{font-size:20px;font-weight:600;color:#1e293b;text-align:center;margin-bottom:8px}
.sub{font-size:14px;color:#64748b;text-align:center;margin-bottom:24px}
label{display:block;font-size:13px;font-weight:500;color:#374151;margin-bottom:6px}
input{width:100%;padding:10px 12px;border:1px solid #d1d5db;border-radius:6px;font-size:14px;outline:none;transition:border-color .2s}
input:focus{border-color:#3b82f6}
.field{margin-bottom:16px}
button{width:100%;padding:10px;background:#3b82f6;color:#fff;border:none;border-radius:6px;font-size:14px;font-weight:500;cursor:pointer;transition:background .2s}
button:hover{background:#2563eb}
<!--ERR-->
</style>
</head>
<body>
<div class="box">
<h2>程序授权验证</h2>
<p class="sub">此程序需要授权才能使用，请输入授权信息</p>
<form method="POST">
<input type="hidden" name="_sys_action" value="activate">
<div class="field"><label>授权码</label><input type="text" name="_sys_lk" placeholder="XXXX-XXXX-XXXX-XXXX" required></div>
<div class="field"><label>程序ID</label><input type="text" name="_sys_pid" placeholder="请输入程序ID" required></div>
<div class="field"><label>授权服务器地址</label><input type="text" name="_sys_url" placeholder="https://your-auth-server.com" required></div>
<button type="submit">激活授权</button>
</form>
</div>
</body>
</html>`
}
