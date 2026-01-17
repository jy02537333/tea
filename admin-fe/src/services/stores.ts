import { api, PaginatedResult, unwrap, unwrapPagination } from './api';
import type { Coupon } from './types';
import type { Product } from './products';

export interface Store {
  id: number;
  name: string;
  address?: string;
  phone?: string;
  latitude?: number;
  longitude?: number;
  business_hours?: string;
  images?: string;
  status?: number;
  created_at?: string;
  updated_at?: string;
}

export interface StoreListParams {
  page?: number;
  limit?: number;
  status?: number;
}

export interface StorePayload {
  name: string;
  address?: string;
  phone?: string;
  latitude?: number;
  longitude?: number;
  business_hours?: string;
  images?: string;
  status?: number;
}

export interface StoreOrderStats {
  store_id: number;
  total_orders: number;
  completed_amount: string;
  status_counts: { status: number; count: number }[];
}

export interface StoreWalletSummary {
  store_id: number;
  total_paid: string;
  total_refunded: string;
  total_withdrawn: string;
  available: string;
}

export interface StoreWithdrawRecord {
  id: number;
  user_id: number;
  store_id: number;
  withdraw_no: string;
  amount: string;
  fee: string;
  actual_amount: string;
  withdraw_type: number;
  status: number;
  remark?: string;
  processed_at?: string | null;
  processed_by?: number;
  created_at?: string;
}

// 门店资金流水（聚合支付/退款/提现）
export interface StoreFinanceTransaction {
  id: number;
  store_id: number;
  type: 'payment' | 'refund' | 'withdraw';
  direction: 'in' | 'out';
  amount: string; // 金额（元）
  fee: string; // 手续费（元），支付/退款为0
  related_id: number; // 关联订单ID或0
  related_no: string; // 支付/退款/提现单号
  method: number; // 支付方式（支付/退款有效）
  remark?: string;
  created_at?: string;
}

export interface StoreProduct {
  id: number;
  store_id: number;
  product_id: number;
  stock: number;
  price_override: string;
  biz_type: number; // 1:服务 2:外卖 3:其他
  created_at?: string;
  updated_at?: string;
  product?: Product;
}

export interface StoreProductListParams {
  page?: number;
  limit?: number;
  biz_type?: number;
  status?: number;
}

export interface StoreActivity {
  id: number;
  store_id?: number;
  name: string;
  type: number;
  start_time: string;
  end_time: string;
  rules?: string;
  status: number;
  priority?: number;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface StoreActivityRegistration {
  id: number;
  store_id: number;
  activity_id: number;
  user_id: number;
  user_name?: string;
  user_phone?: string;
  order_id?: number | null;
  order_status?: number;
  order_pay_status?: number;
  status: number; // 1已报名（待支付/免费） 2已支付报名 3已退款
  fee: string;
  refund_amount: string;
  refund_reason?: string;
  refunded_at?: string | null;
  created_at?: string;
}

export interface StoreBankAccount {
  id: number;
  store_id: number;
  account_type: string;
  account_name: string;
  account_no: string;
  bank_name?: string;
  is_default: boolean;
  created_at?: string;
}

export async function listStoreCoupons(id: number, params?: { status?: number }) {
  const res = await api.get(`/api/v1/stores/${id}/coupons`, { params });
  // 后端当前返回的是列表而非分页
  return unwrap<Coupon[]>(res);
}

export interface StoreActivityPayload {
  name: string;
  type: number; // 1限时折扣 2满减活动 3买赠活动
  start_time: string; // RFC3339
  end_time: string; // RFC3339
  rules?: string;
  status: number;
  priority?: number;
  description?: string;
}

export async function listStoreActivities(id: number, params?: { status?: number }) {
  const res = await api.get(`/api/v1/stores/${id}/activities`, { params });
  return unwrap<StoreActivity[]>(res);
}

export async function createStoreActivity(id: number, payload: StoreActivityPayload) {
  const res = await api.post(`/api/v1/stores/${id}/activities`, payload);
  return unwrap<StoreActivity>(res);
}

export async function updateStoreActivity(id: number, activityId: number, payload: StoreActivityPayload) {
  const res = await api.put(`/api/v1/stores/${id}/activities/${activityId}`, payload);
  return unwrap<StoreActivity>(res);
}

export async function getStoreActivityRegistrations(
  storeId: number,
  activityId: number,
  params?: { page?: number; limit?: number; status?: number }
) {
  const res = await api.get(`/api/v1/stores/${storeId}/activities/${activityId}/registrations`, { params });
  return unwrapPagination<StoreActivityRegistration>(res);
}

export async function refundStoreActivityRegistration(
  storeId: number,
  activityId: number,
  registrationId: number,
  payload: { reason?: string }
) {
  const res = await api.post(
    `/api/v1/stores/${storeId}/activities/${activityId}/registrations/${registrationId}/refund`,
    payload
  );
  return unwrap<StoreActivityRegistration>(res);
}

export interface StoreCouponPayload {
  name: string;
  type: number; // 1满减 2折扣 3免单
  amount?: number;
  discount?: number;
  min_amount?: number;
  total_count: number;
  status: number;
  start_time: string; // RFC3339
  end_time: string; // RFC3339
  description?: string;
}

function buildCouponRequestBody(payload: StoreCouponPayload) {
  return {
    name: payload.name,
    type: payload.type,
    amount: payload.amount != null ? String(payload.amount) : '',
    discount: payload.discount != null ? String(payload.discount) : '',
    min_amount: payload.min_amount != null ? String(payload.min_amount) : '',
    total_count: payload.total_count,
    status: payload.status,
    start_time: payload.start_time,
    end_time: payload.end_time,
    description: payload.description ?? '',
  };
}

export async function createStoreCoupon(id: number, payload: StoreCouponPayload) {
  const body = buildCouponRequestBody(payload);
  const res = await api.post(`/api/v1/stores/${id}/coupons`, body);
  return unwrap<Coupon>(res);
}

export async function updateStoreCoupon(id: number, couponId: number, payload: StoreCouponPayload) {
  const body = buildCouponRequestBody(payload);
  const res = await api.put(`/api/v1/stores/${id}/coupons/${couponId}`, body);
  return unwrap<Coupon>(res);
}

export async function getStores(params: StoreListParams): Promise<PaginatedResult<Store>> {
  const res = await api.get('/api/v1/stores', { params });
  return unwrapPagination<Store>(res);
}

export async function getStore(id: number) {
  const res = await api.get(`/api/v1/stores/${id}`);
  return unwrap<Store>(res);
}

export async function createStore(payload: StorePayload) {
  const res = await api.post('/api/v1/stores', payload);
  return unwrap<Store>(res);
}

export async function updateStore(id: number, payload: StorePayload) {
  const res = await api.put(`/api/v1/stores/${id}`, payload);
  return unwrap(res);
}

export async function deleteStore(id: number) {
  const res = await api.delete(`/api/v1/stores/${id}`);
  return unwrap(res);
}

export async function getStoreOrderStats(id: number) {
  const res = await api.get(`/api/v1/admin/stores/${id}/orders/stats`);
  return unwrap<StoreOrderStats>(res);
}

// 门店后台：门店维度订单统计（支持 days 参数）
export async function getStoreOrderStatsScoped(id: number, params?: { days?: number }) {
  const res = await api.get(`/api/v1/stores/${id}/orders/stats`, { params });
  return unwrap<StoreOrderStats>(res);
}

export async function getStoreWallet(id: number) {
  const res = await api.get(`/api/v1/stores/${id}/wallet`);
  return unwrap<StoreWalletSummary>(res);
}

export async function getStoreWithdraws(id: number, params?: { page?: number; limit?: number; status?: number }) {
  const res = await api.get(`/api/v1/stores/${id}/withdraws`, { params });
  return unwrapPagination<StoreWithdrawRecord>(res);
}

export async function applyStoreWithdraw(id: number, payload: { amount: number; remark?: string; withdraw_type?: number }) {
  const res = await api.post(`/api/v1/stores/${id}/withdraws`, payload);
  return unwrap<StoreWithdrawRecord>(res);
}

export async function listStoreProducts(id: number, params: StoreProductListParams) {
  const res = await api.get(`/api/v1/admin/stores/${id}/products`, { params });
  return unwrapPagination<StoreProduct>(res);
}

// 门店后台：门店维度商品列表（门店管理员使用）
export async function listStoreProductsScoped(id: number, params: StoreProductListParams) {
  const res = await api.get(`/api/v1/stores/${id}/products`, { params });
  return unwrapPagination<StoreProduct>(res);
}

// 门店收款账户
export async function listStoreAccounts(storeId: number) {
  const res = await api.get(`/api/v1/stores/${storeId}/accounts`);
  return unwrap<StoreBankAccount[]>(res);
}

export async function createStoreAccount(
  storeId: number,
  payload: Partial<Pick<StoreBankAccount, 'account_type' | 'account_name' | 'account_no' | 'bank_name' | 'is_default'>>,
) {
  const res = await api.post(`/api/v1/stores/${storeId}/accounts`, payload);
  return unwrap<StoreBankAccount>(res);
}

export async function updateStoreAccount(
  storeId: number,
  accountId: number,
  payload: Partial<Pick<StoreBankAccount, 'account_type' | 'account_name' | 'account_no' | 'bank_name' | 'is_default'>>,
) {
  const res = await api.put(`/api/v1/stores/${storeId}/accounts/${accountId}`, payload);
  return unwrap(res);
}

export async function deleteStoreAccount(storeId: number, accountId: number) {
  const res = await api.delete(`/api/v1/stores/${storeId}/accounts/${accountId}`);
  return unwrap(res);
}

export async function upsertStoreProduct(
  id: number,
  payload: { product_id: number; stock: number; price_override?: string; biz_type?: number }
) {
  const res = await api.post(`/api/v1/admin/stores/${id}/products`, payload);
  return unwrap<StoreProduct>(res);
}

// 门店后台：门店维度商品绑定/编辑（门店管理员使用）
export async function upsertStoreProductScoped(
  id: number,
  payload: { product_id: number; stock: number; price_override?: string; biz_type?: number }
) {
  const res = await api.post(`/api/v1/stores/${id}/products`, payload);
  return unwrap<StoreProduct>(res);
}

export async function deleteStoreProduct(id: number, productId: number) {
  const res = await api.delete(`/api/v1/admin/stores/${id}/products/${productId}`);
  return unwrap(res);
}

// 门店后台：门店维度商品解绑（门店管理员使用）
export async function deleteStoreProductScoped(id: number, productId: number) {
  const res = await api.delete(`/api/v1/stores/${id}/products/${productId}`);
  return unwrap(res);
}

// 门店资金流水列表（分页）
export async function getStoreFinanceTransactions(
  id: number,
  params?: { page?: number; limit?: number; start?: string; end?: string; type?: 'payment' | 'refund' | 'withdraw' | '' }
) {
  const res = await api.get(`/api/v1/stores/${id}/finance/transactions`, { params });
  return unwrapPagination<StoreFinanceTransaction>(res);
}

// 门店资金流水导出（CSV）
export async function exportStoreFinanceTransactions(
  id: number,
  params?: { start?: string; end?: string; type?: 'payment' | 'refund' | 'withdraw' | '' }
) {
  const res = await api.get(`/api/v1/stores/${id}/finance/transactions/export`, {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}

// 门店桌号管理
export interface StoreTable {
  id: number;
  store_id: number;
  table_no: string;
  capacity?: number;
  status?: number;
  note?: string;
  created_at?: string;
}

export async function listStoreTables(id: number) {
  const res = await api.get(`/api/v1/stores/${id}/tables`);
  return unwrap<StoreTable[]>(res);
}

export async function createStoreTable(id: number, payload: { table_no: string; capacity?: number; note?: string }) {
  const res = await api.post(`/api/v1/stores/${id}/tables`, payload);
  return unwrap<StoreTable>(res);
}

export async function deleteStoreTable(id: number, tableId: number) {
  const res = await api.delete(`/api/v1/stores/${id}/tables/${tableId}`);
  return unwrap(res);
}

// 门店自有商品创建（特供，仅该门店可用）
export async function createStoreExclusiveProduct(
  id: number,
  payload: { name: string; category_id: number; price: number; description?: string; images?: string; stock?: number; price_override?: number }
) {
  const body = {
    name: payload.name,
    category_id: payload.category_id,
    price: payload.price.toFixed(2),
    description: payload.description ?? '',
    images: payload.images ?? '',
    stock: payload.stock ?? 0,
    price_override: payload.price_override != null ? payload.price_override.toFixed(2) : '',
  };
  const res = await api.post(`/api/v1/stores/${id}/exclusive-products/new`, body);
  return unwrap<StoreProduct>(res);
}
