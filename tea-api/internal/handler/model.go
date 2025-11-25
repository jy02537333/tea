package handler

import (
	"context"
	"net/http"
	"tea-test/pkg/env"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/pkg/modelclient"
	"tea-api/pkg/utils"
)

type ModelHandler struct{}

// modelClientFactory 可以在测试中被替换以注入 mock
var modelClientFactory = defaultModelClientFactory

func defaultModelClientFactory() (modelclient.Client, error) {
	ai := config.Config.AI
	apiKeyEnv := ai.APIKeyEnv
	if apiKeyEnv == "" {
		apiKeyEnv = "MODEL_API_KEY"
	}
	apiKey := env.Get(apiKeyEnv, "")
	apiURL := ai.APIURL
	model := ai.DefaultModel
	timeout := time.Duration(ai.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return modelclient.NewAnthropicClient(apiKey, apiURL, model, timeout)
}

func NewModelHandler() *ModelHandler {
	return &ModelHandler{}
}

// Generate 接口接收 { prompt: string } 并返回模型生成结果
func (h *ModelHandler) Generate(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Prompt == "" {
		utils.InvalidParam(c, "prompt 不能为空")
		return
	}

	// 允许通过 config.AI.Enabled 或 环境变量 MODEL_API_KEY 来启用
	ai := config.Config.AI
	apiKeyEnv := ai.APIKeyEnv
	if apiKeyEnv == "" {
		apiKeyEnv = "MODEL_API_KEY"
	}
	apiKey := env.Get(apiKeyEnv, "")
	if !ai.Enabled && apiKey == "" {
		utils.Forbidden(c, "AI 模型未启用或未配置 API Key")
		return
	}

	client, err := modelClientFactory()
	if err != nil {
		utils.ServerError(c, "创建模型客户端失败: "+err.Error())
		return
	}

	// 设置超时（以配置优先）
	timeout := time.Duration(ai.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	out, err := client.Generate(ctx, req.Prompt)
	if err != nil {
		// 如果是超时或上下文取消，返回 504
		if err == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"code": 504, "message": "模型调用超时"})
			return
		}
		utils.ServerError(c, "模型调用失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"output": out})
}
