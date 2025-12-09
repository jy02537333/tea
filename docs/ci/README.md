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
