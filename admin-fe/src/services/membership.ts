import { api, unwrapPagination, unwrapResponse, type PaginatedResult } from './api';

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

export interface PartnerLevel {
  id: number;
  name: string;
  purchase_discount_rate?: string | number;
  direct_commission_rate?: string | number;
  team_commission_rate?: string | number;
  upgrade_reward_rate?: string | number;
  note?: string;
  created_at?: string;
  updated_at?: string;
}

export interface MembershipQuery {
  page?: number;
  limit?: number;
  type?: string;
}

export async function listMembershipPackages(params: MembershipQuery = {}): Promise<PaginatedResult<MembershipPackage>> {
  const res = await api.get('/api/v1/admin/membership-packages', { params });
  return unwrapPagination<MembershipPackage>(res);
}

export type MembershipPackagePayload = {
  name: string;
  price: number;
  tea_coin_award?: number;
  discount_rate?: number;
  purchase_discount_rate?: number;
  direct_commission_rate?: number;
  team_commission_rate?: number;
  upgrade_reward_rate?: number;
  type?: string;
};

export async function createMembershipPackage(payload: MembershipPackagePayload): Promise<MembershipPackage> {
  const res = await api.post('/api/v1/admin/membership-packages', payload);
  return unwrapResponse<MembershipPackage>(res);
}

export async function updateMembershipPackage(id: number, payload: Partial<MembershipPackagePayload>): Promise<MembershipPackage> {
  const res = await api.put(`/api/v1/admin/membership-packages/${id}`, payload);
  return unwrapResponse<MembershipPackage>(res);
}

export async function deleteMembershipPackage(id: number): Promise<{ ok: boolean }> {
  const res = await api.delete(`/api/v1/admin/membership-packages/${id}`);
  return unwrapResponse<{ ok: boolean }>(res);
}

export async function listPartnerLevels(params: { page?: number; limit?: number } = {}): Promise<PaginatedResult<PartnerLevel>> {
  const res = await api.get('/api/v1/admin/partner-levels', { params });
  return unwrapPagination<PartnerLevel>(res);
}

export type PartnerLevelPayload = {
  name: string;
  purchase_discount_rate?: number;
  direct_commission_rate?: number;
  team_commission_rate?: number;
  upgrade_reward_rate?: number;
  note?: string;
};

export async function createPartnerLevel(payload: PartnerLevelPayload): Promise<PartnerLevel> {
  const res = await api.post('/api/v1/admin/partner-levels', payload);
  return unwrapResponse<PartnerLevel>(res);
}

export async function updatePartnerLevel(id: number, payload: Partial<PartnerLevelPayload>): Promise<PartnerLevel> {
  const res = await api.put(`/api/v1/admin/partner-levels/${id}`, payload);
  return unwrapResponse<PartnerLevel>(res);
}

export async function deletePartnerLevel(id: number): Promise<{ ok: boolean }> {
  const res = await api.delete(`/api/v1/admin/partner-levels/${id}`);
  return unwrapResponse<{ ok: boolean }>(res);
}
