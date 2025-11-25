import api, { unwrapResponse } from './api';
import { CartItem } from './types';

export async function listCart(): Promise<CartItem[]> {
  const res = await api.get('/api/v1/cart');
  return unwrapResponse<CartItem[]>(res);
}

export async function addCartItem(product_id: number, sku_id: number | null, quantity: number): Promise<CartItem> {
  const res = await api.post('/api/v1/cart/items', { product_id, sku_id, quantity });
  return unwrapResponse<CartItem>(res);
}

export async function updateCartItem(id: number, quantity: number): Promise<void> {
  const res = await api.put(`/api/v1/cart/items/${id}`, { quantity });
  return unwrapResponse<void>(res);
}

export async function removeCartItem(id: number): Promise<void> {
  const res = await api.delete(`/api/v1/cart/items/${id}`);
  return unwrapResponse<void>(res);
}

export async function clearCart(): Promise<void> {
  const res = await api.delete('/api/v1/cart/clear');
  return unwrapResponse<void>(res);
}
