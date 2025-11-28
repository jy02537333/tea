# A2 仪表盘（Dashboard）

**目标**：实现 KPI 卡片、主要汇总图表与快速操作按钮

**路由**：`#/dashboard`

**关键接口**：`GET /api/v1/admin/accrual/summary`、（可选）其它统计接口

**组件**：`Card`, `Statistic`, `Chart`（ECharts）

**验收标准**：
- KPI 卡片显示来自 `accrual/summary` 的数值
- 点击“触发计提”弹出确认并调用 `POST /api/v1/admin/accrual/run`，返回成功消息

**估时**：1d
