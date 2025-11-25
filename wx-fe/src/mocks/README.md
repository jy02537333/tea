MSW Mock Setup (示例)

说明：此目录包含 MSW（Mock Service Worker）的示例配置与 handler 模板，供 wx-fe 在 H5 环境或开发时使用。

启用步骤：
1. 安装：`npm install msw --save-dev`
2. 在 `src/mocks/browser.ts` 中初始化并在开发入口调用 `worker.start()`。

示例 handler：
```ts
import { rest } from 'msw';

export const handlers = [
  rest.get('/api/v1/products', (req, res, ctx) => {
    return res(ctx.json({ data: [{ id: 1, name: '示例茶', price: 12.5 }], total:1, page:1, limit:20 }));
  }),
];
```

提示：Taro 在真机小程序环境不使用 service worker，MSW 仅用于 H5 / 浏览器开发场景。
