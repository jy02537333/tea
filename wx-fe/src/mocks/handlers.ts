import { rest } from 'msw';

export const handlers = [
  rest.get('/api/v1/products', (req, res, ctx) => {
    return res(
      ctx.json({
        data: [{ id: 1, name: '示例茶', price: 12.5, stock: 100 }],
        total: 1,
        page: 1,
        limit: 20,
      })
    );
  }),
];
