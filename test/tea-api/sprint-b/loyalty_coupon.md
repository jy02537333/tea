# Sprint B — 积分/优惠券功能测试（占位）

目标：获取→下单抵扣→状态校验，覆盖可用券筛选与积分流水。

步骤草案：

- `GET /api/v1/coupons/templates` 列出模板
- `POST /api/v1/coupons/claim` 领取用户券
- `GET /api/v1/coupons` 校验状态分类（待用/已用/过期）
- 下单流程中注入可用券列表并抵扣（接口对齐 Sprint A 下单）
- `GET /api/v1/points` 与 `GET /api/v1/points/transactions` 校验积分变更
