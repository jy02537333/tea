// Minimal mock API to satisfy admin-fe Auth + Orders page
// Routes:
//  POST /api/v1/user/dev-login -> { data: { token } }
//  GET  /api/v1/user/info      -> { data: { id, role } }
//  GET  /api/v1/admin/rbac/user-permissions?user_id=1 -> { data: [permissions] }
//  GET  /api/v1/admin/orders   -> { data: [], total, page, limit }
const http = require('http');
const url = require('url');

const port = Number(process.env.MOCK_API_PORT || 9292);

function send(res, status, payload, headers = {}) {
  res.statusCode = status;
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  for (const k of Object.keys(headers)) res.setHeader(k, headers[k]);
  res.end(JSON.stringify(payload));
}

// in-memory fixtures
let withdraws = [
  {
    id: 7001,
    withdraw_no: 'WD-MOCK-0001',
    user_id: 101,
    amount: '200.00',
    fee: '2.00',
    actual_amount: '198.00',
    status: 'pending',
    requested_at: new Date().toISOString(),
    processed_at: null,
  },
];

let orders = [
  // 可调价：待付款 + 未支付
  {
    id: 10000,
    order_no: 'MOCK-20250101-0000',
    store_id: 1,
    user_id: 1,
    pay_amount: 66.66,
    status: 1, // 待付款
    pay_status: 1, // 未支付
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  // 可退款/管理员取消：已付款 + 已支付
  {
    id: 10001,
    order_no: 'MOCK-20250101-0001',
    store_id: 1,
    user_id: 1,
    pay_amount: 128.88,
    status: 2, // 已付款
    pay_status: 2, // 已支付
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const server = http.createServer(async (req, res) => {
  const parsed = url.parse(req.url || '/', true);
  const { pathname, query } = parsed;

  // Simple CORS for browser XHR if needed
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  res.setHeader('Access-Control-Allow-Methods', 'GET,POST,PUT,DELETE,OPTIONS');
  if (req.method === 'OPTIONS') return send(res, 200, { ok: true });

  if (req.method === 'POST' && pathname === '/api/v1/user/dev-login') {
    return send(res, 200, { data: { token: 'mock-token' } });
  }
  if (req.method === 'GET' && pathname === '/api/v1/user/info') {
    return send(res, 200, { data: { id: 1, role: 'admin', nickname: 'Mock Admin' } });
  }
  if (req.method === 'GET' && pathname === '/api/v1/admin/rbac/user-permissions') {
    // grant all to simplify UI, include adjust
    return send(res, 200, { data: ['order:deliver', 'order:complete', 'order:cancel', 'order:refund', 'order:adjust'] });
  }
  if (req.method === 'GET' && pathname === '/api/v1/admin/withdraws') {
    const page = Number(query.page || 1);
    const limit = Number(query.limit || 20);
    return send(res, 200, { data: withdraws, page, limit, total: withdraws.length });
  }
  if (req.method === 'POST' && pathname && /\/api\/v1\/admin\/withdraws\/(\d+)\/approve$/.test(pathname)) {
    const id = Number((pathname.match(/(\d+)/) || [])[0]);
    withdraws = withdraws.map(w => w.id === id ? { ...w, status: 'approved', processed_at: new Date().toISOString() } : w);
    return send(res, 200, { data: { ok: true } });
  }
  if (req.method === 'POST' && pathname && /\/api\/v1\/admin\/withdraws\/(\d+)\/reject$/.test(pathname)) {
    const id = Number((pathname.match(/(\d+)/) || [])[0]);
    withdraws = withdraws.map(w => w.id === id ? { ...w, status: 'rejected', processed_at: new Date().toISOString() } : w);
    return send(res, 200, { data: { ok: true } });
  }
  if (req.method === 'GET' && pathname === '/api/v1/admin/orders') {
    const page = Number(query.page || 1);
    const limit = Number(query.limit || 20);
    return send(res, 200, { data: orders, total: orders.length, page, limit });
  }

  // Admin: list store orders
  if (req.method === 'GET' && pathname && /\/api\/v1\/admin\/stores\/(\d+)\/orders$/.test(pathname)) {
    const storeId = Number((pathname.match(/(\d+)/) || [])[0]);
    const page = Number(query.page || query.page_size ? query.page || 1 : 1);
    const limit = Number(query.page_size || query.limit || 20);
    const statusFilter = query.status ? Number(query.status) : undefined;
    let list = orders.filter(o => o.store_id === storeId);
    if (typeof statusFilter === 'number' && !Number.isNaN(statusFilter)) {
      list = list.filter(o => o.status === statusFilter);
    }
    return send(res, 200, { data: list, total: list.length, page, limit });
  }

  if (req.method === 'GET' && pathname && pathname.startsWith('/api/v1/admin/orders/')) {
    const id = Number(pathname.split('/').pop());
    const ord = orders.find(o => o.id === id) || orders[1];
    const payload = {
      order: ord,
      items: [
        { id: 1, product_id: 1, product_name: '测试商品A', sku_name: '默认', quantity: 1, price: ord.pay_amount, amount: ord.pay_amount },
      ],
    };
    return send(res, 200, { data: payload });
  }

  // Orders action endpoints with reason validation
  if (req.method === 'POST' && pathname?.startsWith('/api/v1/orders/')) {
    // parse body
    let body = '';
    await new Promise(resolve => { req.on('data', chunk => (body += chunk)); req.on('end', resolve); });
    let json = {};
    try { json = body ? JSON.parse(body) : {}; } catch {}

    // adjust path
    if (/\/adjust$/.test(pathname)) {
      const id = Number((pathname.match(/(\d+)/) || [])[0]);
      const amt = Number(json?.new_pay_amount);
      if (!Number.isFinite(amt) || amt < 0) {
        return send(res, 400, { error: 'invalid_amount' });
      }
      orders = orders.map(o => o.id === id ? { ...o, pay_amount: amt, updated_at: new Date().toISOString() } : o);
      return send(res, 200, { data: { ok: true } });
    }

    // endpoints that require reason
    if (/\/(admin-cancel|refund|refund\/start|refund\/confirm)$/.test(pathname)) {
      const reason = (json?.reason ?? '').toString().trim();
      if (!reason) {
        return send(res, 400, { error: 'reason_required' });
      }
      return send(res, 200, { data: { ok: true } });
    }

    // default ok
    return send(res, 200, { data: { ok: true } });
  }

  return send(res, 404, { error: 'not_found', path: pathname });
});

server.listen(port, '127.0.0.1', () => {
  // eslint-disable-next-line no-console
  console.log(`[mock-admin-api] listening on http://127.0.0.1:${port}`);
});
