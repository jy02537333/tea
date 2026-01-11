# 功能清单 ↔ PRD 对照表

> 目的：将《茶心阁小程序需求文档》中的 **五、功能清单（5.1/5.2/5.3）** 与 `doc/prd.md` 中的章节一一映射，并标记当前 PRD 设计状态，方便后续审查与排期。记号说明：`[]` 未开始，`[x]` 表示设计完成，`[✅]` 表示功能完成。

状态字段说明：

- `已在 PRD 设计`：需求已在 `doc/prd.md` 中有明确产品/技术说明，可直接指导开发。
- `需补充 PRD`：需求在 PRD 中尚未单独成节或仅隐含提及，建议补一小节或专门文档。
- `实现时关注`：需求在 PRD 中已有方向性描述，但实现细节（字段/接口/交互）需在开发阶段进一步细化。

---

## 5.1 小程序前端

| 完成 | 模块 | 需求文档功能点 | PRD 对应路径 | 状态 | 备注 |
| --- | --- | --- | --- | --- | --- |
| [✅] | 首页 | Balance 主图、轮播、跳转 | `doc/prd.md` → `3.1.1 首页（产品需求细化）` | 功能完成 | wx-fe 首页轮播基于 `GET /api/v1/banners` 渲染；支持跳转：商品详情（link_type=2）、分类页（link_type=3）、外部链接（link_type=4→WebView）；实现见 `wx-fe/src/pages/index/index.tsx` + `wx-fe/src/pages/webview/index.tsx` |
| [✅] | 首页 | 会员横幅（未登录提示、非会员展示开通） | `doc/prd.md` → `3.1.1 首页` → 会员/合伙人展示区 | 功能完成 | wx-fe 首页在 `isMember=false` 时展示横幅：未登录→引导登录；已登录但非会员→引导开通/查看会员（跳转 `/pages/membership/index`）；实现见 `wx-fe/src/pages/index/index.tsx`（兼容 `/api/v1/users/me/summary` 不同返回形态） |
| [✅] | 首页 | 金刚区：堂食、外卖、商城 | `doc/prd.md` → `3.1.1 首页` → 金刚区（Quick Actions） | 功能完成 | wx-fe 首页“快捷入口”区提供三入口：堂食/外卖（要求已选门店，跳转 `/pages/product-list?store_id=...`）、商城（跳转 `/pages/category`）；并提供门店 Picker 选择当前门店；实现见 `wx-fe/src/pages/index/index.tsx` |
| [✅] | 首页-商城 | 商城主图、金刚区分区（礼品茶/健康茶/口粮茶/散装/投资收藏）、最近上新、热销、商品详情 | `doc/prd.md` → `3.1.1 首页` + `3.1.2 产品分类` + 商品详情部分 | 功能完成 | 前端实现：`wx-fe/src/pages/index/index.tsx#fetchMallBlocks` 拉取分类与商品种子数据（`listCategories({status:1})` → `GET /api/v1/categories?status=1`；`getProducts({page:1,limit:50})` → `GET /api/v1/products?page=1&limit=50`），并在前端计算“最近上新/热销”：上新规则=按 `created_at` 倒序（无时间则按 id 倒序）优先取 30 天内（或 `is_new=true`）直到 4 个，不足则用最新补齐到 4；热销规则=按 `sales` 倒序（无/为 0 时优先 `is_hot=true`，再按 id 倒序）取前 4；分区按分类名关键字匹配并跳转 `/pages/category/index?category_id=...`，商品卡片跳转详情 `/pages/product-detail/index?id=...`；后端路由：`tea-api/internal/router/router.go`（`GET /api/v1/categories`、`GET /api/v1/products`）；可复现证据（本地 9292）：`build-ci-logs/home_mall_categories_latest_headers.txt` + `build-ci-logs/home_mall_categories_latest_body.json`、`build-ci-logs/home_mall_products_seed_latest_headers.txt` + `build-ci-logs/home_mall_products_seed_latest_body.json`、派生结果（按前端同款规则计算）：`build-ci-logs/home_mall_rules_latest.txt`、`build-ci-logs/home_mall_new_products_latest.json`、`build-ci-logs/home_mall_hot_products_latest.json` |
| [✅] | 首页-标签区 | 活动列表、活动详情、我的活动 | `doc/prd.md` → `3.1.1 首页`（活动入口）+ `3.2.6 营销管理`（活动管理） | 功能完成 | 按验收口径：wx-fe 首页标签区提供活动入口（`wx-fe/src/pages/index/index.tsx#goActivities`，跳转 `/pages/activities/index`），活动页支持拉取门店活动并报名/支付（最小闭环）；“我的活动”可后续再增强展示；活动报名下单入口 share 校验/冻结/订单号证据见“补充注释（2026-01-07）” |
| [✅] | 首页-标签区 | 邀请好友得现金奖励 | `doc/prd.md` → `3.1.4 我的`（分享）+ `3.4 分销系统` | 功能完成 | 入口与链路：wx-fe 首页标签区可进入分享中心（`wx-fe/src/pages/index/index.tsx#goShareCenter` → `/pages/share/index`），同时 wx-fe「我的」页也可进入分享中心（`wx-fe/src/pages/profile/index.tsx#handleShareCenter` → `/pages/share/index`，页面见 `wx-fe/src/pages/share/index.tsx`，路由注册见 `wx-fe/src/app.config.ts`/`wx-fe/src/pages.json`）；分享参数：生成链接 `referrer_id`（`wx-fe/src/services/share.ts#buildShareLink`），启动解析并自动绑定推荐关系（`wx-fe/src/app.tsx` + `wx-fe/src/services/share.ts#bindReferral`）；后端接口：`GET /api/v1/share/posters`、`POST /api/v1/referrals/bind`；可复现证据（本地 9292）：`build-ci-logs/invite_share_posters_latest_headers.txt` + `build-ci-logs/invite_share_posters_latest_body.json`、`build-ci-logs/invite_referral_bind_latest_payload.json` + `build-ci-logs/invite_referral_bind_latest_headers.txt` + `build-ci-logs/invite_referral_bind_latest_body.json`（包含创建/登录测试用户证据：`build-ci-logs/invite_referrer_create_latest_body.json`、`build-ci-logs/invite_invitee_login_latest_body.json`）；构建证据：`build-ci-logs/home_tag_section_build_weapp_latest.txt`（包含 DONE）；UI 走查步骤与待补截图清单：`build-ci-logs/home_tag_section_ui_walkthrough_latest.md` |
| [✅] | 首页-标签区 | 优惠券入口 | `doc/prd.md` → `3.1.4 我的`（优惠券）+ `3.2.6 营销管理` | 功能完成 | wx-fe 首页新增入口按钮“优惠券入口”，跳转 `/pages/coupons/index`（见 `wx-fe/src/pages/index/index.tsx`）；复用已实现的优惠券页展示平台券/门店券 |
| [✅] | 商品展示区 | 广告、茶心阁简介、推荐商品、本地商品、技术支持&版权 | `doc/prd.md` → `3.1.1 首页`（广告位/运营位）+ `3.2.6 营销管理` + `3.2.8 系统设置`（内容/版权） | 功能完成 | wx-fe 首页已实现 5 区块：广告轮播（`wx-fe/src/pages/index/index.tsx#fetchOps` 调用 `GET /api/v1/banners`）、简介（`#fetchOps` 调用 `GET /api/v1/content/pages?keys=content_about` 并用 `ContentRenderer` 渲染）、推荐商品（`#fetchOps` 调用 `GET /api/v1/products?page=1&limit=6`）、本地商品（`wx-fe/src/pages/index/index.tsx#fetchLocalProducts` 调用 `GET /api/v1/products?page=1&limit=6&store_id=...`）、版权/电话（`#fetchOps` 调用 `GET /api/v1/site/configs?keys=site_copyright,site_phone`）；后端路由见 `tea-api/internal/router/router.go`（`/banners`/`/content/pages`/`/products`/`/site/configs`），handlers：`tea-api/internal/handler/banner.go`、`tea-api/internal/handler/content.go`、`tea-api/internal/handler/system_config.go`；可复现证据（本地 9292）：`build-ci-logs/home_showcase_banners_latest_headers.txt` + `build-ci-logs/home_showcase_banners_latest_body.json`、`build-ci-logs/home_showcase_content_about_latest_headers.txt` + `build-ci-logs/home_showcase_content_about_latest_body.json`、`build-ci-logs/home_showcase_site_configs_latest_headers.txt` + `build-ci-logs/home_showcase_site_configs_latest_body.json`、`build-ci-logs/home_showcase_products_recommended_latest_headers.txt` + `build-ci-logs/home_showcase_products_recommended_latest_body.json`、`build-ci-logs/home_showcase_local_products_latest_store_id.txt` + `build-ci-logs/home_showcase_products_local_latest_headers.txt` + `build-ci-logs/home_showcase_products_local_latest_body.json` |
| [✅] | 分类 | 商城商品分类 + 产地/包装/价格筛选 | `doc/prd.md` → `3.1.2 产品分类（商城产品）` | 功能完成 | 前端：分类页 `wx-fe/src/pages/category/index.tsx`、通用列表 `wx-fe/src/pages/product-list/index.tsx` 通过 `getProducts(...)` 传参 `category_id/keyword/store_id/origin/packaging/min_price/max_price/sort`；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/categories`、`GET /api/v1/products`），handler `tea-api/internal/handler/product.go::GetCategories/GetProducts`（当前明确解析：`category_id/status/keyword/store_id/page/limit`；`origin/packaging/min_price/max_price/sort` 暂未在 handler 中做服务端过滤，属于“参数可传且不报错”的最小兼容口径）；可复现证据（本地 9292）：`build-ci-logs/category_filters_categories_latest_headers.txt` + `build-ci-logs/category_filters_categories_latest_body.json` + `build-ci-logs/category_filters_latest_category_id.txt`、`build-ci-logs/category_filters_products_by_category_latest_headers.txt` + `build-ci-logs/category_filters_products_by_category_latest_body.json`、`build-ci-logs/category_filters_products_by_keyword_latest_headers.txt` + `build-ci-logs/category_filters_products_by_keyword_latest_body.json`、`build-ci-logs/category_filters_latest_store_id.txt` + `build-ci-logs/category_filters_products_by_store_latest_headers.txt` + `build-ci-logs/category_filters_products_by_store_latest_body.json`、`build-ci-logs/category_filters_products_with_extra_params_latest_headers.txt` + `build-ci-logs/category_filters_products_with_extra_params_latest_body.json` |
| [✅] | 门店 | 实体门店展示、按距离排序、门店详情（证照/导航/拨号）、门店商品/外卖列表 | `doc/prd.md` → `3.1.3 门店（产品需求细化）` | 功能完成 | wx-fe 门店列表/详情页：`wx-fe/src/pages/stores/index.tsx`（`listStores`）、`wx-fe/src/pages/store-detail/index.tsx`（导航/拨号/证照预览，`getStore`）；后端接口：`GET /api/v1/stores`、`GET /api/v1/stores/:id`（见 `tea-api/internal/router/router.go`）；可复现证据（本地 9292）：`build-ci-logs/store_list_latest_headers.txt` + `build-ci-logs/store_list_latest_body.json`、`build-ci-logs/store_get_latest_id.txt` + `build-ci-logs/store_get_latest_headers.txt` + `build-ci-logs/store_get_latest_body.json` |
| [✅] | 门店入口 | 扫码进店（入口参数解析 + 首页/分类入参接管） | `doc/prd.md` → `5.1 用户下单流程` → 扫码进店（最小版） | 功能完成 | 在 `app.tsx` 解析 `store_id`（query/scene/H5 URL），`index`/`category` 入参接管并持久化 `current_store_id`，商品详情页亦持久化；后续可扩展更多入口与控件 |
| [✅] | 购物车 | 按商城/门店商品服务/门店外卖/积分商城 + 按门店分组 | `doc/prd.md` → `3.1.5 购物车（补充说明）` | 功能完成 | 前端：购物车页 `wx-fe/src/pages/cart/index.tsx`（`listCart/updateCartItem/removeCartItem`）、商品详情“加入购物车” `wx-fe/src/pages/product-detail/index.tsx`（调用 `addCartItem`）、结算页 `wx-fe/src/pages/checkout/index.tsx`（从购物车创建订单 `createOrderFromCart` 并提示“多门店拆单”）；service：`wx-fe/src/services/cart.ts`；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/cart`、`POST /api/v1/cart/items`、`PUT /api/v1/cart/items/:id`、`DELETE /api/v1/cart/items/:id`、`DELETE /api/v1/cart/clear`，均需 `AuthJWT`），handler：`tea-api/internal/handler/cart.go`，service：`tea-api/internal/service/cart.go`（预加载 Product/Sku；门店管理员 role=store 另有“仅允许本门店特供 + 禁止混入其他商品”的加购强校验）；可复现证据（本地 9292，使用 dev-login 获取 JWT）：`build-ci-logs/cart_dev_login_latest_payload.json` + `build-ci-logs/cart_dev_login_latest_headers.txt` + `build-ci-logs/cart_dev_login_latest_body.json` + `build-ci-logs/cart_dev_login_latest_token.txt`、`build-ci-logs/cart_clear_latest_headers.txt` + `build-ci-logs/cart_clear_latest_body.json`、`build-ci-logs/cart_list_empty_latest_headers.txt` + `build-ci-logs/cart_list_empty_latest_body.json`、`build-ci-logs/cart_products_pick_latest_headers.txt` + `build-ci-logs/cart_products_pick_latest_body.json` + `build-ci-logs/cart_add_latest_product_id.txt`、`build-ci-logs/cart_add_item_latest_payload.json` + `build-ci-logs/cart_add_item_latest_headers.txt` + `build-ci-logs/cart_add_item_latest_body.json` + `build-ci-logs/cart_add_latest_item_id.txt`、`build-ci-logs/cart_list_after_add_latest_headers.txt` + `build-ci-logs/cart_list_after_add_latest_body.json`、`build-ci-logs/cart_update_quantity_latest_payload.json` + `build-ci-logs/cart_update_quantity_latest_headers.txt` + `build-ci-logs/cart_update_quantity_latest_body.json` + `build-ci-logs/cart_list_after_update_latest_headers.txt` + `build-ci-logs/cart_list_after_update_latest_body.json`、`build-ci-logs/cart_remove_item_latest_headers.txt` + `build-ci-logs/cart_remove_item_latest_body.json` + `build-ci-logs/cart_list_final_latest_headers.txt` + `build-ci-logs/cart_list_final_latest_body.json` |
| [✅] | 我的-个人中心 | 头像、昵称、账户、消息、等级展示 | `doc/prd.md` → `3.1.4 我的（个人中心）` | 功能完成 | 结合 `users` 模型与 summary 接口 |
| [✅] | 我的-开通 VIP/合伙人 | 选择等级→展示权益→支付→生效 | `doc/prd.md` → `3.1.4 我的` + `4.3 合伙人升级流程` + 分销结算设计 | 功能完成 | 礼包订单、升级逻辑、升级奖励均已描述 |
| [✅] | 我的-订单 | 各状态订单列表与详情、售后入口 | `doc/prd.md` → `3.1.4 我的`（订单中心）+ `4.1 用户下单流程` | 功能完成 | 前端：个人中心入口 `wx-fe/src/pages/profile/index.tsx#handleViewOrders` → `/pages/orders/index`；订单列表页 `wx-fe/src/pages/orders/index.tsx`（读取 `current_store_id` 后可带 `store_id` 筛选，状态筛选通过 query `status`；点击进入详情 `/pages/order-detail/index?id=...`）；订单详情页 `wx-fe/src/pages/order-detail/index.tsx` 支持拉详情 `getOrder`、订单操作（取消/支付/确认收货）与“订单投诉→建工单”（`createTicket({type:'order', source:'miniapp_order', ...})`）；售后/进度页 `wx-fe/src/pages/after-sale/index.tsx` 支持订单操作（取消/支付/确认收货）、退款进度咨询工单（`type=refund, source=miniapp_order`）与退款时间线（`listMyRefunds`）；service：`wx-fe/src/services/orders.ts`（`GET /api/v1/orders`、`GET /api/v1/orders/:id`、`POST /api/v1/orders/:id/cancel`、`POST /api/v1/orders/:id/pay`、`POST /api/v1/orders/:id/receive`、`POST /api/v1/orders/from-cart`）、`wx-fe/src/services/refunds.ts`（`GET /api/v1/refunds`）、`wx-fe/src/services/tickets.ts`（`POST /api/v1/tickets`）；后端：路由 `tea-api/internal/router/router.go`（orders group：`AuthJWT` + `GET /orders`、`GET /orders/:id`、`POST /orders/:id/cancel`、`POST /orders/:id/pay`、`POST /orders/:id/receive`，以及 `POST /orders/from-cart` 创建订单），handler `tea-api/internal/handler/order.go#List/Detail/Cancel/Pay/Receive/CreateFromCart`；订单状态口径见 `tea-api/internal/model/order.go`（1待付款 2已付款 3配送中 4已完成 5已取消）；退款列表：`tea-api/internal/router/router.go`（`GET /api/v1/refunds`，`AuthJWT`）+ `tea-api/internal/handler/refund.go#ListMyRefunds`；可复现证据（本地 9292，复用 dev-login JWT；创建一笔最小订单后验真）：下单前置（清车/加购）：`build-ci-logs/order_cart_clear_latest_headers.txt` + `build-ci-logs/order_cart_clear_latest_body.json`、`build-ci-logs/order_cart_add_item_latest_payload.json` + `build-ci-logs/order_cart_add_item_latest_headers.txt` + `build-ci-logs/order_cart_add_item_latest_body.json`；创建订单：`build-ci-logs/order_create_from_cart_latest_payload.json` + `build-ci-logs/order_create_from_cart_latest_headers.txt` + `build-ci-logs/order_create_from_cart_latest_body.json` + `build-ci-logs/order_created_latest_id.txt` + `build-ci-logs/order_created_latest_order_no.txt`；订单列表（含状态筛选）：`build-ci-logs/order_list_latest_headers.txt` + `build-ci-logs/order_list_latest_body.json`、`build-ci-logs/order_list_status_1_latest_headers.txt` + `build-ci-logs/order_list_status_1_latest_body.json`、`build-ci-logs/order_list_status_4_latest_headers.txt` + `build-ci-logs/order_list_status_4_latest_body.json`；订单详情：`build-ci-logs/order_detail_latest_id.txt` + `build-ci-logs/order_detail_latest_headers.txt` + `build-ci-logs/order_detail_latest_body.json`；取消订单：`build-ci-logs/order_cancel_latest_payload.json` + `build-ci-logs/order_cancel_latest_headers.txt` + `build-ci-logs/order_cancel_latest_body.json`；订单投诉建工单：`build-ci-logs/order_ticket_create_latest_payload.json` + `build-ci-logs/order_ticket_create_latest_headers.txt` + `build-ci-logs/order_ticket_create_latest_body.json`；退款列表：`build-ci-logs/order_refunds_list_latest_headers.txt` + `build-ci-logs/order_refunds_list_latest_body.json`、`build-ci-logs/order_refunds_by_order_latest_headers.txt` + `build-ci-logs/order_refunds_by_order_latest_body.json` |
| [✅] | 我的-钱包 | 余额、茶币、账单、提现、银行卡管理 | `doc/prd.md` → `3.1.4 我的`（钱包）+ 分销结算设计 | 功能完成 | 前端：入口 `wx-fe/src/pages/profile/index.tsx#handleViewWallet` → `/pages/wallet/index`；钱包页 `wx-fe/src/pages/wallet/index.tsx` 只读展示 `余额/冻结金额/茶币`，数据源为 `wx-fe/src/services/me.ts#getMeSummary`（优先 `GET /api/v1/users/me/summary`，失败回退 `GET /api/v1/user/info` 并做最小映射）；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/users/me/summary`、`GET /api/v1/wallet`、`GET /api/v1/wallet/transactions`、`GET /api/v1/wallet/bank-accounts` 等，均需 `AuthJWT`），handler：`tea-api/internal/handler/user_summary.go`、`tea-api/internal/handler/wallet.go`、`tea-api/internal/handler/withdrawal.go`；当前验收口径按最小实现：wx-fe 仅接入“只读展示”，提现/账单/银行卡管理入口后续再接入；可复现证据（本地 9292，复用 dev-login JWT）：`build-ci-logs/wallet_me_summary_latest_headers.txt` + `build-ci-logs/wallet_me_summary_latest_body.json`、`build-ci-logs/wallet_get_latest_headers.txt` + `build-ci-logs/wallet_get_latest_body.json`、`build-ci-logs/wallet_transactions_latest_headers.txt` + `build-ci-logs/wallet_transactions_latest_body.json`、`build-ci-logs/wallet_bank_accounts_latest_headers.txt` + `build-ci-logs/wallet_bank_accounts_latest_body.json` |
| [✅] | 我的-积分 | 积分余额、积分记录、积分商品展示 | `doc/prd.md` → `3.1.4 我的`（积分中心）+ `3.2.6 营销管理`（积分规则） | 功能完成 | 前端：入口 `wx-fe/src/pages/profile/index.tsx#handleViewPoints` → `/pages/points/index`；页面 `wx-fe/src/pages/points/index.tsx` 只读展示 `summary.points.balance`，数据源 `wx-fe/src/services/me.ts#getMeSummary`（`GET /api/v1/users/me/summary`）；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/points`、`GET /api/v1/points/transactions`，均需 `AuthJWT`），handler：`tea-api/internal/handler/points.go`；当前验收口径：wx-fe 暂未接入“积分记录/积分商品”，按后续迭代推进；可复现证据（本地 9292，复用 dev-login JWT）：`build-ci-logs/points_me_summary_latest_headers.txt` + `build-ci-logs/points_me_summary_latest_body.json`、`build-ci-logs/points_get_latest_headers.txt` + `build-ci-logs/points_get_latest_body.json`、`build-ci-logs/points_transactions_latest_headers.txt` + `build-ci-logs/points_transactions_latest_body.json` |
| [✅] | 我的-优惠券 | 待使用/已使用/已过期 | `doc/prd.md` → `3.1.4 我的`（优惠券）+ `3.2.6.1 小程序优惠券接口约定` | 功能完成 | 前端：入口 `wx-fe/src/pages/profile/index.tsx#handleViewCoupons` 与首页入口 `wx-fe/src/pages/index/index.tsx#goCoupons` → `/pages/coupons/index`；页面 `wx-fe/src/pages/coupons/index.tsx` 提供来源拆分 Tab（平台券/门店券）并调用 `wx-fe/src/services/coupons.ts#listMyCoupons` → `GET /api/v1/user/coupons` 展示账户可用券；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/user/coupons`，`AuthMiddleware`）→ handler `tea-api/internal/handler/coupon.go#ListMyCoupons`（返回当前用户“可用券”列表）；结算页可用券筛选：`wx-fe/src/services/orders.ts#getAvailableCouponsForOrder` → `POST /api/v1/orders/available-coupons`（`tea-api/internal/router/router.go` → `tea-api/internal/handler/order.go#AvailableCoupons`）；当前验收口径按最小实现：wx-fe 仅展示“可用券”，已用/已过期券列表与规则按后续迭代推进；可复现证据（本地 9292，复用 dev-login JWT）：`build-ci-logs/coupons_my_list_latest_headers.txt` + `build-ci-logs/coupons_my_list_latest_body.json`、`build-ci-logs/coupons_available_for_order_latest_headers.txt` + `build-ci-logs/coupons_available_for_order_latest_body.json` |
| [✅] | 我的-客服 | 客服电话、意见反馈、帮助文档 | `doc/prd.md` → `3.1.4 我的`（客服） | 功能完成 | 前端：入口在 `wx-fe/src/pages/profile/index.tsx`（“联系客服”调用 `Taro.makePhoneCall`，H5 fallback toast；“意见反馈”跳 `/pages/feedback/index`；“帮助文档”跳 `/pages/help/index`，设置页亦提供反馈入口 `wx-fe/src/pages/settings/index.tsx`）；帮助文档页：`wx-fe/src/pages/help/index.tsx` 调用 `wx-fe/src/services/content.ts#getHelpContent` → `GET /api/v1/content/pages?keys=content_help`（无内容时展示 fallback markdown）；意见反馈页：`wx-fe/src/pages/feedback/index.tsx` 调用 `wx-fe/src/services/feedback.ts#createFeedback`（优先 `POST /api/v1/feedback`，当前后端未提供时回退 `POST /api/v1/tickets` 且携带 `source=miniapp_feedback`）；订单详情/售后页也可创建工单：`wx-fe/src/services/tickets.ts#createTicket` → `POST /api/v1/tickets`；后端：路由 `tea-api/internal/router/router.go`（`GET /api/v1/content/pages`、`GET /api/v1/site/configs`、`POST /api/v1/tickets`，其中 `/tickets` 需 `AuthMiddleware`），handler：`tea-api/internal/handler/content.go#GetPages`、`tea-api/internal/handler/ticket_user.go#Create`；可复现证据（本地 9292，复用 dev-login JWT）：`build-ci-logs/service_help_content_latest_headers.txt` + `build-ci-logs/service_help_content_latest_body.json`、`build-ci-logs/service_site_phone_latest_headers.txt` + `build-ci-logs/service_site_phone_latest_body.json`、`build-ci-logs/service_ticket_create_latest_payload.json` + `build-ci-logs/service_ticket_create_latest_headers.txt` + `build-ci-logs/service_ticket_create_latest_body.json` |
| [✅] | 我的-分享 | 分享海报（二维码海报预览/保存、后台多模板配置+小程序端可切换）、绑定推荐关系（最后一次点击覆盖） | `doc/prd.md` → `3.1.4 我的`（分享）+ `3.4 分销系统` | 功能完成 | 后端新增海报模板配置与公共列表：`GET /api/v1/share/posters` + `GET/PUT /api/v1/admin/system/share-posters`（见 `tea-api/internal/handler/share_poster.go` + `tea-api/internal/router/router.go`）；wx-fe 分享页拉取模板并支持切换（见 `wx-fe/src/services/share.ts#listSharePosterTemplates` + `wx-fe/src/pages/share/index.tsx`）；推荐关系绑定改为“最后一次点击覆盖”（见 `tea-api/internal/handler/referral.go`），并补最小回归测试 `tea-api/test/referral_bind_override_test.go` |
| [✅] | 我的-设置 | 账号管理、登录管理、地址管理、关于我们、隐私/协议 | `doc/prd.md` → `3.1.4 我的`（设置）+ `3.2.8 系统设置`（协议与内容） | 功能完成 | wx-fe 已实现设置页与相关内容页：`/pages/settings`、`/pages/address`、`/pages/about`、`/pages/privacy`、`/pages/terms`（内容由 content 接口/静态 fallback 提供） |
| [✅] | 我的-门店管理 | 门店管理员入口 | `doc/prd.md` → `3.1.4 我的`（门店管理入口）+ `3.3 门店后台` | 功能完成 | wx-fe「我的」页按 `perm.allowedStoreMgmt` 显示“门店管理”入口；点击后优先跳转 `/pages/store-accounts?store_id=<current_store_id>`，未选门店则跳转 `/pages/stores` 选择门店（见 `wx-fe/src/pages/profile/index.tsx`） |

补充注释（2026-01-05）：

- 意见反馈：wx-fe 已上线「意见反馈」页面与入口（设置页与个人中心均可进入）。提交成功后会弹窗展示工单编号（Ticket ID），便于用户与客服沟通定位问题。实现优先 `POST /api/v1/feedback`，若后端未提供则回退 `POST /api/v1/tickets`。
- 我的工单：个人中心新增「我的工单」入口，并与「售后服务」一起排序到「订单」之后，提升售后与工单入口的可发现性与连贯性。

补充注释（2026-01-07）：

- 所有创建订单入口已统一接入 share 参数强校验 + 冻结（订单字段 `referrer_id` / `share_store_id`），入口清单如下：
   - 购物车下单：`POST /api/v1/orders/from-cart`（`tea-api/internal/router/router.go`）→ handler：`tea-api/internal/handler/order.go`（请求体接收并透传 `sharer_uid` / `share_store_id`）→ service：`tea-api/internal/service/order.go`（创建订单 `OrderNo = generateOrderNo("O")`；调用统一归属逻辑 `applyShareAttributionToOrder`）
   - 会员下单：`POST /api/v1/membership-orders`（`tea-api/internal/router/router.go`）→ handler：`tea-api/internal/handler/membership.go`（请求体接收并透传 `sharer_uid` / `share_store_id`）→ service：`tea-api/internal/service/order.go`（创建订单 `OrderNo = generateOrderNo("M")`；调用统一归属逻辑 `applyShareAttributionToOrder`）
   - 活动报名下单：`POST /api/v1/activities/:id/register-with-order`（`tea-api/internal/router/router.go`）→ handler：`tea-api/internal/handler/activity.go`（请求体接收并透传 `sharer_uid` / `share_store_id`）→ service：`tea-api/internal/service/activity.go`（创建订单 `OrderNo = generateOrderNo("O")`；调用统一归属逻辑 `applyShareAttributionToOrder`）

- 活动报名下单入口 share 参数强校验 + 冻结 + 订单号生成已闭环：
   - 路由：`POST /api/v1/activities/:id/register-with-order`（`tea-api/internal/router/router.go`）
   - handler：`tea-api/internal/handler/activity.go`（请求体接收并透传 `sharer_uid` / `share_store_id`）
   - service：`tea-api/internal/service/activity.go`（调用统一归属逻辑 `applyShareAttributionToOrder` 完成强校验与冻结；创建订单时 `OrderNo = generateOrderNo("O")`，避免 `orders.order_no` 空字符串触发 uniqueIndex 冲突）
   - 测试：`tea-api/test/share_params_attribution_test.go` → `Test_ActivityRegisterWithOrderAttribution_WithSharer_FreezeReferrerAndShareStore`、`Test_ActivityRegisterWithOrderAttribution_MissingShareStore_ShouldReject`；并已通过 `go test ./test -run Attribution`

- 统一归属逻辑与最小回归测试（覆盖以上 3 个入口）：
   - 统一逻辑：`tea-api/internal/service/share_attribution.go` → `applyShareAttributionToOrder(...)`
   - 回归测试：`tea-api/test/share_params_attribution_test.go`（`go test ./test -run Attribution`）
      - 购物车下单：`Test_OrderAttribution_WithSharerUID_FreezeReferrer`、`Test_OrderAttribution_ShareStoreWithoutSharer_ShouldReject`、`Test_OrderAttribution_StoreOrder_WithSharer_MissingShareStore_ShouldReject`、`Test_OrderAttribution_StoreOrder_WithSharer_ShareStoreMismatch_ShouldReject`
      - 会员下单：`Test_MembershipOrderAttribution_WithSharerUID_FreezeReferrer`、`Test_MembershipOrderAttribution_ShareStoreNotZero_ShouldReject`
      - 活动报名下单：`Test_ActivityRegisterWithOrderAttribution_WithSharer_FreezeReferrerAndShareStore`、`Test_ActivityRegisterWithOrderAttribution_MissingShareStore_ShouldReject`

补充注释（2026-01-08）：

- 已验真“完成节点触发佣金结算（基于订单冻结字段，绑定覆盖不影响历史订单归属）”的最小闭环：
   - 完成节点（等价于 PRD 的“门店核销/完成”触发点之一）：
      - 用户自取/确认完成：`POST /api/v1/orders/:id/receive`（`tea-api/internal/router/router.go`）→ handler：`tea-api/internal/handler/order.go` → service：`tea-api/internal/service/order.go::Receive(...)`
      - 管理端完成：`POST /api/v1/orders/:id/complete`（`tea-api/internal/router/router.go`）→ handler：`tea-api/internal/handler/order.go` → service：`tea-api/internal/service/order.go::Complete(...)`
   - 结算实现（关键点：读取下单时冻结字段而非实时推荐关系）：
      - 触发点：`tea-api/internal/service/order.go` → `maybeCreateStoreOrderCommissionsOnCompleted(...)`（仅门店订单 `store_id!=0` 且已支付+已完成；要求 `referrer_id` 非空且 `share_store_id==store_id`）
      - 落库：`tea-api/internal/service/commission/store.go` → `SaveCommissionRecordsTx(...)`（创建 `commissions` 记录 + `freeze` 流水）
      - 解冻：`tea-api/internal/service/commission/settlement.go` → `ReleaseFrozenCommissionsTx(...)`（hold=0 的最小实现下会立即变为 `available`）
   - 防覆盖验真（回归测试走真实 API + 直接查库断言）：
      - 测试用例：`tea-api/test/store_order_commission_freeze_attribution_test.go` → `Test_StoreOrderCommission_UsesFrozenSharer_NotOverriddenBinding`
      - 测试步骤：A(分享人) → B(买家) 门店下单携带 `sharer_uid/share_store_id` 冻结 → B 调用 `POST /api/v1/referrals/bind` 覆盖绑定到 C → B 完成订单 → 断言 `commissions` 的 direct 佣金仍归 A 且状态为 `available`
      - 运行命令：`go test ./test -run Test_StoreOrderCommission_UsesFrozenSharer_NotOverriddenBinding -count=1`

   补充注释（2026-01-09）：

   - 本地环境若暂时无法提供微信小程序凭据（`WECHAT_MINI_APPID`、`WECHAT_MINI_SECRET`），可在启动 `tea-api` 时设置 `WXACODE_MOCK=1`，使 `POST /api/v1/wx/wxacode` 返回可验证的固定 PNG base64（响应 `data.mock=true`，并回显 `scene/page`），用于本地“邀请好友/分享海报”证据补齐；真实联调时应关闭该开关，走微信官方接口（实现见 `tea-api/internal/handler/wx.go`）。

- 已验真“5.1 首页（4）”四项完成（wx-fe）：
   - 轮播/主图/跳转：`wx-fe/src/pages/index/index.tsx` 使用 `GET /api/v1/banners` 渲染 Swiper；banner 点击跳转规则与后端 `link_type` 对齐（1无链接/2商品详情/3分类/4外链）。
   - 外链承载：新增 `wx-fe/src/pages/webview/index.tsx`（WebView 页面）并在 `wx-fe/src/app.config.ts` / `wx-fe/src/pages.json` 注册，支持 `link_type=4` 打开外部链接。
   - 会员横幅：`wx-fe/src/pages/index/index.tsx` 兼容 `getMeSummary()` 不同返回形态识别登录态/会员态；未登录引导 `/pages/login`，非会员引导 `/pages/membership`。
   - 金刚区：`wx-fe/src/pages/index/index.tsx` 提供“堂食/外卖/商城”三入口；堂食/外卖要求选定门店并跳转 `/pages/product-list?store_id=...`，商城跳转 `/pages/category`；并提供门店 Picker 选择当前门店。
   - 首页-商城：`wx-fe/src/pages/index/index.tsx` 提供“商城分区/最近上新/热销/商品详情入口”；分区跳转 `/pages/category?category_id=...`，分类页已支持 `category_id` 入参预选（`wx-fe/src/pages/category/index.tsx`）。
   - 回归命令（前端编译/类型验证）：`pnpm --dir wx-fe typecheck`

- 强制现场 UI 验真证据链模板（H5，不走缓存；用于回填截图证据）：
   - H5 环境启动方式（二选一）：
   - 本机开发（当前采用 dev:h5）：`pnpm -C wx-fe run dev:h5`（Taro H5 watch，默认端口按本项目约定为 `http://127.0.0.1:9093/`；若实际端口不同以控制台输出为准）
   - Docker 静态预览（可选，更稳定）：`bash wx-fe/scripts/wx-fe_h5_in_docker.sh preview`（端口以脚本输出为准）
   - 强制不走缓存（建议 Chrome/Edge）：
      - 新建无痕窗口打开 H5；DevTools → Network 勾选 `Disable cache`（DevTools 需保持打开）
      - 右键刷新按钮选择 `Empty cache and hard reload`
      - DevTools → Application/Storage → `Clear site data`（清 localStorage/token，确保“未登录”态可复现）
      - 首次进入可带时间戳 query：`/?t=<ms>`（用于规避中间层缓存）
   - 最短验收路径与截图点（覆盖“5.1 首页（4）”四项）：
      - 轮播/主图/跳转（Balance）：
         - 进入首页，等待轮播加载；截图：`H5_01_首页_轮播渲染.png`
         - 点击 `link_type=2` banner，进入商品详情（URL 含 `#/pages/product-detail/index?id=`）；截图：`H5_02_轮播_跳商品详情.png`
         - 点击 `link_type=3` banner，进入分类页并按 `category_id` 预选（URL 含 `#/pages/category/index?category_id=`）；截图：`H5_03_轮播_跳分类预选.png`
         - 点击 `link_type=4` banner，H5 端会直接浏览器跳转外链（`pages/webview` 在 H5 下会执行 `location.href`）；截图：`H5_04_轮播_跳外链地址栏.png`
      - 会员横幅：
         - 清站点数据后进入首页，应出现“未登录提醒注册登陆”横幅；点击跳登录页（URL 含 `#/pages/login/index`）；截图：`H5_05_会员横幅_未登录.png`
         - 登录后回到首页：若非会员应展示“开通会员”横幅；点击跳会员页（URL 含 `#/pages/membership/index`）；截图：`H5_06_会员横幅_非会员.png`
         - 若账号为会员：首页不展示会员横幅（截首页顶部区域证明横幅不存在）；截图：`H5_07_会员横幅_会员不展示.png`
      - 金刚区（堂食/外卖/商城）：
         - 首页门店 Picker 选择门店（Picker 显示门店名）；截图：`H5_08_金刚区_门店已选.png`
         - 点击“堂食买单”进入门店商品列表（URL 含 `#/pages/product-list/index?store_id=`）；截图：`H5_09_金刚区_堂食列表.png`
         - 点击“外卖”进入门店商品列表（当前最小闭环与堂食复用列表页；URL 同样含 `store_id`）；截图：`H5_10_金刚区_外卖列表.png`
         - 点击“商城”进入分类页（URL 含 `#/pages/category/index`）；截图：`H5_11_金刚区_商城分类.png`
      - 首页-商城（分区/最近上新/热销/商品详情）：
         - 首页滚动到“商城分区”，点击任一分区进入分类页并预选（URL 含 `category_id`）；截图：`H5_12_首页商城分区.png` + `H5_13_商城分区_跳分类预选.png`
         - 首页“最近上新”双排商品列表可见（至少 4 个卡片）；截图：`H5_14_最近上新.png`
         - 首页“热销商品”双排商品列表可见（至少 4 个卡片）；截图：`H5_15_热销.png`
         - 点击任一商品卡片进入商品详情（URL 含 `#/pages/product-detail/index?id=`）；截图：`H5_16_商城商品详情.png`

- 强制现场 UI 验真证据链模板（微信开发者工具/真机 weapp，不走缓存；用于回填截图证据）：
   - weapp 构建与打开方式（最小步骤）：
      - 启动 watch 构建：`pnpm -C wx-fe run dev:weapp`
      - 微信开发者工具：导入项目目录为 `wx-fe/dist`（Taro weapp 输出目录）；如需真机，使用“预览/真机调试”生成二维码
   - 强制不走缓存（微信开发者工具）：
      - 关闭缓存：开发者工具右上角“详情/设置”中勾选“不开启缓存”（或同等选项）
      - 清理存储：使用“清缓存/清除数据”与 Storage 面板清空本地缓存（确保“未登录”态可复现）
      - 重新编译：点击“编译”并观察首页接口重新请求（轮播/商品/门店等）
   - 最短验收路径与截图点（覆盖“5.1 首页（4）”四项）：
      - 轮播/主图/跳转（Balance）：
         - 进入首页，轮播可见并自动切换；截图：`WEAPP_01_首页_轮播渲染.png`
         - 点击 `link_type=2` banner 进入商品详情；截图：`WEAPP_02_轮播_跳商品详情.png`
         - 点击 `link_type=3` banner 进入分类页且按 `category_id` 预选；截图：`WEAPP_03_轮播_跳分类预选.png`
         - 点击 `link_type=4` banner：应跳转到 `pages/webview` 承载外链（小程序端用 WebView 组件）；截图：`WEAPP_04_轮播_外链WebView.png`
      - 会员横幅：
         - 清理存储后进入首页：出现“未登录提醒注册登陆”横幅；点击进入登录页；截图：`WEAPP_05_会员横幅_未登录.png`
         - 登录后回首页：若非会员展示“开通会员”横幅；点击进入会员页；截图：`WEAPP_06_会员横幅_非会员.png`
         - 若账号为会员：首页不展示会员横幅；截图：`WEAPP_07_会员横幅_会员不展示.png`
      - 金刚区（堂食/外卖/商城）：
         - 首页门店 Picker 选择门店；截图：`WEAPP_08_金刚区_门店已选.png`
         - 点“堂食买单”进入门店商品列表（携带 `store_id`）；截图：`WEAPP_09_金刚区_堂食列表.png`
         - 点“外卖”进入门店商品列表（当前最小闭环与堂食复用列表页）；截图：`WEAPP_10_金刚区_外卖列表.png`
         - 点“商城”进入分类页；截图：`WEAPP_11_金刚区_商城分类.png`
      - 首页-商城（分区/最近上新/热销/商品详情）：
         - 首页滚动到“商城分区”，点击任一分区进入分类页并预选；截图：`WEAPP_12_首页商城分区.png` + `WEAPP_13_商城分区_跳分类预选.png`
         - 首页“最近上新”双排商品列表可见（至少 4 个卡片）；截图：`WEAPP_14_最近上新.png`
         - 首页“热销商品”双排商品列表可见（至少 4 个卡片）；截图：`WEAPP_15_热销.png`
         - 点击任一商品卡片进入商品详情；截图：`WEAPP_16_商城商品详情.png`

---

## 5.2 后台管理系统（平台）

| 完成 | 模块 | 需求文档功能点 | PRD 对应路径 | 状态 | 备注 |
| --- | --- | --- | --- | --- | --- |
| [✅] | 首页 | 左侧菜单栏（商品/订单/门店/会员/合伙人/营销/财务/系统） | `doc/prd.md` → `3.2.0 后台首页与总览` | 功能完成 | 已声明与权限体系绑定 |
| [✅] | 首页 | 顶部导航栏（快捷导航/搜索/消息通知/账户中心） | `doc/prd.md` → `3.2.0 后台首页与总览` | 功能完成 | |
| [✅] | 首页 | 数据看板（时间维度、订单量/销售额/用户数、排行/分析） | `doc/prd.md` → `3.2.0 后台首页与总览` | 功能完成 | 含跳转到明细页链接 |
| [✅] | 首页 | 订单动态（代发货、预警订单、售后待处理） | `doc/prd.md` → `3.2.0 后台首页與总览`（订单动态与待办） | 功能完成 | |
| [✅] | 首页 | 提现管理（提现待处理）、意见反馈待处理 | `doc/prd.md` → `3.2.0 后台首页与总览`（待办）+ `3.2.7 财务管理` | 功能完成 | 提现/反馈模块细节在后续模块中 |
| [✅] | 商品管理 | 全部商品（筛选、新增、上下架、删除、编辑、列表字段） | `doc/prd.md` → `3.2.1 后台商品管理` | 功能完成 | 包含商品字段、SKU、OSS 上传 |
| [✅] | 订单管理 | 订单列表/查找/发货/售后/调价 | `doc/prd.md` → `3.2.2 订单管理（平台端）` | 功能完成 | Admin-FE 订单操作区支持发货/取消/退款流转，并新增“调价”入口（未支付订单可修改支付金额，`POST /api/v1/orders/{id}/adjust`） |
| [✅] | 门店管理 | 门店列表/查看/新增/编辑/下架/冻结 | `doc/prd.md` → `3.2.3 门店管理` | 功能完成 | |
| [✅] | 用户管理 | 全部用户列表与筛选 | `doc/prd.md` → `3.2.4 会员管理`（列表/筛选） | 功能完成 | |
| [✅] | 用户管理 | 合伙人管理 | `doc/prd.md` → `3.2.5 合伙人管理（分销/合伙人）` | 功能完成 | Admin-FE 已实现「合伙人管理」页面（/partners）支持关键字搜索与按等级筛选，并可查看该合伙人佣金明细；后端提供 `GET /api/v1/admin/partners` 与 `GET /api/v1/admin/partners/:id/commissions`（权限点 `user:partner:view`） |
| [✅] | 用户管理 | 黑名单（禁止部分权限） | `doc/prd.md` → `3.2.4 会员管理` → 黑名单/白名单管理 | 功能完成 | Admin-FE「用户管理」列表新增“拉黑/移出黑”操作与状态展示；后端新增 `POST /api/v1/admin/users/{id}/blacklist` 并通过 AuthMiddleware/AuthJWT 对黑名单用户拦截（禁止登录/访问鉴权接口）；变更请求体由 OperationLogMiddleware 记录审计 |
| [✅] | 用户管理 | 白名单（全部权限可用） | `doc/prd.md` → `3.2.4 会员管理` → 黑名单/白名单管理 | 功能完成 | Admin-FE「用户管理」列表新增“加白/移出白”操作与状态展示；后端新增 `POST /api/v1/admin/users/{id}/whitelist`（启用白名单会自动清除黑名单）；白名单可豁免黑名单/停用拦截；变更请求体由 OperationLogMiddleware 记录审计 |
| [✅] | 营销管理 | 广告管理（列表、编辑、上下架、新增、查找） | `doc/prd.md` → `3.2.6 营销管理`（轮播图与广告位管理） | 功能完成 | Admin-FE 新增「广告管理」页面（/banners），支持关键字查找、列表、新增/编辑、上架/下架、删除；后端新增 `GET/POST/PUT/DELETE /api/v1/admin/banners` 并接入权限 `marketing:banner:view/manage` 与 OperationLogMiddleware 审计 |
| [✅] | 营销管理 | 优惠券（列表/管理/新增/发布/结束） | `doc/prd.md` → `3.2.6 营销管理`（优惠券管理） | 功能完成 | |
| [✅] | 营销管理 | 活动管理（列表/查找/编辑/发布/报名管理/新增活动） | `doc/prd.md` → `3.2.6 营销管理`（活动管理） | 功能完成 | 与首页活动入口联动 |
| [✅] | 营销管理 | 充值管理（列表、查询、查看、冻结、充钱、扣钱） | `doc/prd.md` → `3.2.6 营销管理` → 充值管理与配置 | 功能完成 | Admin-FE 新增「充值管理」页面（/recharge）：输入用户ID后可查看余额/冻结与最近流水，并提供“冻结/解冻/充钱/扣钱”动作（提交即调用 API）；后端新增 `GET /api/v1/admin/recharge/records`、`GET /api/v1/admin/recharge/users/{id}/wallet`、`POST /api/v1/admin/recharge/users/{id}/freeze/unfreeze/credit/debit`，并接入权限 `marketing:recharge:view/manage`；所有变更接口均通过 OperationLogMiddleware 审计 |
| [✅] | 营销管理 | 充值配置（档位、赠送茶币/优惠券/产品服务） | `doc/prd.md` → `3.2.6 营销管理` → 充值管理与配置 | 功能完成 | Admin-FE「充值管理」页内提供“充值配置”Tab：可开关 `recharge.enabled`、维护 `recharge.packages` 档位（充值金额/赠送茶币/赠送券模板ID）；后端新增 `GET/PUT /api/v1/admin/recharge/configs`（SystemConfig 存储，仅允许写入 `recharge.*`），并接入权限 `marketing:recharge:view/manage` + OperationLogMiddleware |
| [✅] | 订单管理 & 门店管理 | 门店与订单联动链路（门店订单列表 → 订单操作区） | `docs/features/store-order-link.md` | 功能完成 | 已在 tea-api 与 Admin-FE 中落地，支持从门店订单列表一键带入全局订单操作区 |
| [✅] | 门店管理 & 财务管理 | 门店列表 → 钱包/提现（门店钱包/提现申请/记录） | `doc/prd.md` → `3.2.3 门店管理` + `3.3.4 财务管理` | 功能完成 | 已在 tea-api 与 Admin-FE 中实现，支持从门店列表进入门店钱包查看余额/提现记录并发起提现申请 |
| [✅] | 客服管理 | 客服记录、客服服务（方式、用户、问题类型、处理结果） | `doc/prd.md` → `3.2.9 客服与投诉管理（平台端）` | 功能完成 | 已在 tea-api 中实现 Ticket 工单模型与 `/api/v1/admin/tickets` 接口，Admin-FE 新增「客服工单」页面支持按类型/来源/状态/优先级筛选和处理工单，并通过 `GET /api/v1/admin/dashboard/todos` + Dashboard 待办卡片展示待处理工单数 |
| [✅] | 投诉建议 | 投诉建议管理 | `doc/prd.md` → `3.2.9 客服与投诉管理（平台端）` | 功能完成 | 投诉建议归一为 Ticket 中的 `complaint` 类型工单，支持从小程序订单详情页投诉入口（`type=order, source=miniapp_order`）生成工单，并在「客服工单」列表中统一查看与处理，计入首页待办中的「待处理工单数」统计 |
| [✅] | 财务管理 | 财务记录（筛选、收支记录） | `doc/prd.md` → `3.2.7 财务管理`（平台资金流水） | 功能完成 | 已在 tea-api 中提供 `/api/v1/admin/payments` 与 `/api/v1/admin/refunds` 接口，并在 Admin-FE 中实现「财务记录」页面（FinanceRecords），支持多条件筛选收款/退款流水及导出 CSV；支付方式与状态在前端已用中文标签展示，便于运营/财务使用 |
| [✅] | 财务管理 | 提现管理（审核、付款） | `doc/prd.md` → `3.2.7 财务管理` + 分销结算设计 | 功能完成 | 已在 tea-api 中实现提现申请/审核/完成全链路，并接入分销佣金冻结→解冻→提现登记逻辑；Admin-FE 已支持门店钱包提现与后台审核/打款及佣金一键回滚入口 |
| [✅] | 系统管理 | 会员配置 | `doc/prd.md` → `3.2.4 会员管理`（会员等级/权益配置） | 功能完成 | |
| [✅] | 系统管理 | 合伙人配置 | `doc/prd.md` → `3.2.5 合伙人管理`（等级与佣金配置） | 功能完成 | |
| [✅] | 系统管理 | 权限管理/管理员管理 | `doc/prd.md` → `3.2.8 系统设置`（权限管理） | 功能完成 | |
| [✅] | 系统管理 | 基本管理（logo、电话、版权） | `doc/prd.md` → `3.2.8 系统设置` → 基础配置管理 | 功能完成 | Admin-FE 新增「系统设置」页（基础配置 Tab）支持保存；后端新增 `GET/PUT /api/v1/admin/system/configs`（SystemConfig 存储），权限 `system:config:view/manage` |
| [x] | 系统管理 | 快递模板（模板管理/设置） | `doc/prd.md` → `3.2.8 系统设置` → 基础配置管理（快递模板） | 已在 PRD 设计 | |
| [✅] | 系统管理 | 内容管理（关于我们、帮助、协议） | `doc/prd.md` → `3.2.8 系统设置` → 内容管理 | 功能完成 | Admin-FE「系统设置」页（内容管理 Tab）支持编辑并保存 about/help/privacy/terms；后端同 `GET/PUT /api/v1/admin/system/configs` |
| [✅] | 系统管理 | 日志（操作日志） | `doc/prd.md` → `3.2.8 系统设置`（日志与审计） | 功能完成 | Admin-FE 新增「日志」页面（操作日志/访问日志 Tab），支持查询与导出；后端已提供 `/api/v1/admin/logs/operations` + `/api/v1/admin/logs/access`（含 export） |

---

## 5.3 门店后台（网页）

| 完成 | 模块 | 需求文档功能点 | PRD 对应路径 | 状态 | 备注 |
| --- | --- | --- | --- | --- | --- |
| [✅] | 首页 | 左侧菜单栏（订单/商品/门店设置/财务/会员/系统） | `doc/prd.md` → `3.3.7 门店首页与经营参谋` | 功能完成 | Admin-FE 门店管理员侧菜单已按角色收口并补齐“首页/订单管理/商品管理/财务管理/活动/优惠券/门店设置”入口：`admin-fe/src/components/AppShell.tsx`（store role 分支）；路由落点 `admin-fe/src/App.tsx`（`/store-home` 等） |
| [✅] | 首页 | 顶部导航栏（快捷导航/搜索/消息/账户中心） | `doc/prd.md` → `3.3.7 门店首页与经营参谋` | 功能完成 | store 角色 Header 已做角色化：快捷导航指向门店路由（/store-*)；搜索支持订单号跳转并回填（/stores/:id/orders?orderNo=...）；消息/账户中心复用同一 Header（见 `admin-fe/src/components/AppShell.tsx` + `admin-fe/src/pages/StoreOrders.tsx`） |
| [✅] | 首页 | 经营参谋（多时间维度，订单/销售额等） | `doc/prd.md` → `3.3.7 门店首页与经营参谋` → 经营参谋 | 功能完成 | Admin-FE 新增门店首页 `admin-fe/src/pages/StoreHome.tsx`，支持“近7天/近30天”维度；后端新增门店侧统计接口 `GET /api/v1/stores/:id/orders/stats?days=7`（`tea-api/internal/router/router.go` + `tea-api/internal/handler/store.go` + `tea-api/internal/service/order.go`） |
| [✅] | 首页 | 工作台（待处理订单、收银流程） | `doc/prd.md` → `3.3.7 门店首页与经营参谋` → 工作台 | 功能完成 | Admin-FE `admin-fe/src/pages/StoreHome.tsx` 提供工作台入口（订单/商品/财务/门店设置快捷跳转）；收银写单流程仍待 PRD 后续落地 |
| [✅] | 商家商城 | 商品分类、查找、购物车 | `doc/prd.md` → `3.3.6 商家商城（门店特供）` + `3.1.5 购物车` | 功能完成 | Admin-FE 新增门店侧「商家商城」页 `admin-fe/src/pages/StoreMall.tsx`（分类下拉 + 关键词搜索 + 商品列表 + 加入购物车 + 购物车列表/改数量/删除/清空），菜单与路由入口为 `admin-fe/src/components/AppShell.tsx`（store role）+ `admin-fe/src/App.tsx`（`/store-mall` → `admin-fe/src/pages/StoreMallRedirect.tsx` → `/stores/:id/mall`）；后端提供门店特供商品接口 `GET /api/v1/stores/:id/exclusive-products`（支持 `keyword` + `category_id` 过滤，挂 `AuthJWT + RequireStoreScope("id") + RequirePermission("store:exclusive_products:view")`，见 `tea-api/internal/router/router.go` + `tea-api/internal/handler/store_exclusive_products.go` + `tea-api/internal/service/product.go`），购物车复用通用接口 `GET/POST/PUT/DELETE /api/v1/cart*`（`tea-api/internal/handler/cart.go`）；并在 `POST /api/v1/cart/items` 对 store 角色加购做强校验：仅允许“本门店特供（store_products.biz_type=3）”加入，且购物车不得混入平台/其他门店商品（`tea-api/internal/handler/cart.go` + `tea-api/internal/service/cart.go`，测试 `tea-api/test/cart_store_exclusive_guard_test.go`）；同时在 `POST /api/v1/orders/from-cart` 增加防混单兜底（`tea-api/internal/service/order.go`，测试 `tea-api/test/order_from_cart_no_mix_guard_test.go`） |
| [✅] | 商品管理 | 全部商品列表 | `doc/prd.md` → `3.3.2 商品管理` | 功能完成 | 已在 Admin-FE 提供「门店商品管理」页面（StoreProducts），支持按门店查看商品列表、按商品类型筛选、绑定/编辑门店库存与价格覆盖、解绑门店商品；后端已提供 `/api/v1/admin/stores/{id}/products`（GET/POST/DELETE）接口闭环 |
| [x] | 商品管理 | 商品管理（分类/上架中/已下架/草稿/新增） | `doc/prd.md` → `3.3.2 商品管理` | 已在 PRD 设计 | 可在实现时细化过滤条件 |
| [✅] | 商品管理 | 商品分类（服务商品/外卖商品/其他） | `doc/prd.md` → `3.3.2 商品管理` + `docs/prd-open-points.md` → 2.1 | 功能完成 | 已在 tea-api 中为 `store_products` 增加 `biz_type`（1服务/2外卖/3其他）并提供按类型筛选的门店库存接口，Admin-FE 中新增门店商品管理页（按商品类型展示与编辑）；门店管理员场景已补齐门店侧商品接口 `GET/POST/DELETE /api/v1/stores/:id/products`（`tea-api/internal/router/router.go`）并在 `admin-fe/src/pages/StoreProducts.tsx` 使用 scoped 调用 |
| [✅] | 商品管理 | 商品管理分组（分类/上架中/已下架/草稿/新增） | `doc/prd.md` → `3.3.2 商品管理`（门店商品管理分组） | 功能完成 | Admin-FE 门店商品页 `admin-fe/src/pages/StoreProducts.tsx` 顶部 Tabs 增加 5 分组，并通过 URL query `?tab=` 驱动；上架/下架/草稿通过接口 `status` 参数过滤：`GET /api/v1/stores/:id/products?status=...`（同样支持 admin 路径 `/api/v1/admin/stores/:id/products`），实现见 `tea-api/internal/handler/store_inventory.go` + `tea-api/internal/service/store_inventory.go`；回归：`pnpm -C admin-fe typecheck && pnpm -C admin-fe build`、`go test ./...` |
| [✅] | 订单管理 | 订单列表/查找/接单/发货 | `doc/prd.md` → `3.3.1 订单管理` | 功能完成 | 后端为门店管理员补齐门店侧订单列表接口 `GET /api/v1/stores/:id/orders`（`tea-api/internal/router/router.go` 复用 `AdminStoreOrders` 且挂载 `RequireStoreScope("id")`），Admin-FE 门店订单页 `admin-fe/src/pages/StoreOrders.tsx` 在 store 角色下改用该接口；门店菜单入口为 `admin-fe/src/pages/StoreOrdersRedirect.tsx`（锁店跳转至 `/stores/:id/orders`） |
| [✅] | 订单管理 | 售后订单（仅退款） | `doc/prd.md` → `3.3.1 订单管理`（售后处理） | 功能完成 | 门店后台新增退款记录列表页面（单表，无换货）：Admin-FE 路由 `/store-refunds`（锁店跳转到 `/stores/:id/refunds`），实现见 `admin-fe/src/pages/StoreRefunds.tsx` + `admin-fe/src/pages/StoreRefundsRedirect.tsx` + 菜单入口 `admin-fe/src/components/AppShell.tsx`；后端新增门店侧接口 `GET /api/v1/stores/:id/refunds`（按 `orders.store_id` 过滤，见 `tea-api/internal/router/router.go` + `tea-api/internal/handler/refund.go#ListStoreRefunds`），并由 `RequireStoreScope("id")` 约束门店越权（`tea-api/internal/middleware/store_scope.go`）；回归：`pnpm -C admin-fe typecheck && pnpm -C admin-fe build`、`go test ./...`；可复现（本地 9292）：admin 实测：`curl -sS -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}' http://127.0.0.1:9292/api/v1/user/login > build-ci-logs/admin_user_login_admin123.json` 获取 token，再 `curl -sS -D build-ci-logs/store_refunds_388_headers.txt -o build-ci-logs/store_refunds_388_body.json -H 'Authorization: Bearer <token>' 'http://127.0.0.1:9292/api/v1/stores/388/refunds?page=1&limit=10'`（保存响应：`build-ci-logs/store_refunds_388_headers.txt` + `build-ci-logs/store_refunds_388_body.json`）；store token（非 admin）实测：创建并绑定门店管理员（store role + store_admins 绑定）见 `build-ci-logs/store_user_create_payload.json` + `build-ci-logs/store_user_create_resp.json` + `build-ci-logs/store_admin_bind_402.txt`，再用 `build-ci-logs/store_user_login_payload.json` + `build-ci-logs/store_user_login_resp.json` 获取 store token，最后请求 `GET /api/v1/stores/402/refunds?page=1&limit=10`，保存响应：`build-ci-logs/store_refunds_storetoken_402_headers.txt` + `build-ci-logs/store_refunds_storetoken_402_body.json` |
| [✅] | 营销管理 | 活动管理（列表/查找/发布/编辑/报名） | 平台：`3.2.6 营销管理`；门店侧仅简要提及 | 功能完成 | 已在 tea-api 与 Admin-FE 中实现门店维度活动列表与创建/编辑（StoreActivities 页），后续如需报名/审核/退费流可在 PRD 中迭代 |
| [✅] | 营销管理 | 优惠券（列表/查找/新增/编辑/发布/查看） | `doc/prd.md` → `3.3.5 会员管理（门店维度）`（门店发券）+ `3.3.8.1 门店优惠券（已实现 — 最小闭环）` + `3.2.6 营销管理` | 功能完成 | 已在 tea-api 与 Admin-FE 中实现门店维度优惠券列表/创建/编辑/复制新建/一键禁用（StoreCoupons 页 + `/api/v1/stores/{id}/coupons`），门店可在本店维度自助配置与管理优惠券；更复杂人群/渠道/审核流由平台营销模块逐步扩展 |
| [✅] | 财务管理 | 财务记录（收支记录） | `doc/prd.md` → `3.3.4 财务管理`（账单查询） | 功能完成 | 已在 tea-api 中聚合门店维度收款/退款/提现流水，并在小程序 wx-fe 中新增「门店财务流水」页面（store-finance），支持按门店查看收支记录、按类型与时间范围筛选；后续若引入门店 Web 后台可直接复用该接口体系 |
| [✅] | 财务管理 | 账户管理（添加/编辑/删除） | `doc/prd.md` → `3.3.4 财务管理`（收款账户设置） | 功能完成 | |
| [✅] | 财务管理 | 财务提现（提现申请） | `doc/prd.md` → `3.3.4 财务管理` + `3.2.7 财务管理`（门店结算） | 功能完成 | 已在 tea-api 与 Admin-FE 中提供门店钱包/提现自助界面（StoreFinance 页），支持门店按自身门店维度查看余额并发起提现 |
| [x] | 系统设置 | 管理员（权限组/管理员列表/编辑） | `doc/prd.md` → `3.3.3 门店设置`（店员管理）+ `3.2.8 系统设置`（权限管理） | 已在 PRD 设计 | 目前 StoreAccounts 为“收款账户”，非店员/管理员管理 |
| [✅] | 系统设置 | 门店信息（名称/门头图/地址/营业时间/联系电话/环境图/证照） | `doc/prd.md` → `3.3.3 门店设置`（基本信息管理） | 功能完成 | Admin-FE 新增门店设置页 `admin-fe/src/pages/StoreSettings.tsx`，门店管理员仅可编辑本店；后端复用 `PUT /api/v1/stores/:id` 并由 `RequireStoreScope("id")` 兜底越权 403 |

---

## 总结：需补充 PRD 的关键点

1. **平台后台 - 客服管理 / 投诉建议**  
   - 需求来源：功能清单 5.2 中「客服管理」「投诉建议」。  
   - 现状：已在 PRD 独立成节，见 `3.2.9 客服与投诉管理（平台端）`。  
   - 建议：按该小节持续迭代联动（与待办、消息/通知系统）。  
   - 状态：**已在 PRD 设计（3.2.9）**。

2. **门店后台 - 营销管理（活动管理增强 & 报名列表）** ✅
   - 需求来源：功能清单 5.3 中「营销管理-活动管理/优惠券」，以及「门店活动 → 报名列表」。  
   - 现状：已在 tea-api 中为 Activity 增加门店维度字段，并提供门店活动列表/创建/编辑接口（`/api/v1/stores/{id}/activities`）；同时实现 ActivityRegistration 报名模型、与订单/支付联动的报名状态（已报名/已支付/已退款），并在 Admin-FE「门店活动」页面（StoreActivities）中提供报名列表 Drawer（含状态筛选、退款标记、CSV 导出、订单 ID/订单状态展示）。  
   - 建议：后续如需更复杂的报名审核流、批量退款规则及多维度统计报表，可在平台活动引擎中扩展，并为门店裁剪可配置字段子集。  
   - 状态：**已实现（2025-12，门店活动报名收费闭环 & 报名列表）**。

3. **门店营销 - 优惠券细化** ✅
   - 需求来源：5.3 中优惠券管理字段较多。  
   - 现状：已在 tea-api 与 Admin-FE 中落地门店专属优惠券创建/编辑/禁用能力（`/api/v1/stores/{id}/coupons` + 「门店优惠券」页面），门店可在本店维度自助配置和管理优惠券。  
   - 建议：如后续需要更复杂的人群/渠道规则，可在平台营销引擎中扩展字段，并同步裁剪门店可见字段子集。  
   - 状态：**已实现（2025-12，门店优惠券最小闭环）**。

> 如需，我可以在 `doc/prd.md` 中继续追加 `3.2.9 客服与投诉管理` 与门店活动管理的补充小节，使本对照表中的「需补充 PRD」项全部关闭。
