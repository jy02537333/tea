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
  rest.post('/api/v1/user/dev-login', (req, res, ctx) => {
    return res(ctx.json({ token: 'dev-token', user: { id: 1, nickname: 'dev' } }));
  }),
];
