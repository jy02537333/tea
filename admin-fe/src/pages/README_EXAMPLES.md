示例页面说明

包含两个演示页面（用于快速启动与验证 services）：

- `ProductList`：产品列表页面，示例使用 `listProducts` 服务，显示表格并提供“新建商品”按钮。
- `OrderDetail`：订单详情页面，示例使用 `getOrder` 服务，展示订单字段。

如何运行（假设已有 admin-fe React 项目）

1. 将 `admin-fe/src/pages/ProductList` 与 `admin-fe/src/pages/OrderDetail` 加入到你的路由配置。
2. 确保 `admin-fe/src/services` 中的 `api.ts`、`products.ts`、`orders.ts` 已存在并能工作。
3. 启动项目并访问相应路由进行验证。

运行示例（React / Vite 项目）

```powershell
cd e:\project\tea\admin-fe
npm install
npm run dev
```

将 `admin-fe/src/pages/ProductList` 与 `admin-fe/src/pages/OrderDetail` 添加到你的路由配置（React Router / Umi / AntD Pro 依据项目而定）。
