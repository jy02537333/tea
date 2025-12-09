import api, { unwrapResponse } from './api';
import { Order, OrderDetailPayload, PaginationResponse } from './types';

export async function createOrderFromCart(payload: { delivery_type: number; address_info?: string; remark?: string; user_coupon_id?: number; store_id?: number; order_type?: number }): Promise<Order> {
  const res = await api.post('/api/v1/orders/from-cart', payload);
  return unwrapResponse<Order>(res);
}

export async function listOrders(params: { page?: number; limit?: number; status?: number; store_id?: number } = {}): Promise<PaginationResponse<Order>> {
  const res = await api.get('/api/v1/orders', { params });
  return unwrapResponse<PaginationResponse<Order>>(res);
}

export async function getOrder(id: number): Promise<OrderDetailPayload> {
  const res = await api.get(`/api/v1/orders/${id}`);
  const data = unwrapResponse<OrderDetailPayload>(res);
  return {
    order: data?.order as Order,
    items: Array.isArray(data?.items) ? data.items : [],
  };
}

export async function cancelOrder(orderId: number, reason?: string): Promise<void> {
  const res = await api.post(`/api/v1/orders/${orderId}/cancel`, { reason });
  unwrapResponse(res);
}

export async function payOrder(orderId: number): Promise<void> {
  const res = await api.post(`/api/v1/orders/${orderId}/pay`);
  unwrapResponse(res);
}

export async function confirmReceive(orderId: number): Promise<void> {
  const res = await api.post(`/api/v1/orders/${orderId}/receive`);
  unwrapResponse(res);
}
