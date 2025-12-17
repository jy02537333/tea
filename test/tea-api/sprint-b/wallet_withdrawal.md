# Sprint B — 钱包提现全链路测试（占位）

目标：申请→风控校验→审批→回款，校验流水与状态机。

步骤草案：

- `GET /api/v1/wallet` 与 `GET /api/v1/wallet/transactions` 校验初始余额
- `POST /api/v1/wallet/bank-accounts` 绑定提现账户
- `POST /api/v1/users/{user_id}/withdrawals` 发起提现（带 Idempotency-Key）
- `GET /api/v1/admin/withdrawals` 管理员审批通过/拒绝
- 再次 `GET /api/v1/wallet/transactions` 校验流水记录
