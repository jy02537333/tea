# 最小集成执行说明

> 目的：提供本地与 CI 的最小可执行路径，用尽量少的步骤串联后台商品与分销解冻链路，并将证据统一归档到 build-ci-logs/。

## 本地一键执行
- 推荐使用 Makefile 目标：

```bash
make run-min-integration
```

- 等价分步执行：

```bash
bash run-tea-api.sh
bash scripts/local_api_check.sh    # 可选，生成用户态最小证据
bash scripts/run_admin_product_min.sh
bash scripts/run_commission_min.sh
```

- 产物位置：build-ci-logs/（包含 api_validation/**、api_validation_stateful/**）。摘要见 build-ci-logs/local_api_summary.txt。

## CI 自动执行
- 工作流：.github/workflows/minimal-integration.yml
  - 在 push 到 feat/withdraw-remark-json-ui-docs 或对 master 的 PR 时运行。
  - 使用内置服务启动 MySQL/Redis，安装 Go 与 jq，执行上述最小脚本。
  - 将 build-ci-logs/**（含子目录）上传为 Actions Artifacts，可在任务详情的 Artifacts 区下载查看证据文件。

## 常用环境变量
- API_BASE：后端地址（默认 http://127.0.0.1:9292）。
- ADMIN_TOKEN：管理员 JWT（未提供时脚本会尝试从日志发现或使用本地登录 admin/pass）。
- ORDER_ID：分销脚本的订单 ID（未提供时会尝试创建并自动发现）。
- TEA_AUTO_MIGRATE：run-tea-api.sh 默认 0；如需启动时自动迁移，设置为 1。
- TEA_JWT_SECRET：后端 JWT 密钥（默认 tea-shop-jwt-secret-key-2023，或按需覆盖）。

## 排障提示
- 端口占用：run-tea-api.sh 会尝试释放 9292；如失败，请手动清理后再运行。
- 缺少 jq：请先安装 jq（CI 会自动安装，本地可通过包管理器安装）。
- OSS 未配置：公开上传步骤若未返回 url，脚本会记录告警并跳过图片回填，不影响最小链路断言。

## 参考
- Makefile 目标：run-min-integration（串联最小链路，产物写入 build-ci-logs/）。
- 脚本：scripts/run_admin_product_min.sh、scripts/run_commission_min.sh。
