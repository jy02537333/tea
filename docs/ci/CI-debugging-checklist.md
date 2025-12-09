# CI 调试清单（GitHub Actions / GitLab CI）

本文针对在 CI 上运行 `scripts/run_api_validation.sh` / `scripts/assert_api_validation.sh` 时常见无法执行或失败的情况，给出逐步手动调试方法与可执行的命令。适用于仓库默认 workflow（`.github/workflows/api-validation.yml`）和 `docs/ci/api-validation.gitlab-ci.yml` 模板。

---

## 一、快速概述（先读这部分）

- 我能在仓库里添加/修改 CI 配置（已完成），但无法直接在你的 CI 帐号/runner 上触发 job；你需要把分支 push 到远程或在 Actions 页面点击 `Run workflow` 来触发。若 job 在你的 CI 上失败或未启动，按下文排查。
- 常见失败点：runner 不支持 services、后端未能成功启动、依赖服务未就绪、断言脚本找不到 token 或请求返回非 200。

---

## 二、在本地复现（首选，可快速定位问题）

推荐先在本地或能使用 Docker 的主机上复现 CI 流程：

1. 使用 `docker-compose` 启动依赖（仓库有 `docker-compose.yml` 的话优先用它）：

```bash
# 启动依赖服务（mysql/redis/rabbitmq）和网络
docker-compose up -d
# 查看日志
docker-compose logs -f
```

2. 在本地构建并运行 `tea-api`：

```bash
cd tea-api
go build -o tea-api ./...
# 后台运行并把 stdout/stderr 重定向到文件，便于排查
./tea-api > ../build-ci-logs/tea-api.log 2>&1 &
# 等待端口
for i in {1..30}; do nc -z localhost 9292 && break || sleep 2; done
```

3. 运行 seeder 和验证脚本：

```bash
# (在 tea-api 或仓库根，根据脚本路径)
cd tea-api
go run scripts/seed_skus.go -product 1 -sku_name "默认规格" -sku_code "SKU-P1-001" -price "99.00" -stock 20 || true
cd ..
sh scripts/run_api_validation.sh
bash scripts/assert_api_validation.sh
```

4. 查看输出与 artifacts：

```bash
ls -la build-ci-logs/api_validation/
cat build-ci-logs/api_validation/summary.txt
cat build-ci-logs/admin_login_response.json | jq .
tail -n 200 build-ci-logs/tea-api.log
```

---

## 三、在 GitHub Actions 上调试（手动触发 & 日志位置）

1. 触发 workflow：在 GitHub 仓库 -> Actions -> 选择 `API Validation` workflow -> `Run workflow`（或 push 到分支触发）。
2. 打开失败的 run，查看每一步的日志（点击 step 名称）。重要 step：
  - `Build tea-api`（是否成功 `go build`）
  - `Start tea-api (background)`（是否报错，是否等待成功）
  - `Run API validation script`（输出响应位置）
  - `Run API assertion checks`（断言失败会显示原因）
3. 下载 artifacts：Actions 页面 -> 失败或成功的 run -> `Artifacts` 下拉 -> 下载 `api-validation-logs`。
4. 若想在 workflow 中增加更多 debug，请临时在 workflow 中加入：

```yaml
- name: Dump tea-api stdout
  if: always()
  run: |
    echo '=== tea-api log ==='
    cat build-ci-logs/tea-api.log || true

- name: Dump mysql logs (if available)
  if: always()
  run: |
    docker logs ${{ job.services.mysql.id }} || true
```

注意：`job.services.mysql.id` 在 Actions 中并非总可用，runner 类型不同可能需替代方法。

---

## 四、在 GitLab CI 上调试（手动重跑 & artifacts）

1. 在 GitLab Pipeline 页面找到对应 job，点击 `Play`（手动重跑）或重新运行 pipeline。
2. 查看 job 日志（控制台输出），关注 `go build`、`./tea-api` 启动、`sh scripts/run_api_validation.sh` 的输出路径。
3. 下载 artifacts：在 job 页面找到 `Artifacts` 并下载 `build-ci-logs` 文件夹。

---

## 五、如果 job 一直卡在“等待服务”或超时

- 检查 runner 是否支持 Docker services：
  - GitHub Actions shared runners 支持 services；自托管 runner 可能需要 Docker 启用。若 runner 不支持，请使用自托管且已安装 Docker 的 runner，或改为使用 `docker-compose`。
- 增加等待重试：在 workflow 中延长等待循环次数或 sleep 时间，例如从 30 次改为 60 次或增加间隔。示例：

```bash
for i in {1..60}; do nc -z mysql 3306 && nc -z redis 6379 && nc -z rabbitmq 5672 && break || sleep 3; done
```

---

## 六、断言脚本失败的具体排查（`scripts/assert_api_validation.sh`）

1. 常见失败原因：
  - `build-ci-logs/admin_login_response.json` 不存在或格式不同（脚本从 `.data.token` / `.token` / `.access_token` 提取）
  - token 无效导致 api 返回 401/403
  - 某个 admin endpoint 返回 500/404/其它非 200
2. 调试步骤：

```bash
# 检查文件是否存在
ls -la build-ci-logs/admin_login_response.json
# 打印 JSON
cat build-ci-logs/admin_login_response.json | jq .
# 手动用 token 检查某个 endpoint
TOKEN=$(jq -r '.data.token // .token // .access_token' build-ci-logs/admin_login_response.json)
curl -i -H "Authorization: Bearer $TOKEN" http://127.0.0.1:9292/api/v1/admin/orders
```

3. 如果 token 字段名不同，请在 `scripts/assert_api_validation.sh` 中添加或调整 jq 路径，或在 `sh scripts/run_api_validation.sh` 中确保 admin dev-login 响应保存到 `build-ci-logs/admin_login_response.json` 并包含标准字段。

---

## 七、当构建失败（go build）时

- 查看完整 build 输出（Actions/GitLab job 日志里会有），常见问题：
  - 缺少系统依赖（apt 安装必要工具）
  - CGO/权限/环境变量导致的编译差异
- 可在 workflow 中临时增加：

```yaml
- name: Show go env
  run: go env
- name: Show modules
  run: go list -m all
```

---

## 八、资源限制与超时

- 若 runner 报内存/磁盘不足：在 job 配置中使用更大 runner（self-hosted）或拆分 job 减少并发。
- 数据库存储问题：确保 MySQL 的卷/磁盘空间足够，或使用轻量级内存数据库做 CI 测试（需后端支持）。

---

## 九、当服务不可达（端口/绑定问题）

- 确认 `tea-api` 绑定地址为 `0.0.0.0:9292`（容器中需要绑定 0.0.0.0）。如果只绑定 `127.0.0.1`，外部服务与 runner 无法访问。
- 在启动命令中显式传入绑定地址或配置文件：

```bash
# 示例（假如 tea-api 支持 --bind 参数）
./tea-api --bind 0.0.0.0:9292 > ../build-ci-logs/tea-api.log 2>&1 &
```

---

## 十、如果 runner 不支持 services，如何改用 docker-compose（推荐自托管 runner）

1. 在 workflow 中用 `docker` 执行 `docker-compose`：

```yaml
- name: Start stack with docker-compose
  run: |
    docker-compose -f docker-compose.yml up -d
    sleep 10
```

2. 要求 runner 有 Docker 权限（在 GitHub Actions 中需要 self-hosted runner 或使用 `docker` 扩展），或在 CI job 中用 `docker run` 手动启动依赖容器。

---

## 十一、快速 checklist（每次 CI 失败时按顺序执行）

1. 在 Actions/GitLab 页面下载 artifacts（若有），打开 `build-ci-logs/` 查看 `summary.txt` 与 `admin_login_response.json`。
2. 检查 `tea-api` 启动日志（`build-ci-logs/tea-api.log`）是否有 panic / 绑定错误。
3. 确认依赖服务（mysql/redis/rabbitmq）状态与连接（runner 日志或增加 docker logs）。
4. 手动用 `curl` 和 `token` 验证关键接口，排查 401/403/500/404 等错误。
5. 若需要更详细信息，在 workflow 中临时把关键目录 `build-ci-logs` 的全部输出 `cat` 出来或延长等待时间。

---

## 十二、我可以为你做的具体修改（可选，告诉我你要哪项）

- 把断言 endpoints 抽到 `docs/ci/assert_endpoints.json` 并让脚本读配置（便于不同环境复用）。
- 在 workflow 中把 `tea-api` stdout/stderr 写入 `build-ci-logs/tea-api.log` 并上传（我可以改现有 workflow）。
- 增加 `fail_on_assertion` 可选输入，使手动触发时可选择忽略断言。

---

如果你同意，我现在会：

- 把 `docs/ci/CI-debugging-checklist.md` 写入仓库（已完成），
- 或按你的选择继续做上面列出的任一可选修改（例如把断言 endpoints 抽成配置或把 `tea-api` 日志保存为 artifact）。

告诉我你接下来想让我做哪件（或已足够）。