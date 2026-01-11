import { api, PaginatedResult, unwrapPagination } from './api';
import type { Product } from './products';

export interface StoreExclusiveProduct extends Product {
  store_stock?: number | null;
  store_price_override?: string | null;
}

export interface StoreExclusiveProductListParams {
  page?: number;
  limit?: number;
  keyword?: string;
  category_id?: number;
}

export async function getStoreExclusiveProducts(
  storeId: number,
  params: StoreExclusiveProductListParams
): Promise<PaginatedResult<StoreExclusiveProduct>> {
  const res = await api.get(`/api/v1/stores/${storeId}/exclusive-products`, { params });
  return unwrapPagination<StoreExclusiveProduct>(res);
}
