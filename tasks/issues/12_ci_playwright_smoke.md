---
title: "CI: 将 Playwright 登录 smoke 脚本集成为可选 CI job"
labels: "ci, frontend, automation"
assignees: "jy02537333"
---

描述：把 `tools/browser-login/login-test.js` 集成为 CI pipeline 的可选 smoke job，运行时产出 JSON 报告并避免在报告中泄露 token。

验收标准：
- 在 CI 中增加一个可选 Job（例如 `smoke-test`），执行 Playwright 脚本；
- 生成 `report-*.json` 并上传为 job artifact；
- 报告中敏感 token 字段要 redacted 或不包含实际 token 值。

建议负责人：DevOps + 前端工程师
预计工作量：2-3 人日
