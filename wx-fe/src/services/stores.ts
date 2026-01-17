import api, { unwrapResponse } from './api';
import { getProducts } from './products';
import { Store, PaginationResponse, Product } from './types';

export interface StoreFinanceQuery {
  start?: string;
  end?: string;
  type?: string; // payment | refund | withdraw
}

export interface StoreFinanceTransaction {
  id: number;
  store_id?: number;
  type: string;
  direction: string;
  amount: string | number;
  fee?: string | number;
  related_id?: number;
  related_no?: string;
  method?: number;
  remark?: string;
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

export async function listStores(params: any = {}): Promise<PaginationResponse<Store>> {
  // 默认仅拉取可用门店（status=1）；允许调用方显式覆盖
  const finalParams = { status: 1, ...params };
  const res = await api.get('/api/v1/stores', { params: finalParams });
  const data = unwrapResponse<PaginationResponse<Store>>(res);
  // 过滤禁用门店（status=0 不展示）
  if (data && Array.isArray(data.data)) {
    data.data = data.data.filter((s: any) => {
      const st = typeof s?.status === 'number' ? s.status : undefined;
      return st === undefined || st !== 0;
    });
  }
  return data;
}

export async function getStore(id: number): Promise<Store> {
  const res = await api.get(`/api/v1/stores/${id}`);
  return unwrapResponse<Store>(res);
}

// 门店特供（商家商城）商品列表
export async function listStoreExclusiveProducts(
  storeId: number,
  params: { page?: number; limit?: number; keyword?: string } = {},
): Promise<PaginationResponse<Product>> {
  const res = await api.get(`/api/v1/stores/${storeId}/exclusive-products`, { params });
  return unwrapResponse<PaginationResponse<Product>>(res);
}

// 智能拉取门店商品：优先尝试 exclusive-products；若无门店权限(code=1003/401/403)，则回退到公开商品接口 products?store_id=
export async function listStoreProductsSmart(
  storeId: number,
  params: { page?: number; limit?: number; keyword?: string } = {},
): Promise<PaginationResponse<Product>> {
  try {
    const res = await api.get(`/api/v1/stores/${storeId}/exclusive-products`, { params });
    return unwrapResponse<PaginationResponse<Product>>(res);
  } catch (e: any) {
    const code = e?.response?.data?.code;
    const status = e?.response?.status;
    const noPerm = code === 1003 || status === 401 || status === 403;
    if (!noPerm) throw e;
    const resp = await getProducts({ page: params.page, limit: params.limit, keyword: params.keyword, store_id: storeId });
    return resp;
  }
}

// 门店收款账户列表
export async function listStoreAccounts(storeId: number): Promise<StoreBankAccount[]> {
  const res = await api.get(`/api/v1/stores/${storeId}/accounts`);
  return unwrapResponse<StoreBankAccount[]>(res);
}

// 新增门店收款账户
export async function createStoreAccount(
  storeId: number,
  payload: Partial<Pick<StoreBankAccount, 'account_type' | 'account_name' | 'account_no' | 'bank_name' | 'is_default'>>,
): Promise<StoreBankAccount> {
  const res = await api.post(`/api/v1/stores/${storeId}/accounts`, payload);
  return unwrapResponse<StoreBankAccount>(res);
}

// 更新门店收款账户
export async function updateStoreAccount(
  storeId: number,
  accountId: number,
  payload: Partial<Pick<StoreBankAccount, 'account_type' | 'account_name' | 'account_no' | 'bank_name' | 'is_default'>>,
): Promise<{ ok: boolean }> {
  const res = await api.put(`/api/v1/stores/${storeId}/accounts/${accountId}`, payload);
  return unwrapResponse<{ ok: boolean }>(res);
}

// 删除门店收款账户
export async function deleteStoreAccount(storeId: number, accountId: number): Promise<{ ok: boolean }> {
  const res = await api.delete(`/api/v1/stores/${storeId}/accounts/${accountId}`);
  return unwrapResponse<{ ok: boolean }>(res);
}

// 门店资金流水列表
export async function listStoreFinanceTransactions(
  storeId: number,
  query: StoreFinanceQuery & { page?: number; limit?: number } = {},
): Promise<PaginationResponse<StoreFinanceTransaction>> {
  const res = await api.get(`/api/v1/stores/${storeId}/finance/transactions`, { params: query });
  return unwrapResponse<PaginationResponse<StoreFinanceTransaction>>(res);
}

// 导出门店资金流水（按当前筛选条件）
// 小程序场景：仅向后端发起导出请求，不处理文件下载
export async function exportStoreFinanceTransactions(storeId: number, query: StoreFinanceQuery = {}): Promise<void> {
  await api.get(`/api/v1/stores/${storeId}/finance/transactions/export`, {
    params: query,
    responseType: 'blob',
  });
}
