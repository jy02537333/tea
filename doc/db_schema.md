# 茶心阁 小程序 — 数据库设计（ER 概览 + 说明）

说明：本设计以 MySQL（InnoDB，utf8mb4）为目标，覆盖需求文档中列出的所有功能点：用户/会员/合伙人、商品/SKU/库存、门店/门店商品、购物车、订单/支付、分销/佣金、优惠券/积分、钱包/提现、后台与门店后台相关表、打印与通知队列等。

目录：
- 概览和设计原则
- 主要实体与关系（表清单与字段要点）
- ER 关系简述
- 使用说明与后续工作
- SQL DDL 文件：`../db/schema.sql`

## 设计原则
- 以事务一致性为主，财务类操作（钱包、佣金、提现）使用独立流水表并保证幂等。
- 对订单、商品、库存等高并发表使用行级锁或乐观锁策略（业务代码实现库存预占）。
- 所有涉及金额、数量的字段使用整数存储最小单位（分、最小库存单位）。
- 时间字段统一使用 UTC（DATETIME），并在应用层展示本地时区。

## 主要表（概要）
下面列出主要表和作用，详细建表请参考 `db/schema.sql`。

- `users`：用户主体（消费者/合伙人/管理员均是用户）
- `user_profiles`：用户扩展信息（昵称、头像、手机号等）
- `roles` / `user_roles`：权限角色与分配
- `stores`：门店信息
- `store_staff`：门店员工与角色
- `categories`：商品分类
- `products`：商品主表（SPU）
- `product_skus`：SKU（规格）与价格、库存基线
- `product_images`：商品图片
- `store_products`：门店特供/门店售卖的商品映射（价格/是否外卖/堂食）
- `inventories`：库存实时/可用/预占记录（按 sku / store）
- `cart_items`：购物车项（可多处保存，支持门店/商城区分）
- `addresses`：用户收货地址
- `orders`：订单主表（支持类型：mall/store/delivery/积分订单）
- `order_items`：订单明细
- `payments`：支付流水（与外部渠道回调对接）
- `coupons_templates`：优惠券模板
- `coupons`：用户领取的优惠券实例
- `points_transactions`：积分流水
- `wallets`：用户钱包（余额/冻结）与 `wallet_transactions` 流水
- `membership_packages`：会员/合伙人礼包配置
- `user_memberships`：用户购买的会员/合伙人记录
- `referrals`：推荐关系（上级-下级）
- `commissions`：佣金冻结/可用/已提现记录
- `commission_transactions`：佣金发放/提现流水
- `print_jobs`：门店打印任务队列
- `notifications`：消息/模板推送记录
- `admin_audit_logs`：关键操作审计日志
- `shipping_templates` / `shipping_rules`：运费模版
- `activities` / `activity_registrations`：活动/报名管理
 - `permissions` / `role_permissions`：细粒度权限与角色-权限映射
 - `partner_levels`：合伙人等级策略（拿货折扣、佣金比例等）
 - `user_bank_accounts`：用户提现/收款账户（银行卡/支付宝/微信）
 - `withdrawal_requests`：提现申请记录（含发票/手续费字段）
 - `referrals_closure`：推荐闭包表，支持多层分销计算

## ER 简述（核心关系）
- `users` 1 — N `orders`
- `users` 1 — N `wallet_transactions` / `points_transactions` / `commissions`
- `products` 1 — N `product_skus`
- `stores` 1 — N `store_products`
- `orders` 1 — N `order_items` (each item references `product_skus`)
- `users` N — 1 `referrals`（被推荐人记录上级）

## 使用说明与后续工作
- 请后端与 DBA 评审表字段长度、索引与分区策略（若需要可基于时间分区 `orders`）。
- 我可以把 `db/schema.sql` 转成 Flyway/Go migrate 或者生成 ORM 模型（GORM / ent / sqlc）。

说明：
- `permissions` 与 `role_permissions` 用于把角色拆成可管理的权限项，便于后台菜单与接口级别控制。
- `partner_levels` 单独维护合伙人等级策略（拿货折扣、直推/团队佣金、升级奖励），`membership_packages` 可引用或映射到等级。
- `user_bank_accounts` 与 `withdrawal_requests` 支持完整提现流程（含手续费、发票要求、审核/打款时间），满足 PRD 中的财务合规要求。
- `referrals_closure` 使用闭包表以支持 N 层分销佣金统计（避免递归查询带来的性能问题）。

详细建表语句见：`db/schema.sql`。
