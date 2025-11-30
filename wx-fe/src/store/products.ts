import type { Product } from '../services/types';

const cache = new Map<number, Product>();

export function getCachedProduct(id: number): Product | null {
  return cache.get(id) || null;
}

export function cacheProduct(p: Product) {
  if (p && typeof p.id === 'number') cache.set(p.id, p);
}

export function cacheProducts(list: Product[]) {
  list?.forEach(cacheProduct);
}
