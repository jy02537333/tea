# 计息功能说明

## 功能概述
- 按日为用户余额计息，记录每笔计息明细，支持复利。
- 管理端可手动触发计息、导出记录、查看汇总。
- 支持每日定时自动计息（可选 Redis 分布式锁）。
- 金额与利率采用定点 decimal 精度，避免浮点误差。

## 数据模型
- User.Balance: decimal(12,2)
- User.InterestRate: decimal(8,6)，当 > 0 时覆盖默认日利率（按账号差异化）
- InterestRecord: principal_before, rate, interest_amount, principal_after 均为 decimal（分别保留 2 位与 6 位精度）。
- InterestRecord 具备 (user_id, date) 复合唯一索引，保障同日同用户只记一次（并发幂等）。

## 接口
- 手动触发计息（需 accrual 权限）
  - POST /api/v1/admin/accrual/run
  - Body: { "date":"YYYY-MM-DD", "rate":0.001, "user_id":123? }
  - 返回: { updated: N }
- 导出计息记录（CSV/XLSX）
  - GET /api/v1/admin/accrual/export?start=YYYY-MM-DD&end=YYYY-MM-DD&user_id=&format=csv|xlsx&lang=zh|en&fields=a,b,c&zip=0|1
  - format: 输出格式，默认 csv；xlsx 需依赖 excelize。
  - lang: 表头语言，默认 zh（支持 zh/en）。
  - fields: 选择导出字段，默认全部（user_id,date,principal_before,rate,interest_amount,principal_after,method）。
  - zip: 当为 1 且 format=xlsx 时，打包 zip 下载。
- 汇总统计
  - GET /api/v1/admin/accrual/summary?start=YYYY-MM-DD&end=YYYY-MM-DD
- 用户查询自身计息记录（需登录）
  - GET /api/v1/user/interest-records?page=1&limit=20

## 权限
- 路由与权限拆分：
  - `/api/v1/admin/users` 仅 admin。
  - 计息相关路由通过权限控制（支持非 admin 但授予权限的角色访问）：
    - POST `/api/v1/admin/accrual/run` 需要 `accrual:run`
    - GET  `/api/v1/admin/accrual/export` 需要 `accrual:export`
    - GET  `/api/v1/admin/accrual/summary` 需要 `accrual:summary`
- 权限校验顺序：
  1) admin 直通。
  2) 数据库鉴权（User → UserRole → RolePermission → Permission(name)）。
  3) 回退配置 `finance.accrual.allowed_roles`（兼容旧逻辑）。
- 初始化建议：
  - 在 `permissions` 写入 `accrual:run|export|summary` 等权限。
  - 在 `roles` 创建需要的角色（如 auditor）。
  - 通过 `role_permissions` / `user_roles` 建立授权关系。

## 定时调度
- 配置位置（可选）：finance.accrual
  - enabled: true/false
  - time: "HH:MM"（24 小时制）
  - rate: 日利率（如 0.001）
  - timezone: 时区（如 Asia/Shanghai）
  - skip_weekends: 跳过周末
  - holidays: ["YYYY-MM-DD", ...] 节假日列表
  - use_redis_lock: true/false（建议开启，避免多实例重复执行）
  - lock_ttl_second: 锁过期时间，默认 3600
- 服务启动后自动读取配置并启动调度（若未启用则忽略）。

## 注意事项
- 若 Redis 不可用且启用调度，将无分布式锁，只能靠单实例运行保证不重复。
- 生产环境建议：
  - 打开 use_redis_lock
  - 指定准确的执行时刻（如 02:00）
  - 保持时区与服务器一致（默认使用系统时区）

## 示例

1) 只对某个用户、以 0.5%/日 计提（若该用户配置了 InterestRate>0，将使用其利率覆盖请求/默认值）：

POST /api/v1/admin/accrual/run
Body:
{
  "date": "2025-11-12",
  "rate": 0.005,
  "user_id": 123
}

2) 导出英文表头的 XLSX，仅导出指定字段并打包为 zip：

GET /api/v1/admin/accrual/export?start=2025-11-01&end=2025-11-12&format=xlsx&lang=en&fields=user_id,date,interest_amount&zip=1

3) 汇总统计（按时间范围）：

GET /api/v1/admin/accrual/summary?start=2025-11-01&end=2025-11-12
