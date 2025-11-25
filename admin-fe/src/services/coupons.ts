import api, { unwrapResponse } from './api';
import { Coupon, PaginationResponse } from './types';

export async function listCoupons(params: any = {}): Promise<PaginationResponse<Coupon>> {
  const res = await api.get('/api/v1/coupons', { params });
  return unwrapResponse<PaginationResponse<Coupon>>(res);
}

export async function createCoupon(payload: Partial<Coupon>): Promise<Coupon> {
  const res = await api.post('/api/v1/coupons', payload);
  return unwrapResponse<Coupon>(res);
}

export async function grantCoupon(payload: any): Promise<Coupon> {
  const res = await api.post('/api/v1/coupons/grant', payload);
  return unwrapResponse<Coupon>(res);
}

export async function listMyCoupons(): Promise<Coupon[]> {
  const res = await api.get('/api/v1/user/coupons');
  return unwrapResponse<Coupon[]>(res);
}
