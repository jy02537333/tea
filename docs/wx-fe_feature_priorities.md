# wx-fe 功能优先级与验收清单

> 目的：明确小程序（wx-fe）侧的近期开发优先级，依据 PRD MVP 路线梳理页面/接口与验收要点，保障首版可用路径尽快闭环。

---

## 范围与基线

- 技术栈：Taro 4 + React 18（H5/微信小程序双端构建）
- 入口与 TabBar：见 `src/app.config.ts`（首页/购物车/我的），页面目录见 `src/pages/*`
- API 基址：`WX_API_BASE_URL | VITE_API_BASE_URL | 127.0.0.1:9292`（详见 `src/services/api.ts`）

---

## MVP 优先级（P0 → P1）

- [P0] 购物车 → 下单 → 支付 → 订单详情
  - 页面：`pages/cart`、`pages/checkout`、`pages/order-detail`、`pages/orders`
  - 接口：`GET/POST /api/v1/cart/*`、`POST /api/v1/orders/from-cart`、`POST /api/v1/payments/unified-order`、订单查询/取消/确认收货
  - 验收：能完成一次下单并在订单详情中展示金额与状态；取消/确认收货可操作；可用券列表可展示并在下单中应用。

- [P0] 门店列表与门店点单入口（最小闭环）
  - 页面：`pages/index`（门店选择与商品展示）、`pages/category`、`pages/product-list`、`pages/product-detail`
  - 接口：`GET /api/v1/stores`、`GET /api/v1/products`（支持 `store_id`、筛选/搜索）
  - 验收：可选择门店后浏览商品，并将商品加入购物车；门店维度下单路径可用。

- [P1] 我的聚合页与核心条目（钱包/积分/优惠券/订单）
  - 页面：`pages/profile`、`pages/wallet`、`pages/points`、`pages/coupons`、`pages/orders`
  - 接口：`GET /api/v1/users/me/summary`（聚合视图），钱包/积分/优惠券只读接口
  - 验收：我的页能展示头像、昵称、账户与聚合数量；各子页能正常分页与查看。

- [P1] 会员套餐购买与生效
  - 页面：`pages/membership`、`pages/membership-orders`
  - 接口：`GET /api/v1/membership-packages`、`POST /api/v1/membership-orders`、支付链路；购买后 `GET /api/v1/users/me/summary` 生效
  - 验收：能选择套餐生成订单并完成支付模拟；我的页显示会员等级与权益字段。

---

## 非MVP（P2+，按迭代推进）

- 门店财务流水与收款账户（商家侧自助）
  - 页面：`pages/store-finance`、`pages/store-accounts`
  - 接口：门店维度收支记录与账户配置；与后台配置联动

- 活动报名与列表（门店维度）
  - 页面：`pages/activities`
  - 接口：门店活动列表/报名/订单映射；报名状态与退款联动

- 客服/意见反馈与帮助
  - 页面：`pages/feedback`、`pages/help`
  - 接口：`POST /api/v1/tickets` 创建工单；帮助由系统内容管理提供

- 售后入口与流程
  - 页面：`pages/after-sale`
  - 接口：售后申请/进度查询，与订单维度联动

---

## 验收与自测建议

- 运行方式（H5 开发预览）：

```bash
pnpm -C wx-fe run dev:h5
```

- 运行方式（微信小程序构建）：

```bash
pnpm -C wx-fe run build:weapp
```

- 环境变量：`WX_API_BASE_URL` 或 `VITE_API_BASE_URL` 指向 `tea-api`（建议维持 `9292`）。
- Token 注入：登录后由 `Taro.setStorageSync('token')` 持久化；也可在 H5 模式用 `localStorage.token` 自助注入用于联调。

- 自测路径：
  - 首页选择门店 → 分类/商品列表 → 加车 → 结算 → 订单详情
  - 我的 → 优惠券/钱包/积分只读页打开 → 订单列表分页 → 订单取消/确认收货
  - 会员套餐页 → 生成会员订单 → 模拟支付后查看我的聚合生效

---

## 下一步任务（建议）

1. 打通订单支付模拟（H5 下本地联调）并固定订单状态文案映射。
2. 完善 `pages/checkout` 用券选择与金额展示一致性，确保与后台/脚本金额模型一致。
3. 我的聚合视图一致性核对（字段命名、单位/精度、空态文案）。
4. 门店维度商品筛选与搜索体验优化（关键字、价格/销量排序）— 基础筛选与排序控件已就绪（产地/包装/价格区间/排序），可继续在后端完善参数支持与服务端排序。
5. 补最小 UI 回归脚本（可复用 Playwright H5 模式）用于首页→下单→订单详情路径的烟雾测试。
