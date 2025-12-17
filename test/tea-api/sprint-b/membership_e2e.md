# Sprint B — 会员购买与权益验证 E2E（占位）

目标：登录→购买会员→支付回调→权益发放→前端展示验证。

步骤草案：

- 发起 `POST /api/v1/auth/login` 获取 token
- `POST /api/v1/membership/purchase` 创建订单，记录 idempotency-key
- 模拟支付回调（待接入）→ 校验用户等级与权益发放
- `GET /api/v1/users/me/summary` 校验聚合数据字段
