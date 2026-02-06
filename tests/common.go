package tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	DefaultBaseURL = "https://localhost:8890/api/v1"
)

// GetBaseURL 获取测试基础 URL
func GetBaseURL() string {
	v := os.Getenv("SPARKX_BASE_URL")
	if v == "" {
		return DefaultBaseURL
	}
	return strings.TrimRight(v, "/")
}

// GetenvDefault 获取环境变量，如果不存在则返回默认值
func GetenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==========================================
// HTTP Client for Tests
// ==========================================

// TestClient 测试用 HTTP 客户端
type TestClient struct {
	T     *testing.T
	Token string
}

// SetToken 设置 JWT Token
func (c *TestClient) SetToken(token string) {
	c.Token = token
}

// Post 发送 POST 请求
func (c *TestClient) Post(path string, body interface{}, target interface{}) {
	c.Do("POST", path, body, target)
}

// Put 发送 PUT 请求
func (c *TestClient) Put(path string, body interface{}, target interface{}) {
	c.Do("PUT", path, body, target)
}

// Get 发送 GET 请求
func (c *TestClient) Get(path string, target interface{}) {
	c.Do("GET", path, nil, target)
}

// Delete 发送 DELETE 请求
func (c *TestClient) Delete(path string) {
	c.Do("DELETE", path, nil, nil)
}

// Do 发送 HTTP 请求
func (c *TestClient) Do(method, path string, body interface{}, target interface{}) {
	c.DoWithStatusCheck(method, path, body, target, true)
}

// DoWithStatusCheck 发送 HTTP 请求，可选择是否检查状态码
func (c *TestClient) DoWithStatusCheck(method, path string, body interface{}, target interface{}, checkStatus bool) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			c.T.Fatalf("Failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, GetBaseURL()+path, bodyReader)
	if err != nil {
		c.T.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		c.T.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.T.Fatalf("Failed to read response body: %v", err)
	}

	if checkStatus && resp.StatusCode >= 400 {
		c.T.Fatalf("API Error %d: %s", resp.StatusCode, string(respBytes))
	}

	if target != nil {
		if err := json.Unmarshal(respBytes, target); err != nil {
			c.T.Fatalf("Failed to unmarshal response: %v. Body: %s", err, string(respBytes))
		}
	}
}

// DeleteProjectBestEffort 尽力删除项目（用于清理）
func DeleteProjectBestEffort(t *testing.T, token string, id int64) {
	req, err := http.NewRequest("DELETE", GetBaseURL()+fmt.Sprintf("/projects/%d", id), nil)
	if err != nil {
		t.Logf("Cleanup failed to create request: %v", err)
		return
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Cleanup request failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		t.Logf("Cleanup API Error %d: %s", resp.StatusCode, string(b))
	}
}
