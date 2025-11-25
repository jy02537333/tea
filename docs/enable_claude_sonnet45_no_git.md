启用 Claude Sonnet 4.5（非侵入性运维指南，**不包含 git 操作**）

目标：
- 在运行环境（本地/服务器/容器/云实例）中把对话/推理请求切换或接入到 `claude-sonnet-4.5`，并保证可观测、可回滚、可控成本。该文档不包含任何 git 或代码仓库操作，仅介绍配置/运行/验证步骤，便于运维或平台控制台直接应用。

1) 前提条件
- 拥有 Anthropic（或目标模型提供方）API Key 与合规权限。
- 确认目标服务：例如后端 `tea-api`、Admin-FE 的后端代理或专门的模型中间层。
- 在变更前记录当前环境配置（用于回滚），例如当前 `MODEL_NAME`、`MODEL_API_URL`、服务启动方式等。

2) 推荐环境变量（把这些注入到服务运行环境或容器 Secrets）
- `MODEL_PROVIDER`：`anthropic`（可选，区分不同模型供应商）
- `MODEL_NAME`：`claude-sonnet-4.5`
- `MODEL_API_KEY`：机密值（严格通过 Secret 管理，不要写入日志或代码）
- `MODEL_API_URL`：如需覆盖默认 endpoint，可指定给定 URL

示例（Windows PowerShell 中设置环境变量，供运维/测试临时使用）：
```powershell
# 示例：仅用于临时测试，生产环境请使用 Secret 管理
setx MODEL_PROVIDER "anthropic"
setx MODEL_NAME "claude-sonnet-4.5"
setx MODEL_API_KEY "<你的_API_KEY>"
```

3) 服务端接入要点（供开发/运维配合时参考）
- 推荐抽象一层模型适配器（例如 `pkg/modelclient`）：根据 `MODEL_PROVIDER` 与 `MODEL_NAME` 选择实现。
- 适配器应封装：HTTP 客户端、超时、重试、速率限制、日志（不可记录敏感 prompt/response）以及错误码映射。
- 在业务代码中仅通过适配器接口调用（例如 `Generate(ctx, prompt, opts)`），便于运行时更换模型。
- 强制设定超时（如 10-20s），并对失败做降级处理（返回友好默认文本或调用缓存响应）。

4) 本地/单机验证（不需要仓库修改）
- 在运行服务的主机或测试机器上设置好环境变量（见第2步），然后重启服务进程。
  - 如果服务通过已存在的启动脚本运行，请用该脚本重启；如果通过 `go run` 运行，可先停止现有进程再执行 `go run main.go`（均非 git 操作）。
- 发送 smoke 请求验证：
  - 示例（PowerShell）：
```powershell
$body = @{ prompt = "测试：Hello" } | ConvertTo-Json
Invoke-RestMethod -Uri 'http://localhost:8080/api/v1/model/generate' -Method POST -Body $body -Headers @{ 'Authorization' = "Bearer $env:MODEL_API_KEY"; 'Content-Type' = 'application/json' }
```
- 检查返回文本是否合理，并记录延迟与错误码。

5) Smoke Tests 与负载验证
- 功能验证：在真实业务路径（例如客服自动回复或后台摘要任务）做一次端到端调用，确认语义质量与格式符合预期。
- 性能验证：对目标路径进行并发 10-50 次短时请求，记录 P50/P95 延迟与失败率。
- 成本控制：在测试中关注调用频率与 token 消耗，避免高频调用导致账单激增。

6) 回滚与降级（无需 git）
- 回滚方法：把 `MODEL_NAME` 改回原值或把 `MODEL_PROVIDER` 设为空/`none`，然后重启服务进程即可回退；记录旧配置以便快速恢复。
- 降级策略：实现本地短路（当模型调用失败时返回默认文本或从缓存读取），避免影响上游用户体验。

7) 监控与告警（建议）
- 指标：`model_calls_total`, `model_calls_failed`, `model_call_latency_seconds`。
- 告警：错误率超过阈值（如 2%）或 P95 延迟超过阈值（如 5s）时触发。
- 日志：记录 trace_id、耗时、状态码与错误摘要（注意不要记录完整 prompt/response）。

8) 运营/安全与权限
- `MODEL_API_KEY` 要通过 Secret 管理并限定访问人员。
- 将运维步骤写入内部 Wiki 或运维手册，并明确责任人（谁注入 Secret、谁重启服务、谁验证、谁监控费用）。

备注：
- 本文档仅说明运行时如何切换/接入模型且不包含任何仓库操作或 git 相关指令，便于在权限受限场景下由运维或平台控制台直接修改环境并验证。
- 如需将适配器代码、示例客户端或自动化验证脚本（如 Puppeteer/Playwright 的浏览器端登录验证脚本）加入仓库，请明确授权我修改代码仓库，我会把代码补丁与运行说明一并提供。
