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
];
