# A9 报表 / 计提（Accruals / Exports）

**目标**：报表筛选、展示、导出与触发计提动作

**路由**：`#/accrual`

**关键接口**：`GET /api/v1/admin/accrual/summary`、`GET /api/v1/admin/accrual/export`、`POST /api/v1/admin/accrual/run`

**验收标准**：
- 支持多种导出格式（CSV/XLSX/ZIP）并正确下载
- 触发计提后给出 `updated` 数值提示

**估时**：1.5d
