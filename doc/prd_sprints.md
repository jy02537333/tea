# 茶心阁 小程序 — Sprint 任务拆解与接口定义

说明：本文件将 `doc/prd.md` 中的里程碑进一步拆解为每个 Sprint 的具体开发任务（前端/后端/测试/运维），并提供关键接口（REST 风格）定义草案，便于前后端对齐与编写 API 文档。

> 补充：涉及具体字段、数据库表结构、订单状态机、详细 API 入参/出参时，请结合 `doc/prd.md` 中的「四、关联文档索引」查阅对应的技术文档（如 `doc/db_schema.md`、`tea-api/docs/api-*.md`、`docs/prd-open-points.md` 等），以保证 Sprint 任务与最新设计保持一致。

---

## Sprint A（第1-3周） — 核心下单能力

目标：实现商品浏览、门店列表、购物车、下单与支付前端流程及后端基础接口。

- 前端任务
  - 首页与商品列表页面：分类筛选、分页、主图/价格显示
  - 商品详情页：多图、参数、加入购物车、分享按钮
  - 门店列表与门店详情：按距离排序、门店点单入口
  - 购物车页面：按来源（门店/商城）分组、数量修改、删除
  - 结算页：选择收货地址/门店、优惠券/茶币/积分选择、下单发起
  - 支付回调接入（小程序端的支付回调处理）
- 后端任务
  - 设计并实现商品、分类与门店基础表与接口
  - 购物车 API（临时保存在服务端或客户端）
  - 订单创建、支付流水记录、订单状态机（待付款、已支付、已发货/待取、完成）
  - 支付集成：微信支付统一下单接口与回调接收
- 测试任务
  - 核心路径 E2E：浏览->下单->支付->订单状态变更
  - 并发下单场景：库存预占与回滚

### Sprint A — 关键 API（草案）

- 用户鉴权说明：绝大多数用户相关或订单创建接口需要 `Authorization: Bearer <token>`（JWT 或 session token）。
  
 统一登录与聚合（JWT v5）：
 - 登录：`POST /api/v1/auth/login`（返回 `token` 与用户基础信息）。
 - 个人中心聚合：`GET /api/v1/users/me/summary`（需携带 `Authorization: Bearer <token>`）。
 - 联调与测试统一采用以上路径，避免与历史接口（如 `/api/v1/user/login`、`/api/v1/user/info`）混用。

- GET /api/v1/products
  - 功能：获取商品列表（分页/筛选）
  - Query：`category_id?`, `keyword?`, `page`, `size`, `sort?`（销量/价格/上新）
  - Auth：可选
  - Response: `{data:[{id,name,price,primary_image,labels}], total, page}`

- GET /api/v1/products/{id}
  - 功能：商品详情
  - Auth：可选
  - Response: `{id,name,price,images[],stock,description,attributes[],sku[]}`

- GET /api/v1/stores
  - 功能：门店列表，按距离排序
  - Query：`lat`, `lng`, `radius?`, `page`, `size`
  - Response: `{data:[{id,name,lat,lng,address,open_hours,phone,distance}], total}`

- GET /api/v1/stores/{id}
  - 功能：门店详情与门店商品分类
  - Response: `{id,name,menus:[{category,items:[{product_id,price,stock}]}],info}`

- POST /api/v1/cart
  - 功能：加入购物车
  - Body: `{user_id?, store_id?, product_id, sku_id?, quantity}`
  - Auth：推荐
  - Response: `{cart_id, items[]}`

- GET /api/v1/cart
  - 功能：获取用户购物车
  - Auth：需要
  - Response: `{items:[{cart_item_id,product_id,sku_id,qty,store_id,price}], totals}`

- POST /api/v1/orders
  - 功能：创建订单（支持门店点单/外卖/商城）
  - Body: `{user_id, items:[{product_id,sku_id,qty,store_id}], address_id?, delivery_type: "store"|"delivery", coupon_id?, use_points?, pay_method}`
  - Response: `{order_id, pay_info}`
  - 注意：下单时校验库存、预占库存、计算运费与优惠

- POST /api/v1/payments/unified-order
  - 功能：调用支付渠道创建预支付单（微信）
  - Body: `{order_id, pay_method: "wechat"}`
  - Response: `{prepay_id, pay_params}`

- POST /api/v1/payments/callback
  - 功能：支付渠道回调（微信支付回调）
  - Auth：公网回调无需用户身份，但需签名校验
  - Body: 渠道回调格式
  - Response: 返回渠道要求的格式（XML/JSON）并在本地完成订单状态更新

#### 下单与抵扣验证（本地记录）

- 订单：`id=156`，`order_no=O20251216183530c37821`
- 门店：`store_id=1`
- 金额：`total_amount=396`，`discount_amount=60`（`coupon_id=24`），`pay_amount=336`
- 严格校验：`pay_amount = total_amount - discount_amount`，代入 `396 - 60 = 336`，机器校验结果 `check=true`
- 证据文件：
  - `build-ci-logs/order_from_cart_store1_coupon24.json`
  - `build-ci-logs/order_amounts_summary.json`
  - `build-ci-logs/order_detail_156.json`
  - `build-ci-logs/order_detail_156_checked.json`

> 自动化与 CI 保护：
> - 本地/CI 状态化脚本：`scripts/local_api_check.sh` 会在登录后完成最小下单路径（购物车→下单→订单详情），并生成以上订单金额证据文件与 `build-ci-logs/local_api_summary.txt`。
> - 严格断言脚本：`scripts/assert_api_validation.sh` 读取 `order_detail_*_checked.json` / `order_amounts_summary.json` 等证据，严格校验 `pay_amount = total_amount - discount_amount`，`make verify-sprint-a[-strict]` 为统一入口。
> - CI 集成：`.github/workflows/api-validation.yml` 中的 `stateful-api-check` job 会在公共 API、自状态化检查后执行 `make verify-sprint-a-strict`，并将上述证据文件作为工件归档，用于回归与审计。

> 迭代策略（以 A 为主）：当前迭代将 Sprint A 作为主线与 CI 阻断标准；支付回调 ST（`POST /api/v1/payments/callback`）作为关键校验点，会在统一下单后通过模拟支付回调验证签名与订单状态流转。Sprint B 的检查作为“观察项”（非阻断），不影响合并结论。

> 主线门禁提示（master）：已启用分支保护并将 “API Validation” 设为必需状态检查（strict=true）。Sprint A 断言失败会阻断合并；Sprint B 为非阻断、仅归档证据。详见 `doc/prd.md` 的“主分支保护与 CI 门禁（master）”。


---

## Sprint B（第4-5周） — 用户与会员体系

目标：实现用户注册、登录、个人中心、钱包/茶币/积分/优惠券功能。支持会员开通购买流程。

- ### Sprint B — 任务板（含接口/前端/测试）

| 序号 | 模块 | 子任务 / 交付物 | 关键接口 / 代码位置 | 负责人 | 预计完成 | 完成情况 | 依赖 / 备注 |
| ---- | ---- | ---------------- | -------------------- | ------ | -------- | -------- | ------------ |
| B-01 | 后端 | 会员体系相关数据库迁移：新增/更新 `users` 扩展字段、`memberships`、`wallet_accounts`、`wallet_ledger`、`points_ledger`、`coupon_templates`、`user_coupons`、`user_bank_accounts`、`withdrawal_requests` | `db/migrations/*.sql`、`tea-api/pkg/database/migrate.go` | 陈伟（后端） | W4-D2 | [ ] | 需与 DBA 对齐字段命名及幂等策略 |
| B-02 | 后端 | 聚合个人中心数据服务：实现 `GET /api/v1/users/me/summary`，补充 service/cache 层 | `tea-api/internal/handler/user_summary.go`、`tea-api/internal/service/profile` | 陈伟（后端） | W4-D3 | [进行中] | 已接入 `SummaryDeps` 并返回钱包/积分/券/会员聚合；依赖 B-01 表结构完善 |
| B-03 | 后端 | 鉴权登录升级：支持手机号+验证码/微信登录合并，返回 token+用户信息 | `POST /api/v1/auth/login`、`tea-api/internal/handler/auth.go` | 刘敏（后端） | W4-D3 | [进行中] | 已提供占位校验与统一 JWT 签发，后续接入短信/微信网关 |
| B-04 | 后端 | 钱包接口：`GET /api/v1/wallet`、`GET /api/v1/wallet/transactions`、`POST /api/v1/wallet/withdrawals` 及风控校验 | `tea-api/internal/handler/wallet.go`、`tea-api/internal/service/wallet` | 刘敏（后端） | W4-D5 | [部分完成] | 新增用户钱包余额与流水只读接口；提现申请手续费（用户端）已完成；风控待接入 |
| B-05 | 后端 | 积分接口：查询、流水、可兑换商品列表；实现积分增减通用服务 | `GET /api/v1/points*`、`tea-api/internal/service/points` | 王磊（后端） | W4-D5 | [ ] | 与订单服务约定积分获取/消费事件 |
| B-06 | 后端 | 优惠券接口：模板列表、领取、用户券列表；下单可用券查询 RPC | `GET /api/v1/coupons*`、`POST /api/v1/coupons/claim`、`tea-api/internal/service/coupon` | 王磊（后端） | W5-D1 | [ ] | 依赖营销规则配置，提供错误码 |
| B-07 | 后端 | 提现账户管理与审批流：`/wallet/bank-accounts` CRUD、`/users/{id}/withdrawals`、`/admin/withdrawals` 审批 | `tea-api/internal/handler/withdrawal.go`、`tea-api/internal/service/withdrawal` | 陈伟（后端） | W5-D1 | [进行中] | 用户端手续费已完成；管理员 `paid/complete` 流已串联钱包扣减与手续费记账；`reject` 场景已实现解冻并返还可用余额；钱包流水 `remark` 统一为可解析 JSON；门店提现备注结构已统一为 JSON；前端/管理端列表已解析展示 remark JSON 字段；限额与幂等待补充 |
| B-08 | 后端 | 会员购买流程：`POST /api/v1/membership/purchase` -> 支付回调 -> 权益发放（茶币/优惠券/等级升级） | `tea-api/internal/handler/membership.go`、`tea-api/internal/service/membership` | 刘敏（后端） | W5-D2 | [部分完成] | 已提供套餐列表与创建会员订单；支付回调与权益发放待联调 |
| B-09 | 后端 | 分享/门店管理员接口：`GET /api/v1/users/me/share-stats`、`GET /api/v1/users/me/store-role` | `tea-api/internal/handler/profile_extra.go` | 王磊（后端） | W5-D2 | [ ] | 需从分销/门店模块获取数据 |
| B-10 | 前端 | 登录/注册页面逻辑（手机号验证码、微信授权）、Token 管理、中台登录态拦截 | `wx-fe/src/pages/auth/*`、`wx-fe/src/store/auth.ts` | 林晓（前端） | W4-D3 | [ ] | 依赖 B-03 登录接口 |
| B-11 | 前端 | 「我的」首页聚合：头像、会员等级、余额/积分/券数量、订单入口 | `wx-fe/src/pages/mine/index.tsx` | 林晓（前端） | W4-D4 | [ ] | 依赖 B-02 API；与 UI 对齐 |
| B-12 | 前端 | 钱包模块：余额详情、账单列表、提现申请、提现账户管理 | `wx-fe/src/pages/wallet/*`、`wx-fe/src/components/withdrawal/*` | 赵婷（前端） | W5-D1 | [ ] | 依赖 B-04、B-07 接口；需考虑多币种展示 |
| B-13 | 前端 | 积分与优惠券模块：积分记录、积分商品跳转、券列表及状态切换 | `wx-fe/src/pages/points/*`、`wx-fe/src/pages/coupons/*` | 赵婷（前端） | W5-D1 | [ ] | 依赖 B-05、B-06 接口 |
| B-14 | 前端 | 会员开通页：等级权益展示、购买流程、支付前后状态 | `wx-fe/src/pages/membership/*` | 林晓（前端） | W5-D2 | [ ] | 依赖 B-08 接口及支付 SDK |
| B-15 | 前端 | 客服/分享/设置子模块：意见反馈、推广分享页、门店管理入口、设置页 | `wx-fe/src/pages/support/*`、`wx-fe/src/pages/share/*`、`wx-fe/src/pages/settings/*` | 周岚（前端） | W5-D2 | [ ] | 依赖 B-09 接口 |
| B-16 | 测试 | 编写会员购买与权益验证 E2E：登录→购买→权益发放→会员中心展示 | `test/tea-api/sprint-b/membership_e2e.md`、`scripts/run_membership_integration.sh` | 朱洋（QA） | W5-D2 | [ ] | 依赖 B-08、B-11、B-14 |
| B-17 | 测试 | 钱包提现全链路测试：申请、风控、审批、回款 | `test/tea-api/sprint-b/wallet_withdrawal.md`、`scripts/run_membership_integration.sh` | 朱洋（QA） | W5-D2 | [ ] | 依赖 B-04、B-07、B-12 |
| B-18 | 测试 | 积分/优惠券功能测试：获取→下单抵扣→状态校验 | `test/tea-api/sprint-b/loyalty_coupon.md` | 朱洋（QA） | W5-D2 | [ ] | 依赖 B-05、B-06、B-13 |

- 前端任务
  - 登录/注册流程，手机号 + 验证码登录、微信一键登录
  - 我的页面：会员入口、余额/茶币/积分/优惠券入口
  - 会员开通页面：展示会员等级、权益、购买入口
  - 优惠券领取/使用流程
- 后端任务
  - 用户/会员表设计，钱包与积分流水表
  - 购买会员（支付流程）后续权益开通逻辑（折扣、茶币发放）
  - 优惠券系统：券模板、发放、核销规则
  - 积分规则接口与兑换接口
- 测试任务
  - 会员购买与权益生效验证
  - 优惠券在订单结算时生效并正确计算金额

> 自动化与 CI 保护（会员购买成功路径）：
> - 状态化脚本：`scripts/run_membership_integration.sh` 负责跑通会员套餐列表 → 创建会员订单 → 统一下单 → 模拟支付回调 → 查询会员订单 → 最后调用 `GET /api/v1/users/me/summary`，并将聚合视图快照与会员等级检查结果落盘为 `build-ci-logs/membership_summary_after_purchase.json`、`build-ci-logs/membership_b_flow_checked.json`。
> - 严格断言脚本：`scripts/assert_membership_flow.sh` 读取 `membership_b_flow_checked.json`，要求 `.ok == true`（例如会员等级从 visitor 升级为具体等级）；`make verify-sprint-b[-strict]` 提供本地与 CI 统一入口。
> - CI 集成：在 `.github/workflows/api-validation.yml` 的 `stateful-api-check` job 中，Sprint A 严格断言之后会执行上述会员脚本与 `make verify-sprint-b-strict`，并将相关 JSON 及日志文件作为工件归档，作为 Sprint B 会员成功路径的自动化保护。
>
> 注意：本迭代以 A 为主，Sprint B 在 CI 中为非阻断检查（失败不阻塞合并），用于提前暴露问题与收集证据，具体见 workflow 的 `continue-on-error` 设置。

### Sprint B — 我的/个人中心 & 钱包/积分/优惠券（任务拆分）

- 前端任务（小程序「我的」页面基础）
  - 个人中心首页：展示头像、昵称、会员等级、合伙人状态、消息入口
  - 订单中心入口：全部/待付款/待发货/待收货/待评价 标签切换与列表跳转
  - 钱包模块：余额、茶币、账单列表、提现入口、银行卡/提现账户管理页
  - 积分模块：积分余额、积分记录、可兑换积分商品列表（跳转到积分商城）
  - 优惠券模块：待使用/已使用/已过期 Tab 切换与可用券高亮

- 后端任务（账号与资产）
  - `GET /api/v1/users/me/summary`：返回个人中心聚合数据（基本信息、会员等级、钱包/积分/优惠券数量汇总）
  - 钱包相关接口：`GET /api/v1/wallet`, `GET /api/v1/wallet/transactions`, `POST /api/v1/wallet/withdrawals`
  - 积分相关接口：`GET /api/v1/points`, `GET /api/v1/points/transactions`, `GET /api/v1/points/products`
  - 优惠券相关接口：`GET /api/v1/coupons`（按状态筛选）、在下单时注入可用优惠券列表
  - 银行卡/提现账户管理：`GET/POST/DELETE /api/v1/wallet/bank-accounts`

- 测试任务
  - 校验个人中心聚合接口与各模块入口显示的数据一致（钱包/积分/券数量）
  - 钱包提现流程：发起申请、风控/限额校验、状态流转（pending/approved/rejected/paid）
  - 积分获得与消费：下单获得积分、兑礼消耗积分、积分流水账单正确
  - 优惠券列表/状态切换正确，过期与已使用券不出现在可用列表

### Sprint B — 客服/分享/门店管理/设置（子 Sprint）

- 前端任务
  - 客服与帮助：客服电话一键拨打、意见反馈表单、FAQ/帮助中心入口
  - 分享模块：分享海报/链接入口，展示推广人数、累计佣金/奖励等摘要
  - 门店管理入口：仅门店管理员可见，跳转到门店后台 H5/小程序子包
  - 设置模块：账户管理（手机号、密码、微信绑定）、地址管理、关于我们、隐私/协议页

- 后端任务
  - 客服相关：意见反馈提交接口，后台查看与处理状态字段
  - 分享相关：`GET /api/v1/users/me/share-stats`（推广人数、累计佣金概览）
  - 门店管理员标识与门店管理入口接口：`GET /api/v1/users/me/store-role`
  - 协议与关于我们内容接口：`GET /api/v1/content/pages?slug=about/privacy/agreement`

- 测试任务
  - 非门店管理员账号「门店管理」入口不可见，门店管理员可见且能正确跳转
  - 分享统计数据与实际推广/佣金记录一致
  - 意见反馈提交流程、后台查看与状态流转
  - 协议/关于我们等内容在前端展示与后台配置保持一致

### Sprint B — 关键 API（草案）

- POST /api/v1/auth/login
  - 功能：手机号/验证码登录或微信登录
  - Body: `{phone, code}` 或 `{wechat_code}`
  - Response: `{token, user}`

- GET /api/v1/users/me
  - 功能：获取当前用户信息
  - Auth：需要
  - Response: `{id,name,avatar,member_level,wallet_balance,tea_coin,points}`

- POST /api/v1/membership/purchase
  - 功能：用户购买会员/合伙人礼包
  - Body: `{user_id, package_id, pay_method}`
  - Response: `{order_id, pay_info}`
  - 注意：购买成功后触发权益与茶币发放、等级升级

- GET /api/v1/coupons/templates
  - 功能：获取可发放优惠券模板
  - Response: `[{id,name,type,discount,valid_from,valid_to,conditions}]`

- POST /api/v1/coupons/claim
  - 功能：用户领取优惠券
  - Body: `{user_id, template_id}`
  - Response: `{coupon_id, expires_at}`

- POST /api/v1/points/adjust
  - 功能：增加/扣减积分（内部调用或管理员）
  - Body: `{user_id, amount, reason}`

---

## Sprint C（第6-8周） — 后台管理 + 分销体系 + 门店后台

目标：实现平台后台的商品/订单/门店管理；分销的基础链路（分享、关联、佣金累积）；门店接单与打印功能。

- 前端任务（后台/门店后台）
  - 平台：商品管理页面（新增/编辑/上下架）、订单管理、门店管理
  - 门店后台：接单页面、打印按钮（调用打印服务/触发小票打印）、桌码二维码生成
  - 分销：生成分享卡片/海报、分享入口展示佣金统计
  - 商品管理页面上传：支持从后台管理端直接上传商品主图/详情图到阿里云 OSS（前端向后端请求上传凭证并直传）
    - 子任务（前端示例）:
      - 示例实现位置：`admin-fe/src/services/oss.ts`、`admin-fe/src/components/ProductImageUploader.tsx`
      - 示例逻辑：调用 `POST /api/v1/admin/storage/oss/policy` 获取 policy/credentials -> 使用 OSS 表单直传或 OSS SDK 上传 -> 上传成功后获取 `object_key`/`url` 并回填到商品编辑表单 -> 提交商品保存接口。
      - 前端校验：校验文件类型（jpg/png/webp）、大小上限（2MB）与图片比例提示。
- 后端任务
  - 后台 API：商品 CRUD、订单列表/处理、门店 CRUD、用户权限与角色管理
  - 分销逻辑：记录推荐关系、订单完成时计算并冻结佣金到钱包（status: frozen/available）
  - 打印/通知：支持接单语音通知、打印小票任务队列（可通过 RabbitMQ/Redis 列表实现）
  - 对接阿里云 OSS：实现 `POST /api/v1/admin/storage/oss/policy`（请求示例：`{business, file_name, content_type, file_size}`；返回示例：`{policy, signature, accessKeyId, expire_at, object_key_template}`），并支持返回 STS 临时凭证（可选）。配置 CDN 域名与对象命名策略（示例：`products/{yyyy}/{MM}/{product_id}/{uuid}.jpg`），并建议启用回调/服务端校验以防止非法回填。
    - 子任务（接口定义）：在 `doc/api/oss.md` 或 `doc/prd_sprints.md` 增加完整请求/响应示例、错误码与安全说明（如何校验 object_key、过期策略与权限边界）。
  - 后台商品管理交付：完成 `products`/`product_skus`/`product_media`/`product_categories`/`brands` 的 CRUD 接口，SKU/库存维护接口（含锁库存策略），并在 `pkg/database` 注册自动迁移。
    - 子任务（DB 迁移 SQL）：在 `db/migrations/` 新增 migration SQL 文件，包括但不限于：
      - `product_media(id BIGINT PK, product_id BIGINT, type VARCHAR, object_key VARCHAR, url VARCHAR, sort INT, created_at DATETIME)`
      - `product_skus(id, product_id, sku_name, barcode, price_cents, cost_cents, stock, status, created_at)`
      - `products(id, name, type, category_id, brand_id, base_price_cents, shipping_template_id, status, main_image_url, detail_rich_text, extra_attrs_json, created_at)`
      - `brands(id, name, logo_url, origin_region_id, description)`
      - `product_categories(id, parent_id, name, sort_order, is_visible, use_case_flags)`
    - 子任务（后端实现与迁移注册）：在 `pkg/database` 添加对应 GORM 模型并在初始迁移路径注册 autoMigrate。
    - 子任务（测试用例）：编写后端单元/集成测试以验证 policy 接口、media 记录创建、以及 SKU 库存锁定/释放逻辑（建议放在 `tea-api/internal/.../*_test.go`）。
- 测试任务
  - 分销佣金计算场景、提现阈值与提现流程测试
  - 门店接单并打印流程测试（本地或模拟打印机服务）

### Sprint C — 关键 API（草案）

- POST /api/v1/admin/products
  - 功能：创建/编辑商品（Admin 权限）
  - Body: `{name,category_id,price,stock,images,shipping_template_id,...}`
  - Auth：Admin

- GET /api/v1/admin/orders
  - 功能：后台订单查询（带筛选）
  - Query：`status, date_from, date_to, store_id, page, size`

- POST /api/v1/referral/record
  - 功能：记录推荐关系（当用户通过带参链接或扫推荐码注册）
  - Body: `{referrer_id, referred_user_id, source}`

- POST /api/v1/commission/calc
  - 功能：（内部）在订单完成时计算并冻结佣金
  - Body: `{order_id}`
  - Response: `{commissions:[{user_id,amount,type,level}]}`

- GET /api/v1/partner/packages
  - 功能：合伙人礼包列表（购买入口）

- POST /api/v1/print/jobs
  - 功能：门店打印任务提交（门店后台提交后交给打印队列）
  - Body: `{store_id, order_id, template_type: "kitchen"|"receipt", copies}`

---

## Sprint D（第9-11周） — QA、性能、安全、灰度上线

目标：修复 BUG、完善监控日志、进行压测、完成上线准备与回滚策略。

- 任务清单
  - 完成所有关键路径的自动化测试（单元 + 集成 + E2E）
  - 性能压测脚本与目标（接口列表、并发目标）
  - 日志与监控：接入灰度发布/报警规则、关键业务指标监控（订单成功率、支付失败率）
  - 上线准备：数据库迁移脚本、回滚脚本、数据备份、运维 Runbook

---

## 通用注意事项与接口设计原则

- 采用 RESTful 风格，错误码统一：`{code, message, data}`；HTTP 状态码与业务码分离（200 + code 表示成功）。
- 分页统一：`page`, `size`；返回 `total`。
- 时间字段统一使用 UTC ISO8601 格式。
- 所有对财务/提现/佣金的变更都记录流水表并支持幂等操作。
- 安全：敏感接口（提现、合伙人升级）需要二次验证/人工审核标记。

---

## 后续工作（可选）

- 将上面的接口草案转为 OpenAPI (Swagger) 文档，自动生成后端路由与前端 Mock 数据。
- 根据实际开发团队人数与每 Sprint 周期把任务拆成更细的工单（JIRA/Trello 格式），并分配负责人与估时。

---

## 管理权限 / 提现 / 合伙人等级 — OpenAPI/接口契约（草案）

说明：以下为针对新补充的数据库表（`permissions`、`role_permissions`、`partner_levels`、`user_bank_accounts`、`withdrawal_requests` 等）设计的管理与用户端接口契约草案，便于快速生成 OpenAPI（Swagger）并对接前后端。

鉴权与安全：

- 管理后台接口均需 `Authorization: Bearer <token>` 且用户需具备 `admin` 或对应管理角色权限。
- 用户发起提现需登录且通过二次验证（如支付密码或短信验证码），敏感操作管理员审批需要记录 `operator_id` 与审批日志。
- 幂等：对可能被重复提交的 POST 接口（提现、权限分配等）建议支持 `Idempotency-Key` 请求头。

统一错误模型：

```json
{ "code": 4001, "message": "错误描述", "data": null }
```

统一分页：Query 参数 `page`、`size`，响应包含 `{total, page, size, items:[]}`。

---

### 权限管理（permissions）

- GET /api/v1/admin/permissions
  - 描述：权限列表（支持分页/搜索）
  - Query：`q?`, `page`, `size`
  - Response: `{total,page,size,items:[{id,code,name,description,created_at,updated_at}]}`

- POST /api/v1/admin/permissions
  - 描述：新增权限
  - Body: `{code: string, name: string, description?: string}`
  - Response: `{id,code,name,description,created_at}`

- GET /api/v1/admin/permissions/{id}
  - 描述：权限详情

- PUT /api/v1/admin/permissions/{id}
  - 描述：更新权限
  - Body: `{name?, description?}`

- DELETE /api/v1/admin/permissions/{id}
  - 描述：删除权限（需先检查是否被角色引用）

示例权限对象：

```json
{ "id": 101, "code": "orders:manage", "name": "订单管理", "description": "查看/编辑订单", "created_at": "2025-12-01T12:00:00Z" }
```

---

### 角色-权限绑定（role_permissions）

- GET /api/v1/admin/roles/{role_id}/permissions
  - 描述：获取角色拥有的权限列表
  - Response: `{role_id, permissions:[{id,code,name}]}`

- POST /api/v1/admin/roles/{role_id}/permissions
  - 描述：批量设置角色权限（覆盖式或增量，建议使用覆盖或提供 `mode`）
  - Body: `{permission_ids: [1,2,3], mode?: "replace"|"append"}`
  - Response: `{role_id, permissions:[...]}`

- DELETE /api/v1/admin/roles/{role_id}/permissions/{permission_id}
  - 描述：移除单个权限

注意：前端后台权限点请与 `permissions.code` 保持一致以便统一校验（如 `admin:users`、`finance:withdrawals`）。

---

### 合伙人等级 / 套餐（partner_levels / membership_packages 相关）

- GET /api/v1/admin/partner-levels
  - 描述：合伙人等级/套餐列表
  - Response: `{items:[{id,name,level,purchase_amount_threshold,direct_commission_rate,team_commission_rate,discount_rate,benefits}], total}`

- POST /api/v1/admin/partner-levels
  - 描述：创建合伙人等级
  - Body: `{name:string, level:int, purchase_amount_threshold:int, direct_commission_rate: decimal, team_commission_rate: decimal, discount_rate: decimal, benefits?:object}`

- GET /api/v1/admin/partner-levels/{id}
- PUT /api/v1/admin/partner-levels/{id}
- DELETE /api/v1/admin/partner-levels/{id}

示例对象（注意：金额以分为单位）：

```json
{ "id": 1, "name": "银牌合伙人", "level": 2, "purchase_amount_threshold": 19900, "direct_commission_rate": 0.06, "team_commission_rate": 0.02, "discount_rate": 0.05, "benefits": {"tea_coins":100} }
```

前端注意：合伙人购买页需展示权益、折扣、预计返佣示例（下单金额×佣金比例），并在购买成功后触发等级升级与佣金/茶币发放流程。

---

### 用户银行卡/提现账户（user_bank_accounts）

- GET /api/v1/users/{user_id}/bank-accounts
  - 描述：用户名下的提现账户列表（支持银行卡/微信/支付宝/对公账户）

- POST /api/v1/users/{user_id}/bank-accounts
  - 描述：新增提现账户（需验证例如银行卡四要素或短信验证）
  - Body: `{type: "bank"|"wechat"|"alipay", account_no, account_name, bank_name?, card_tail?}`
  - Response: `{id, masked_account, type, created_at}`

- PUT /api/v1/users/{user_id}/bank-accounts/{id}
- DELETE /api/v1/users/{user_id}/bank-accounts/{id}

注意：返回给前端的 `account_no` 应脱敏（只保留尾号），敏感信息后端加密存储。

---

### 提现申请（withdrawal_requests）

用户端：

- POST /api/v1/users/{user_id}/withdrawals
  - 描述：发起提现申请
  - Headers: `Idempotency-Key` 推荐
  - Body: `{bank_account_id:int, amount_cents:int, currency:"CNY", note?:string, verify_code?:string}`
  - 验证：校验余额、最低提现额度、手续费规则；记录流水并将提现状态设置为 `pending`。
  - Response: `{withdrawal_id, status: "pending", requested_at}`

- GET /api/v1/users/{user_id}/withdrawals
  - 描述：用户查看自己的提现记录（分页）

管理员端：

- GET /api/v1/admin/withdrawals
  - 描述：管理员分页查询提现申请，支持过滤 `status`, `user_id`, `date_from`, `date_to`

- POST /api/v1/admin/withdrawals/{id}/approve
  - 描述：审批通过（需记录 `operator_id`, `operator_note`），触发实际出款或调用第三方打款接口，状态流转 `pending`→`approved`→`paid`。
  - Body: `{operator_id:int, operator_note?:string, paid_at?:string, external_txn_id?:string}`

- POST /api/v1/admin/withdrawals/{id}/reject
  - 描述：拒绝提现（需返回资金到用户钱包并记录流水）
  - Body: `{operator_id:int, reason:string}`

提现对象示例：

```json
{ "id": 1001, "user_id": 5001, "bank_account_id": 77, "amount_cents": 100000, "fee_cents": 300, "net_amount_cents": 99700, "status": "pending", "requested_at": "2025-12-05T10:00:00Z" }
```

审计与合规：管理员审批必须记录凭证（外部流水号/打款截图/接口返回），并支持导出 CSV 供财务核对。

---

### 接口契约说明（附加条款）

- 所有写操作（POST/PUT/DELETE）应返回标准响应：`{code:int, message:string, data: {...}}`。
- 金额使用整数（分）以避免浮点误差。
- 幂等键（`Idempotency-Key`）建议用于：`POST /withdrawals`、`POST /admin/roles/{id}/permissions`（批量）等接口。
- 审批类操作管理员端应支持批量操作（可选），但应保证逐条记录审核日志以便回溯。
- 建议将上述契约导出为 OpenAPI 3.0 YAML，用于：后端路由生成、前端 Mock Server 以及自动化测试。

---

下一步建议：

- 将本草案转为 OpenAPI YAML（我可以帮生成）并同步到 `doc/api/partner_admin_openapi.yaml`。
- 或者我可按你偏好生成 `tea-api` 服务端的接口 stub（Go Gin/chi）与前端 Mock。

---

### 佣金分配与结算 — OpenAPI 摘要（已生成）

说明：已将基于需求 `2.3 佣金分配机制与结算` 的 OpenAPI 文档写入 `doc/openapi/commission.yaml`。
该规范覆盖：佣金计算预览、佣金记录持久化、佣金解冻（系统/管理员触发）、提现申请与审批、合伙人等级与平台财务配置查询等接口，并对货币使用“分（整数）”做了统一约定。

关键信息（摘录）

- 文档路径：`doc/openapi/commission.yaml`
- 约定：所有金额字段使用整数“分”（schema 名称：`MoneyCents`）
- 幂等：`POST /commissions` 与 `POST /withdrawals` 建议支持 `Idempotency-Key` 请求头

示例：计算佣金（预览，不写库）

请求：

`POST /api/v1/orders/ord_20251208_0001/commissions/calculate`

Body:

```json
{
  "order_id": "ord_20251208_0001",
  "payer_user_id": "user_999",
  "referrer_user_id": "user_123",
  "items": [
    {"sku_id":"sku_1","unit_price_cents":100000,"quantity":1,"discount_cents":0}
  ],
  "shipping_cents":0,
  "coupon_cents":0,
  "order_level_discount_cents":0
}
```

响应示例（201）：

```json
{
  "id": "comm_batch_20251208_0001",
  "order_id": "ord_20251208_0001",
  "created_commissions": 2,
  "status": "created"
}
```

示例：持久化佣金记录（订单完成后调用）

请求：

`POST /api/v1/commissions`

Headers: `Idempotency-Key: order-ord_20251208_0001-commissions`

Body:

```json
{
  "order_id":"ord_20251208_0001",
  "calculated_at":"2025-12-08T12:00:00Z",
  "breakdown":[ /* 与计算返回的 items 同结构 */ ]
}
```

响应示例（201）：

```json
{ "created_count": 2, "details": [ /* CommissionRecord 列表 */ ] }
```

示例：用户发起提现

请求：

`POST /api/v1/withdrawals`

Headers: `Idempotency-Key: wd-user_123-20251208-01`

Body:

```json
{
  "user_id":"user_123",
  "amount_cents":10000,
  "bank_account_id":"ba_77",
  "note":"提现到银行卡尾号 8899"
}
```

响应示例（201）：

```json
{
  "id":"wd_20251208_0001",
  "user_id":"user_123",
  "amount_cents":10000,
  "fee_cents":100,
  "net_amount_cents":9900,
  "status":"pending"
}
```

备注：详细 schema 与全部接口定义见 `doc/openapi/commission.yaml`，可用于生成 Mock Server 与客户端 SDK。
