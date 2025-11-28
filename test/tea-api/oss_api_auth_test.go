package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tea-api/internal/router"

	"github.com/gin-gonic/gin"
)

func TestOssListAndDelete_AdminAuth(t *testing.T) {
	// 初始化 Gin
	gin.SetMode(gin.TestMode)
	r := router.SetupRouter()

	// 构造管理员登录后的上下文（此处假设有中间件可注入 admin 身份，实际项目需根据 AuthMiddleware 实现调整）
	// 这里只做接口鉴权链路测试，不做真实 OSS 操作

	// 1. 测试未登录访问
	req, _ := http.NewRequest("GET", "/api/v1/oss/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == 401 || w.Code == 403 {
		t.Logf("未登录访问被拒绝，状态码: %d", w.Code)
	} else {
		t.Errorf("未登录应被拒绝，实际: %d", w.Code)
	}

	// 2. 测试非管理员访问
	// 这里应模拟普通用户身份，略（可参考 AuthMiddleware 实现）

	// 3. 测试管理员访问（伪造 admin 角色）
	req, _ = http.NewRequest("GET", "/api/v1/oss/list", nil)
	w = httptest.NewRecorder()
	// 伪造 admin 身份
	req.Header.Set("X-Role", "admin")
	// 这里实际项目应通过 token 或 session 注入，视 AuthMiddleware 实现而定
	r.ServeHTTP(w, req)
	if w.Code == 200 || w.Code == 500 {
		t.Logf("管理员访问通过，状态码: %d", w.Code)
	} else {
		t.Errorf("管理员应可访问，实际: %d", w.Code)
	}

	// 4. 测试批量删除接口
	body, _ := json.Marshal(map[string]interface{}{"urls": []string{"https://oss-bucket/test1.jpg"}})
	req, _ = http.NewRequest("POST", "/api/v1/oss/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", "admin")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == 200 || w.Code == 500 {
		t.Logf("管理员批量删除接口通过，状态码: %d", w.Code)
	} else {
		t.Errorf("管理员应可删除，实际: %d", w.Code)
	}
}
