import api, { unwrapResponse } from './api';
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
  const res = await api.get('/api/v1/stores', { params });
  return unwrapResponse<PaginationResponse<Store>>(res);
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
