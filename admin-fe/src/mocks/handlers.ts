import { http, HttpResponse } from 'msw';

export const handlers = [
  http.get('/api/v1/products', async (_req) => {
    return HttpResponse.json({
      data: [{ id: 1, name: '示例茶', price: 12.5, stock: 100 }],
      total: 1,
      page: 1,
      limit: 20,
    });
  }),
  http.post('/api/v1/user/dev-login', async (_req) => {
    return HttpResponse.json({ token: 'dev-token', user: { id: 1, nickname: 'dev' } });
  }),
];
