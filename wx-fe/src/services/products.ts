import api, { unwrapResponse } from './api';
import { Product, PaginationResponse } from './types';

export interface ProductQueryParams {
  page?: number;
  limit?: number;
  category_id?: number;
  keyword?: string;
  store_id?: number;
  // 细化筛选与排序（后端支持则走服务端过滤；否则前端可做最小兜底）
  origin?: string; // 产地，例如 "华东/华南/西南..." 或具体产区名
  packaging?: string; // 包装：散装/袋装/罐装/礼盒
  min_price?: number; // 最低价（单位与后端约定）
  max_price?: number; // 最高价
  sort?: string; // 排序：price_asc | price_desc | sales_desc | default
}

export async function getProducts(params: ProductQueryParams = {}): Promise<PaginationResponse<Product>> {
  const res = await api.get('/api/v1/products', { params });
  return unwrapResponse<PaginationResponse<Product>>(res);
}

export async function getProduct(id: number, store_id?: number): Promise<Product> {
  const res = await api.get(`/api/v1/products/${id}`, { params: store_id ? { store_id } : {} });
  return unwrapResponse<Product>(res);
}

// alias for example pages
export const listProducts = getProducts;
