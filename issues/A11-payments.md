# A11 财务/退款/提现（Payments/Refunds/Withdraws）

**目标**：财务记录与提现审批工作流（列表/导出/审核动作）

**路由**：`#/finance`, `#/refunds`, `#/withdraws`

**关键接口**：`GET /api/v1/admin/payments`、`GET /api/v1/admin/refunds`、`GET /api/v1/admin/withdraws`、审核 POST `/api/v1/admin/withdraws/:id/approve|complete|reject`

**验收标准**：审批操作后状态变更显示并能导出记录

**估时**：2d
