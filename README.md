# API Validation CI Templates

> 主线门禁：`master` 分支已启用分支保护，合并必须通过 “API Validation” 工作流（Sprint A 阻塞、Sprint B 非阻塞）。详见 `doc/prd.md` 与 `doc/prd_sprints.md`。

[![API Validation](https://github.com/jy02537333/tea/actions/workflows/api-validation.yml/badge.svg?branch=master)](https://github.com/jy02537333/tea/actions/workflows/api-validation.yml)

前端联调快捷入口： [admin-fe/README.md](admin-fe/README.md) · [wx-fe/README.md](wx-fe/README.md)

管理端演示提示：可在后台配合 `scripts/dev-order-cli.sh` 一键演示发货/完成/取消（deliver/receive/cancel），详见本文“端到端订单联调与开发者 CLI（演示）”。

## 最小 UI 测试（基于 Mock API，一键运行）

本仓已内置最小前端 UI 用例（Playwright）：

- 订单页：点击“管理员取消/退款”前置确认后应弹出“原因输入”弹窗。
- 合伙人提现：列表进入“审核”弹窗，能“受理”或“拒绝并解冻”。

无需真实后端，使用内置 Mock API 启动并一键跑通：

```bash
# 在仓库根目录
chmod +x scripts/dev-e2e.sh
scripts/dev-e2e.sh
```

脚本行为：
- 启动 Mock API（默认 http://127.0.0.1:9393）与静态 SPA 服务（http://127.0.0.1:5173）。
- 安装依赖、构建 admin-fe、安装 Playwright 浏览器。
- 运行两条最小 UI 测试：
  - admin-fe/tests/orders-reason-modal.spec.ts
  - admin-fe/tests/partner-withdrawal.spec.ts
- 产物输出到 `build-ci-logs/playwright/`（截图、Trace、服务日志）。

高级用法：
```bash
# 仅运行其中一条测试（示例：订单原因弹窗）
ADMIN_FE_URL=http://127.0.0.1:5173 \
API_BASE=http://127.0.0.1:9393 \
pnpm -C admin-fe exec playwright test tests/orders-reason-modal.spec.ts --workers=1
```

说明：前端在测试时通过 `window.__TEA_RUNTIME_CONFIG__` 将 API 基址覆盖为 `API_BASE`，从而接入 Mock API。用例会将截图与 Trace 写入 `build-ci-logs/playwright/` 便于排查。

CI 手动触发：在 GitHub 仓库的 Actions 里选择 “E2E UI (Mock API)” 工作流，点击 “Run workflow” 即可在当前分支生成并上传 UI 证据工件（截图/Trace/日志）。
可选输入：
- tests：all | orders | withdrawal | 自定义文件/glob
- trace：true/false（是否生成 Playwright Trace）
- workers：并发数（默认 1）
- api_port：Mock API 端口（默认 9393）
- spa_port：SPA 服务端口（默认 5173）

### GitHub Secrets 占位说明（后端连通性 & 截图）

部分 CI 工作流会在 PR 上自动补充“后端连通性摘要”评论，并产出 wx-fe 财务页截图工件。为避免运行失败，请在仓库 Secrets 中预先配置以下键（均为可选，但推荐在正式环境中设置）：

- `API_BASE`：后端 API 基地址，例如 `https://api.example.com` 或 `http://127.0.0.1:9292`。
  - 用途：`tools/automation/backend-connectivity-summary.sh` 用于健康检查、门店列表与门店财务接口连通性检测，并将结果写入 `build-ci-logs/backend_connectivity.md`，由 CI 追加到 PR 评论中。
- `ADMIN_TOKEN`：具备门店与财务查询权限的管理员 JWT Token（`Authorization: Bearer <token>`）。
  - 用途：在有值时，CI 会在同一工作流中调用 `GET /api/v1/stores` 与门店财务相关接口，验证后端路由是否按预期升级并返回 200；为空时，仅执行健康检查并在 PR 评论中标注为“未提供管理员令牌”。

提示：
- 若暂不希望在 CI 中访问真实后端，可仅配置 `API_BASE` 指向测试环境，或完全不配置上述 Secrets，此时 CI 会跳过后台连通性检查步骤。
- Secrets 配置入口：GitHub 仓库 → Settings → Secrets and variables → Actions → New repository secret。

This folder contains CI job templates to run the repository's `scripts/run_api_validation.sh` in CI and collect artifacts.

Provided files:

- `.github/workflows/api-validation.yml` — GitHub Actions workflow that: checks out the repo, starts MySQL/Redis/RabbitMQ as services, builds and runs `tea-api`, seeds example SKU, runs `scripts/run_api_validation.sh`, and uploads `build-ci-logs` as artifacts.

- `api-validation.gitlab-ci.yml` — GitLab CI job template with similar steps (services: mysql/redis/rabbitmq). Use it in your project by including or copying into `.gitlab-ci.yml`.
#
How it works (summary):

1. Start dependent services (mysql/redis/rabbitmq) via CI-provided services.
2. Build and run `tea-api` in background on port `9292`.
3. Run seeder steps (example SKU creation) so validation has minimal data.
4. Run `sh scripts/run_api_validation.sh` which performs HTTP requests and writes responses under `build-ci-logs/api_validation/` and `build-ci-logs/admin_login_response.json`.
5. Upload `build-ci-logs` as CI artifacts for QA review.
6. Evidence artifacts: key verification files are saved under `build-ci-logs/` (e.g., `order_amounts_summary.json`, `order_detail_*_checked.json`, and stateful bodies under `api_validation_stateful/`) and are uploaded for easy triage. See generation and assertions in `scripts/local_api_check.sh` (evidence creation) and `scripts/assert_api_validation.sh` (strict checks).

Notes & tips:

- If your CI runner doesn't support Docker services, use a self-hosted runner with Docker and run `docker-compose up -d` instead.
- Adjust seeder steps in the workflow to match your required test data.
- Consider adding a short smoke test assertion that checks `build-ci-logs/api_validation/summary.txt` for expected 200 responses to fail the job when regressions occur.

## 移除 Git 历史中的大文件（推送被 100MB 限制阻断时）

- 使用步骤：
## 本地开发端口约定

- 后端 API 默认在 `:9292` 端口监听，基础地址为：`http://localhost:9292/api/v1`
- 请将前端与工具脚本（如 `API_BASE`、`VITE_API_BASE_URL`、`WX_API_BASE_URL` 等）统一指向 `http://localhost:9292`，避免误连到其他本机服务（如 `:8080`）。
- 如需变更端口，可在 `tea-api/configs/config.yaml` 或对应本地配置中调整，并同步更新前端环境变量。

### 保持 9292 服务可用的运维提示

- 遇到 `listen tcp :9292: bind: address already in use`，说明 9292 已被占用。
- 请优先确保 `tea-api` 独占 9292。推荐直接使用 `./run-tea-api.sh` 启动：脚本会在检测到 9292 被占用时自动强制终止占用进程并重启（必要时使用 `kill -9` 作为最后手段）。
- 如需手动处理，可按下列步骤排查/释放端口：

```
# 查看哪个进程占用了 9292
sudo lsof -iTCP:9292 -sTCP:LISTEN

# 杀死占用 9292 的进程
sudo fuser -k 9292/tcp

# 重新启动 tea-api（选择你的方式之一）
TEA_JWT_SECRET=dev_secret_change_me go run ./tea-api/main.go
# 或
./run-tea-api.sh
```

注意：不要切到 8080（常被其他服务占用）。统一坚持 9292，减少环境不一致带来的问题。
 
## E2E 验证工件（合伙人提现审核）

- 申请前截图（页面）: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/28/partner-withdrawal-before-click.png
- 审核弹窗截图: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/28/partner-withdrawal-modal.png
- Playwright Trace 压缩包: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/28/partner-withdrawal-trace.zip
- Trace Report HTML: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/28/partner-withdrawal-trace-report.html

说明：后端已实现 `POST /api/v1/admin/storage/oss/policy`，上传脚本默认使用表单直传策略以统一治理生命周期与前缀（建议 `ci_artifact/`）。示例直传外链（文本演示）：https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/28/oss_demo.txt。

- Admin OSS Policy 响应（脱敏）: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/29/admin_oss_policy_redacted.json
- Get OSS Policy 响应（脱敏）: https://zdw-img.oss-cn-beijing.aliyuncs.com/ci_artifact/2025/12/29/get_oss_policy_redacted.json

提示：该路由已实现，CI 已自动将 PR 顶部提示块中的该项勾选为完成（见 `.github/workflows/update-pr-checklist.yml`）。

## 统一登录与用户聚合接口（JWT v5）

- 统一采用 JWT v5 登录与鉴权，后端基础地址：`http://localhost:9292/api/v1`。
- 登录入口：`POST /api/v1/auth/login`（返回 `token` 与用户基础信息）。
- 用户聚合信息：`GET /api/v1/users/me/summary`（需 `Authorization: Bearer <token>`）。
- 所有脚本与联调请以以上两条路径为准，避免与历史实现（如 `/api/v1/user/login`、`/api/v1/user/info`）混用导致 Token 校验不一致。

## 统一编译与启动名称（避免多二进制混用）

- 为避免因不同可执行文件路由集不一致导致验证失败，统一仅使用单一二进制名：`tea-api`。
- 编译与启动示例：
  - 编译：在仓库根目录执行 `go build -o tea-api/tea-api ./tea-api`
  - 启动：`TEA_JWT_SECRET=dev_secret_change_me ./tea-api/tea-api`
- 清理旧二进制：如存在 `tea-api/server`、`tea-api/main`、`tea-api/main_no_migrate`，请删除，避免误启动历史路由实现。
  - 赋予脚本执行权限：`chmod +x scripts/remove-large-files.sh`
  - 在仓库根目录运行：`scripts/remove-large-files.sh`（默认阈值 100M，可传参自定义：`scripts/remove-large-files.sh 120M`）
  - 完成后按提示安全推送当前分支：`git push --force-with-lease origin $(git rev-parse --abbrev-ref HEAD)`
- `scripts/remove-secrets.sh`: 按 push protection 告警从工作区与历史移除敏感路径（`--paths-from-file` + `--path-glob`），自动恢复 `origin` 并安全推送；适用于被阻断的功能分支自助修复。
- `scripts/rewrite-master-history.sh`: 维护窗口内重写主分支历史的一键脚本，含备份分支创建、过滤规则、垃圾回收、远端恢复与安全强推；输出协作者对齐与回滚提示。
- `docs/ci/history-rewrite-maintenance-plan.md`: 主分支历史重写的作业计划与操作手册（步骤、脚本引用、风险、迁移、回滚、沟通模版）。
- `docs/ci/history-rewrite-notice.md`: 历史重写维护窗口的中/英通知模版，维护前/开始/完成阶段可直接复用。

## Git 推送方式约定（统一使用 Token）

- 推荐：所有开发者统一使用 HTTPS+GitHub Personal Access Token（PAT）进行推送，避免 SSH key 配置差异导致 push 失败。
- 远程地址形态示例：`https://github.com/jy02537333/tea.git`（而非 `git@github.com:jy02537333/tea.git`）。
- 首次推送：执行 `git push` 时，Git 会提示输入用户名和密码：
  - 用户名：填写你的 GitHub 用户名。
  - 密码：填写 PAT（注意不要使用真实登录密码）。
- 建议：在本地 Git Credential Helper 中缓存 PAT，避免每次输入；如需旋转或撤销 PAT，可在 GitHub "Settings → Developer settings → Personal access tokens" 中管理。
- 若 push 因 push protection / secret scan 被阻断，请参考 `tea-api/README.md` 中“开发者注意事项 / 遇到 Push Protection 阻塞时的自助处理”小节，使用 `scripts/remove-secrets.sh` 自助修复后再使用 HTTPS+Token 方式强推。

## 项目进度与规划文档索引

> 说明：下列路径均以仓库根目录为基准，便于快速找到「现在项目做到哪一步了」「本迭代要做什么」以及「哪些需求还在 Backlog」。当出现新文档时，请同步在此处补充，保持索引完整。

- `doc/prd.md`：茶心阁小程序的主 PRD 文档，描述整体业务需求、角色、功能范围与高层开发计划，是所有进度/任务文档的总入口。
- `doc/prd_sprints.md`：基于 PRD 拆解的 Sprint 任务列表，按迭代列出前端/后端/测试/运维的具体工作项，并给出关键 REST API 草案。
- `doc/prd_feature_checklist.md`：需求文档功能点与 `doc/prd.md` 章节的一一映射表，通过 `[x]/[]/✅` 标记设计与实现完成度，用于评审“哪些功能已在 PRD 中落地”。
- `tea-api/docs/progress-report.md`：后端 `tea-api` 的开发进度报告，按模块列出「已完成 / 进行中 / 待测试 / 下一阶段计划」，适合作为后端开发的进度视图。
- `docs/features/store-order-link.md`：门店面板 ↔ 订单列表 ↔ 订单操作区联动的设计说明，文末包含建议的开发步骤，是门店与订单联动方向的实现指南。
- `docs/prd-open-points.md`：PRD 中尚未完全敲定的开放问题和后续建议动作清单，可视为产品侧 Backlog 与决策待办。
- `docs/frontend-backend-checklist.md`：前后端联调与提测前的检查清单，用于在提测/发布前核对 API、字段、状态枚举等是否与 PRD 和 API 文档保持一致。

## 端到端订单联调与开发者 CLI（演示）

快速走通「结算→统一下单→支付模拟→订单详情」并观察状态变化（待支付→已付款→配送中→已完成/已取消），建议按以下步骤：

- 启动后端：

```bash
bash ./run-tea-api.sh
curl -sS -i http://127.0.0.1:9292/api/v1/health
```

- 启动前端并下单（wx-fe 已集成统一下单与支付模拟）：

```bash
pnpm -C wx-fe run dev:h5
# 或
pnpm -C wx-fe run dev:weapp
```

- 使用统一 CLI 进行订单状态演示（需管理员令牌，脚本会自动获取/复用）：

```bash
# 发货：推进到“配送中(3)”
bash scripts/dev-order-cli.sh deliver <ORDER_ID>

# 确认收货：推进到“已完成(4)”
bash scripts/dev-order-cli.sh receive <ORDER_ID>

# 取消订单：进入“已取消(5)”（可选原因）
bash scripts/dev-order-cli.sh cancel <ORDER_ID> -r "超时未支付"

# 指定端口（示例 9393）：
bash scripts/dev-order-cli.sh deliver <ORDER_ID> -p 9393
```

说明：
- 脚本会把响应保存到 `build-ci-logs/` 目录（如 `deliver_*.json`、`receive_*.json`、`cancel_*.json`）。
- 订单详情页（wx-fe）已开启自动轮询与手动刷新按钮，联调时可直观看到状态变化。

### 数据库设计文档（新增索引）

- `doc/db_schema.md`：数据库表结构设计概览与说明（ER 概览、主要实体与关系、设计原则），用于后端与 DBA 对齐字段与关系。
 - `db/schema.sql`：完整的 MySQL DDL 建表脚本。请以此为准在本地或 CI 环境执行建表/迁移；如需转换到 Flyway/Go migrate，可参考 `docs/migration_instructions_for_dba.md`。

## 首页/分类联调验收（wx-fe）

用于快速验证首页与分类的链路就绪（后端 9292、H5 预览 10088、门店/商品接口可用），并生成验收摘要。

- 一键联调验收：

```bash
bash scripts/wx-fe_link_check.sh
```

- 执行效果：
  - 自动启动 `tea-api`（端口 9292）并获取令牌
  - 启动 `wx-fe` H5 预览（默认端口 10088）并等待就绪
  - 以令牌调用 `GET /api/v1/stores` 与 `GET /api/v1/products`，输出门店与商品的计数摘要

- 验收摘要产物：
  - 最新摘要：见 [build-ci-logs/wx-fe-link-check/summary-latest.txt](build-ci-logs/wx-fe-link-check/summary-latest.txt)
  - 预览日志：位于脚本执行时的 `/tmp/wx-fe-link-check.<timestamp>/h5-preview.out`

- 停止 H5 预览：

```bash
kill $(cat /tmp/wx-fe-link-check.*/*h5-preview.pid)
```

说明：脚本在检测到工作区日志目录不可写时，会自动将临时产物落在 `/tmp/wx-fe-link-check.<timestamp>/`，同时在工作区 `build-ci-logs/wx-fe-link-check/summary-latest.txt` 复制一份摘要，便于归档。

### wx-fe 特性摘要（近期）

- 扫码进店（最小版）已实现：在 [wx-fe/src/app.tsx](wx-fe/src/app.tsx) 解析入口参数 `store_id`（query/scene/H5 URL），首页 [wx-fe/src/pages/index/index.tsx](wx-fe/src/pages/index/index.tsx) 与分类页 [wx-fe/src/pages/category/index.tsx](wx-fe/src/pages/category/index.tsx) 接管路由入参并持久化 `current_store_id`，商品详情页 [wx-fe/src/pages/product-detail/index.tsx](wx-fe/src/pages/product-detail/index.tsx) 亦在存在 `store_id` 时持久化；后续页面（订单列表/详情、购物车、结算）按门店维度展示与过滤。
- 当前门店徽标展示已覆盖：首页、商品详情、购物车、结算、订单详情、订单列表，保证门店上下文的用户可见性与一致性。
- 门店详情（最小版）已实现：在 [wx-fe/src/pages/store-detail/index.tsx](wx-fe/src/pages/store-detail/index.tsx) 支持展示证照与基础信息、地图导航与电话拨打、并跳转到分类页查看本店商品（携带 `store_id` 进行门店商品映射）。入口位于首页门店卡片“查看详情”按钮或门店列表页。
