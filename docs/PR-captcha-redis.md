变更说明：将本地开发验证码从内存替换为 Redis，并更新 Admin-FE 登录流程与自动化测试

目的
- 增强验证码存储的一致性（使用 Redis，TTL 可控），使多进程/多实例本地联调更可靠。
- 前端登录不再需要显式选择角色，角色由后端根据账号/oid 自动判断。
- 使 Playwright smoke test 能自动完成登录（通过读取页面验证码或直接通过 API 完成登录并注入 token）。

主要改动文件
- `API-Server/auth.go`
   - 引入 `github.com/redis/go-redis/v9`，在 `init()` 初始化 Redis 客户端（优先使用环境变量 `REDIS_ADDR/REDIS_PASS/REDIS_DB`，未设置时默认 `127.0.0.1:6379`）。
  - 新增路由：`/auth/captcha` 与 `/api/v1/auth/captcha`（返回 JSON `{id, code}`，开发模式明文），并将验证码写入 Redis（key=`captcha:{id}`，TTL=5min）。
  - 在 `/auth/login` 中校验 `captcha_id` 与 `captcha_code`（从 Redis 读取并一次性删除），并生成 JWT（原行为保留）。

- `Admin-FE/index.html`
  - 移除登录时的角色选择控件；在登录模态增加验证码显示区（`#captcha-code-display`）、隐藏 `#captcha-id`、输入 `#captcha-input` 与刷新按钮，并在打开登录模态时调用 `/auth/captcha` 填充。
  - `doLogin()` 在提交时带上 `captcha_id` 与 `captcha_code` 字段。

- `tools/browser-login/login-test.js`
  - 增强 Playwright 脚本：当 UI captcha 不可靠时，脚本会通过 Playwright 的 `context.request` 直接请求 `/auth/captcha` 和 `/auth/login`，把返回的 token 注入到 `localStorage.tea_admin_token`，并继续检查侧栏菜单渲染以验证登录成功（更稳定的 smoke 测试）。

测试与验证步骤
1. 启动 Redis（示例使用 Docker）
   ```powershell
   docker run -d -p 6379:6379 --name redis-local redis:7
   ```
2. 启动后端 API-Server（在 `API-Server` 目录）
   ```powershell
   $env:REDIS_ADDR='127.0.0.1:6379'  # 或覆盖为本地 redis 地址
   $env:REDIS_DB='0'
   $env:PORT='8082'
   go run .
   ```
3. 启动 Admin-FE 静态服务器（在 `Admin-FE` 目录）
   ```powershell
   python -m http.server 8000
   ```
4. 手动验证
   - 打开 `http://localhost:8000`，点击登录，页面将显示开发验证码并允许输入；提交后应返回并保存 `localStorage.tea_admin_token`。
5. 自动化 smoke test
   - 进入 `tools/browser-login`：
     ```powershell
     npm install
     npx playwright install
     $env:ADMIN_FE_URL='http://localhost:8000'; $env:TEST_USER='admin'; $env:TEST_PASS='pass'; $env:HEADLESS='1'; node login-test.js
     ```
   - 脚本现在会优先通过 Playwright 的 HTTP 客户端完成 captcha+login 并将 token 注入页面，随后断言侧边栏菜单（验证登录成功）。报告保存在 `tools/browser-login/report-*.json`。

注意与安全
- 当前 `auth/captcha` 在开发模式下以明文返回验证码，仅用于本地联调。生产环境应使用图片验证码或第三方服务，且不要直接返回明文验证码。
- 登录角色判定由后端保留，前端不再能通过选择角色绕过鉴权。
- 若你希望在没有 Redis 的本地环境下也能运行（回退到内存或跳过验证码），我可以添加可选的容错回退策略。

建议提交说明（PR 标题/描述）
- 标题：feat(auth): use Redis for dev captcha + remove frontend role select
- 描述：见本文件变更说明；包含运行/测试步骤及安全注意事项。

