import api, { unwrapResponse } from './api';
import { PaginationResponse, Product, ApiResponse } from './types';

export async function getProducts(params: { page?: number; limit?: number; category_id?: number; status?: number; keyword?: string; store_id?: number }): Promise<PaginationResponse<Product>> {
  const res = await api.get('/api/v1/products', { params });
  return unwrapResponse<PaginationResponse<Product>>(res);
}

export async function getProduct(id: number, store_id?: number): Promise<Product> {
  const res = await api.get(`/api/v1/products/${id}`, { params: store_id ? { store_id } : {} });
  return unwrapResponse<Product>(res);
}

export async function createProduct(payload: Partial<Product>): Promise<Product> {
  const res = await api.post('/api/v1/products', payload);
  return unwrapResponse<Product>(res);
}

export async function updateProduct(id: number, payload: Partial<Product>): Promise<void> {
  const res = await api.put(`/api/v1/products/${id}`, payload);
  // some endpoints return empty success; unwrap to keep types consistent
  return unwrapResponse<void>(res);
}

export async function deleteProduct(id: number): Promise<void> {
  const res = await api.delete(`/api/v1/products/${id}`);
  return unwrapResponse<void>(res);
}

export async function updateProductStock(id: number, stock: number, action: 'add' | 'sub' | 'set'): Promise<void> {
  const res = await api.put(`/api/v1/products/${id}/stock`, { stock, action });
  return unwrapResponse<void>(res);
}

// backward-compatible alias used by example pages
export const listProducts = getProducts;
