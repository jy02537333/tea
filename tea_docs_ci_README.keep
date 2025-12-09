# API Validation CI Templates

This folder contains CI job templates to run the repository's `scripts/run_api_validation.sh` in CI and collect artifacts.

Provided files:

- `.github/workflows/api-validation.yml` — GitHub Actions workflow that: checks out the repo, starts MySQL/Redis/RabbitMQ as services, builds and runs `tea-api`, seeds example SKU, runs `scripts/run_api_validation.sh`, and uploads `build-ci-logs` as artifacts.

- `api-validation.gitlab-ci.yml` — GitLab CI job template with similar steps (services: mysql/redis/rabbitmq). Use it in your project by including or copying into `.gitlab-ci.yml`.

How it works (summary):

1. Start dependent services (mysql/redis/rabbitmq) via CI-provided services.
2. Build and run `tea-api` in background on port `9292`.
3. Run seeder steps (example SKU creation) so validation has minimal data.
4. Run `sh scripts/run_api_validation.sh` which performs HTTP requests and writes responses under `build-ci-logs/api_validation/` and `build-ci-logs/admin_login_response.json`.
5. Upload `build-ci-logs` as CI artifacts for QA review.

Notes & tips:

- If your CI runner doesn't support Docker services, use a self-hosted runner with Docker and run `docker-compose up -d` instead.
- Adjust seeder steps in the workflow to match your required test data.
- Consider adding a short smoke test assertion that checks `build-ci-logs/api_validation/summary.txt` for expected 200 responses to fail the job when regressions occur.

## 移除 Git 历史中的大文件（推送被 100MB 限制阻断时）

- 说明：GitHub 单文件限制 100MB。若历史中存在超限大文件，会导致推送失败（GH001）。可使用脚本重写历史移除超大 blob。
- 先决条件：已安装 `git-filter-repo`（推荐 `pipx install git-filter-repo` 或 `pip3 install --user git-filter-repo`）。参考：<https://github.com/newren/git-filter-repo>
- 风险提示：历史重写是破坏性操作，请先与团队确认；推送后协作者需要重新克隆或硬重置到新历史。
- 使用步骤：
  - 赋予脚本执行权限：`chmod +x scripts/remove-large-files.sh`
  - 在仓库根目录运行：`scripts/remove-large-files.sh`（默认阈值 100M，可传参自定义：`scripts/remove-large-files.sh 120M`）
  - 完成后按提示安全推送当前分支：`git push --force-with-lease origin $(git rev-parse --abbrev-ref HEAD)`
  - 建议将产生大文件的路径加入 `.gitignore`，避免再次提交。

## 维护工具用途摘要（Repository Hygiene Toolkit）

- `scripts/remove-large-files.sh`: 当前分支历史大文件清理（剥离 `>100M` blob），用于解除 GitHub 100MB push 阻断；缺少 `git-filter-repo` 时会自动下载到 `~/.local/bin`；执行后需 `--force-with-lease` 推送并通知协作者对齐历史。
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
