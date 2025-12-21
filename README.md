# API Validation CI Templates

> 主线门禁：`master` 分支已启用分支保护，合并必须通过 “API Validation” 工作流（Sprint A 阻塞、Sprint B 非阻塞）。详见 `doc/prd.md` 与 `doc/prd_sprints.md`。

[![API Validation](https://github.com/jy02537333/tea/actions/workflows/api-validation.yml/badge.svg?branch=master)](https://github.com/jy02537333/tea/actions/workflows/api-validation.yml)

This folder contains CI job templates to run the repository's `scripts/run_api_validation.sh` in CI and collect artifacts.

## 版本标签（Releases / Tags）

- `v0.1-a-first`：启用“主线门禁（A-first）”，将 Sprint A 作为阻断检查（严格断言：订单金额校验 + 回调签名），Sprint B 为非阻断，仅归档证据。此标签定位于 `master` 当前稳定点，便于回溯证据文件（位于 `build-ci-logs/`，由 CI 工件归档）。
  - 查看代码：`https://github.com/jy02537333/tea/tree/v0.1-a-first`
  - 查看工作流运行与工件：`https://github.com/jy02537333/tea/actions/workflows/api-validation.yml`

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

## 快速联调（Minimal Integration）

- 文档说明：见 [docs/ci/minimal-integration.md](docs/ci/minimal-integration.md)
- 本地一键执行：

```bash
make run-min-integration
```

- 严格模式示例（可选，默认关闭）：

```bash
STRICT_MIN=1 make run-min-integration
```

- CI 自动执行：工作流 [minimal-integration.yml](.github/workflows/minimal-integration.yml) 会在 push/PR 时运行并将 `build-ci-logs/**` 上传为 Artifacts。
 - 贡献者说明：更多 CI 与联调指南见 [CONTRIBUTING.md](CONTRIBUTING.md)

## 移除 Git 历史中的大文件（推送被 100MB 限制阻断时）

- 使用步骤：
## 本地开发端口约定

- 后端 API 默认在 `:9292` 端口监听，基础地址为：`http://localhost:9292/api/v1`
- 请将前端与工具脚本（如 `API_BASE`、`VITE_API_BASE_URL`、`WX_API_BASE_URL` 等）统一指向 `http://localhost:9292`，避免误连到其他本机服务（如 `:8080`）。
- 如需变更端口，可在 `tea-api/configs/config.yaml` 或对应本地配置中调整，并同步更新前端环境变量。

### 保持 9292 服务可用的运维提示

- 遇到 `listen tcp :9292: bind: address already in use`，说明 9292 已被占用。
- 请优先确保 `tea-api` 独占 9292，按照下列步骤处理：

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

### 数据库设计文档（新增索引）

- `doc/db_schema.md`：数据库表结构设计概览与说明（ER 概览、主要实体与关系、设计原则），用于后端与 DBA 对齐字段与关系。
- `db/schema.sql`：完整的 MySQL DDL 建表脚本。请以此为准在本地或 CI 环境执行建表/迁移；如需转换到 Flyway/Go migrate，可参考 `docs/migration_instructions_for_dba.md`。
