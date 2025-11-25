示例页面说明（微信端）

包含两个演示页面（用于在小程序框架或 Taro/React Native 环境中快速验证 services）：

- `product-list`：产品列表页面，示例使用 `listProducts` 服务，列出商品并提供“加入购物车”按钮（示例仅打印日志）。
- `order-detail`：订单详情页面，示例使用 `getOrder` 服务，展示订单字段。

如何运行（示例说明）

1. 将 `wx-fe/src/pages/product-list` 与 `wx-fe/src/pages/order-detail` 集成到你的页面路由中（Taro/uni-app 或 RN 小程序桥接）。
2. 确保 `wx-fe/src/services` 中的 `api.ts`、`products.ts`、`orders.ts` 已存在并能工作。
3. 在模拟器或小程序真机上访问页面验证接口调用。

运行示例（Taro 项目）

1. 在 `wx-fe` 下安装依赖并启动（示例，依据你的项目管理工具）：

```powershell
cd e:\project\tea\wx-fe
npm install
npx taro dev:weapp
```

2. 页面配置：`wx-fe/src/app.config.ts` 已包含示例页面 `product-list` 与 `order-detail`，可以在 Taro 项目中直接运行。

3. 若使用其他目标（H5 / RN），可参考 Taro 文档把 `app.config.ts` 转换为目标对应的页面配置。
