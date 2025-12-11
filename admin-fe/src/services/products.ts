import { api, PaginatedResult, unwrap, unwrapPagination } from './api';
import type { Category } from './categories';

export interface Product {
  id: number;
  name: string;
  description?: string;
  category_id: number;
  category?: Category;
  price: string;
  original_price?: string;
  images?: string;
  status: number;
  stock: number;
  sales?: number;
  sort?: number;
  is_hot?: boolean;
  is_new?: boolean;
  is_recommend?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface ProductListParams {
  page?: number;
  limit?: number;
  category_id?: number;
  status?: number | string;
  keyword?: string;
}

export interface ProductPayload {
  name: string;
  description?: string;
  category_id: number;
  price: number;
  original_price?: number;
  images?: string;
  status: number;
  stock?: number;
  sort?: number;
  is_hot?: boolean;
  is_new?: boolean;
  is_recommend?: boolean;
}

export interface UpdateStockPayload {
  stock: number;
  action: 'add' | 'sub' | 'set';
}

export async function getProducts(params: ProductListParams): Promise<PaginatedResult<Product>> {
  const res = await api.get('/api/v1/products', { params });
  return unwrapPagination<Product>(res);
}

export async function getProduct(id: number) {
  const res = await api.get(`/api/v1/products/${id}`);
  return unwrap<Product>(res);
}

export async function createProduct(payload: ProductPayload) {
  const res = await api.post('/api/v1/products', payload);
  return unwrap<Product>(res);
}

export async function updateProduct(id: number, payload: ProductPayload) {
  const res = await api.put(`/api/v1/products/${id}`, payload);
  return unwrap<Product>(res);
}

export async function deleteProduct(id: number) {
  const res = await api.delete(`/api/v1/products/${id}`);
  return unwrap(res);
}

export async function updateProductStock(id: number, payload: UpdateStockPayload) {
  const res = await api.put(`/api/v1/products/${id}/stock`, payload);
  return unwrap(res);
}

// 兼容旧版示例页面 `ProductList` 所使用的 listProducts，
// 返回形如 { data, total, page, limit } 的对象，便于示例代码抽取 data 数组。
export async function listProducts(params: ProductListParams) {
  const page = await getProducts(params);
  return {
    data: page.list,
    total: page.total,
    page: page.page,
    limit: page.limit,
  };
}
