import api, { unwrapResponse } from './api';
import type { PaginationResponse } from './types';

export interface MembershipPackage {
  id: number;
  name: string;
  price: string | number;
  tea_coin_award?: string | number;
  discount_rate?: string | number;
  purchase_discount_rate?: string | number;
  direct_commission_rate?: string | number;
  team_commission_rate?: string | number;
  upgrade_reward_rate?: string | number;
  type?: string;
  created_at?: string;
  updated_at?: string;
}

export interface MembershipPackageQuery {
  page?: number;
  limit?: number;
  type?: string;
}

export async function listMembershipPackages(
  params: MembershipPackageQuery = {},
): Promise<PaginationResponse<MembershipPackage>> {
  const res = await api.get('/api/v1/membership-packages', { params });
  return unwrapResponse<PaginationResponse<MembershipPackage>>(res);
}

export interface CreateMembershipOrderPayload {
  package_id: number;
  remark?: string;
  sharer_uid?: number;
  share_store_id?: number;
}

export interface CreateMembershipOrderResult {
  order_id: number;
  order_no: string;
  pay_amount: number;
}

export async function createMembershipOrder(
  payload: CreateMembershipOrderPayload,
): Promise<CreateMembershipOrderResult> {
  const res = await api.post('/api/v1/membership-orders', payload);
  return unwrapResponse<CreateMembershipOrderResult>(res);
}
