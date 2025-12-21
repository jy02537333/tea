# 评审速览 Checklist

用于加速代码评审的轻量清单，建议在标记 Ready for review 时自动附加到 PR 评论。

- 变更范围：以 “Files changed” 页面为准，关注受影响模块与关键路径。
- 风险点：文档链接有效性、CI 工作流执行、核心流程潜在回归。
- 验证步骤：
  - 本地拉取分支并按最小集成说明启动（参考 [docs/ci/minimal-integration.md](minimal-integration.md)）
  - 点击 PR 尾注中的 CI 文档并确认可用
  - 合并后确认尾注链接已自动转换为 master 相对路径且仍可用
- 额外检查（可选）：
  - 必要日志/告警是否完整（见项目 README 与运维指引）
  - 与相邻模块/接口的兼容性（如请求/响应字段、鉴权）
- 当前状态：Ready for review

说明：本清单为评审辅助材料，实际以 PR 描述、变更文件与 CI 结果为准。