import type { Order } from '../services/types';

let cachedList: Order[] = [];
const ordersById = new Map<number, Order>();

export function cacheOrders(list: Order[]) {
  cachedList = Array.isArray(list) ? list.slice() : [];
  cachedList.forEach(o => ordersById.set(o.id, o));
}

export function getCachedOrders(): Order[] {
  return cachedList.slice();
}

export function cacheOrder(order: Order) {
  if (!order) return;
  ordersById.set(order.id, order);
}

export function getCachedOrder(id: number): Order | null {
  return ordersById.get(id) || null;
}
