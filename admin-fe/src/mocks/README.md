MSW Mock Setup (示例)

说明：此目录包含 MSW（Mock Service Worker）的示例配置与 handler 模板。若希望在本地启用 mock：

1. 安装依赖：`npm install msw --save-dev`
2. 在 `src/mocks/browser.ts` 中初始化 worker 并在开发入口（如 `src/main.tsx`）中调用 `worker.start()`。
3. 在 `src/mocks/handlers.ts` 中定义针对关键 API 的 handler（products / orders / auth）。

示例 handlers（TypeScript 风格）：

```ts
import { rest } from 'msw';

export const handlers = [
  rest.get('/api/v1/products', (req, res, ctx) => {
    return res(ctx.json({ data: [{ id: 1, name: '示例茶', price: 12.5 }], total: 1, page: 1, limit: 20 }));
  }),
  rest.post('/api/v1/user/dev-login', (req, res, ctx) => {
    return res(ctx.json({ token: 'dev-token', user: { id: 1, nickname: 'dev' } }));
  }),
];
```

注意：实际启用前请根据项目入口调整 `baseURL` 与 worker 挂载位置。
