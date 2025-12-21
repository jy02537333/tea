# PR 概述

简述本次变更的业务背景与目标。

## 主要变更
- 受影响模块：
- 关键路径：

## 风险与影响
- 文档链接有效性、CI 工作流执行与潜在回归。

## 验证步骤
- 本地按最小集成说明启动（见 docs/ci/minimal-integration.md）。
- 点击下方尾注中的 CI 文档确认可用。
- 合并后检查尾注链接是否自动转换为 master 相对路径且仍可用。

Status: Ready for review

---

### 评审速览 Checklist
- 变更范围：以 “Files changed” 页面为准，关注受影响模块与关键路径。
- 风险点：文档链接有效性、CI 工作流执行、核心流程潜在回归。
- 验证步骤：
  - 本地拉取分支并按最小集成说明启动（参考 docs/ci/minimal-integration.md）
  - 点击 PR 尾注中的 CI 文档并确认可用
  - 合并后确认尾注链接已自动转换为 master 相对路径且仍可用

#### 自动化说明
- 工作流：Attach Review Checklist（.github/workflows/attach-review-checklist.yml）。
- 触发：PR 从草稿切为可评审，或添加标签 “Ready for review”。
- 去重：若评论中已存在“评审速览 Checklist”，工作流不会重复发布。
- 运行链接示例：`https://github.com/<owner>/<repo>/pull/<number>/checks`

---

### PR 描述固定尾注
- CI 临时开启严格模式指引：docs/ci/minimal-integration.md（见文档中的“临时开启严格模式（CI）”小节）
- 更多 CI 与联调指南：CONTRIBUTING.md
