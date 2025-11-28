---
title: "CI: 增加 `go test ./...` 阶段（或使用 scripts/run-tests-test2.ps1）"
labels: "ci, backend, automation"
assignees: "jy02537333"
---

描述：在 CI pipeline 中加入后端单元测试阶段，使用 `scripts/run-tests-test2.ps1`（Windows runner）或等价的 Linux 脚本执行 `go test ./...`。

验收标准：
- CI 能在 PR 提交时执行单元测试并返回通过/失败状态；
- 如仓库支持 Windows runner，可直接使用现有 PowerShell 脚本；否则提供跨平台替代脚本。

建议负责人：后端/DevOps
预计工作量：1-2 人日
