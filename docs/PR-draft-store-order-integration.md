# PR 草案：门店与订单联动 + 真实登录 / captcha -> Redis

概要：
- 按“方向 2（门店与订单联动）”继续开发。
- 后端：新增门店订单相关接口并将验证码(captcha)改为 Redis 存储/校验，登录使用真实 `POST /auth/login` 返回 JWT（1h TTL）。
- 前端：`Admin-FE/index.html` 登录模态改为真实调用 `/auth/login`，移除开发角色选择，加入 captcha 显示与填写（开发模式支持显示验证码）；实现门店面板中订单列表展示，支持“带入并查看”把订单 ID 填入工具区并展示详情。
- 测试：增加 Playwright smoke test `tools/browser-login/login-test.js`，尝试纯 UI 登录流程并在失败时产出诊断报告。最终决定放弃持续调优，保留回退策略以便 CI 必要时注入 token（但当前分支保留纯 UI 优先行为）。

变更文件（主要）:
- `API-Server/auth.go`：captcha -> Redis，新增 `/auth/captcha`，`/auth/login` 校验 captcha 并生成 JWT（1小时过期）。
- `API-Server/main.go`：新增 `Order` 结构与 `storeOrdersHandler`、`orderDetailHandler` 等 mock 接口。
- `Admin-FE/index.html`：移除 role 选择，增加 captcha UI、门店订单表、带入并查看逻辑与详情区。
- `tools/browser-login/login-test.js`：Playwright 脚本增强，增加门店->订单流程断言并输出 `report-*.json`。
- `docs/PR-captcha-redis.md`：变更说明与测试步骤。

测试说明（本地）：
1. 启动 API Server（端口 8082）。
2. 在项目根目录启动静态 Admin-FE（例如 `python -m http.server 8000` 或其他）。
3. 运行 Playwright smoke test（headless）并查看 `tools/browser-login/report-*.json`：

```powershell
$env:ADMIN_FE_URL='http://localhost:8000'; $env:TEST_USER='admin'; $env:TEST_PASS='pass'; $env:HEADLESS='1'; node tools\browser-login\login-test.js
```

注意：Playwright 的纯 UI 路径在不同环境可能存在 timing/selector 差异，若 CI 上不稳定，可在后续 PR 中添加受控回退注入 token 的逻辑并由 CI 使用。

请求：请确认是否要我将这些变更本地提交到 `feature/store-order-integration` 分支并生成 PR（仅在本地提交），或同时将分支推送到远程仓库并创建远程 PR（需要你授权允许推送）。

决策更新（已确认）：
- 已决定：**不在当前工作区执行任何 Git 相关操作**（不初始化 `.git`、不创建本地分支、也不推送到远程）。
- 原因：用户明确要求放弃所有 git 相关处理；本 PR 草案与变更已记录在仓库文档中，后续由项目维护者在适当的 Git 环境中执行提交与代码审查流程。

后续建议：
- 若需要我可以生成一个可直接应用的 patch 文件或 `git-format-patch` 风格的补丁内容，便于你在正式仓库中应用而无需我直接操作 Git。
- 若要我准备该补丁，请回复“生成补丁”，我将创建 `docs/changes.patch` 或类似文件供下载/审阅。
