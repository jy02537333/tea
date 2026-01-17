package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tea-api/internal/config"
	"tea-api/pkg/modelclient"

	"github.com/gin-gonic/gin"
)

// 测试未配置模型时返回 403
func TestGenerate_Forbidden(t *testing.T) {
	// 确保环境为非启用
	config.Config.AI.Enabled = false
	_ = os.Unsetenv("MODEL_API_KEY")
	// 跳过数据库初始化以便单元测试在无外部依赖下运行（注意：不要影响同包其它测试）
	prev, had := os.LookupEnv("TEA_SKIP_DB_INIT")
	_ = os.Setenv("TEA_SKIP_DB_INIT", "1")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("TEA_SKIP_DB_INIT", prev)
			return
		}
		_ = os.Unsetenv("TEA_SKIP_DB_INIT")
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewModelHandler()

	r.POST("/model/generate", h.Generate)

	w := httptest.NewRecorder()
	// 发送有效的 body（包含 prompt），以绕过参数校验并触发权限判断
	body := map[string]string{"prompt": "test-forbidden"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/model/generate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 when model disabled; got %d, body: %s", w.Code, w.Body.String())
	}
}

// 测试启用并通过 mock client 成功返回
func TestGenerate_Success(t *testing.T) {
	// 启用 AI
	config.Config.AI.Enabled = true
	// 覆盖工厂以返回 mock client
	orig := modelClientFactory
	defer func() { modelClientFactory = orig }()

	modelClientFactory = func() (modelclient.Client, error) {
		return &mockClient{}, nil
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewModelHandler()
	r.POST("/model/generate", h.Generate)

	body := map[string]string{"prompt": "hello"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/model/generate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200; got %d, body: %s", w.Code, w.Body.String())
	}
}

type mockClient struct{}

func (m *mockClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "mock-output", nil
}
