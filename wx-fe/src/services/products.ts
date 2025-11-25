import api, { unwrapResponse } from './api';
import { Product, PaginationResponse } from './types';

export async function getProducts(params: { page?: number; limit?: number; category_id?: number; keyword?: string; store_id?: number } = {}): Promise<PaginationResponse<Product>> {
  const res = await api.get('/api/v1/products', { params });
  return unwrapResponse<PaginationResponse<Product>>(res);
}

export async function getProduct(id: number, store_id?: number): Promise<Product> {
  const res = await api.get(`/api/v1/products/${id}`, { params: store_id ? { store_id } : {} });
  return unwrapResponse<Product>(res);
}

// alias for example pages
export const listProducts = getProducts;
