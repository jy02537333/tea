import { api, unwrap, unwrapPagination, type PaginatedResult } from './api';

export interface WalletSummary {
  user_id: number;
  balance_cents: number;
  frozen_cents: number;
}

export interface WalletTx {
  id: number;
  user_id: number;
  type: string;
  amount_cents: number;
  balance_after_cents: number;
  remark?: string;
  created_at?: string;
}

export interface RechargeConfigItem {
  id: number;
  config_key: string;
  config_value: string;
  config_type?: string;
  description?: string;
  status?: number;
}

export interface RechargeRecordQuery {
  user_id?: number;
  types?: string;
  keyword?: string;
  start?: string;
  end?: string;
  page?: number;
  limit?: number;
}

export async function listRechargeRecords(params: RechargeRecordQuery = {}): Promise<PaginatedResult<WalletTx>> {
  const res = await api.get('/api/v1/admin/recharge/records', { params });
  return unwrapPagination<WalletTx>(res);
}

export async function getRechargeUserWallet(userId: number) {
  const res = await api.get(`/api/v1/admin/recharge/users/${userId}/wallet`);
  return unwrap<{ wallet: WalletSummary }>(res);
}

export async function freezeWallet(userId: number, payload: { amount_cents: number; remark?: string }) {
  const res = await api.post(`/api/v1/admin/recharge/users/${userId}/freeze`, payload);
  return unwrap<{ ok: boolean; wallet: WalletSummary }>(res);
}

export async function unfreezeWallet(userId: number, payload: { amount_cents: number; remark?: string }) {
  const res = await api.post(`/api/v1/admin/recharge/users/${userId}/unfreeze`, payload);
  return unwrap<{ ok: boolean; wallet: WalletSummary }>(res);
}

export async function creditWallet(userId: number, payload: { amount_cents: number; remark?: string }) {
  const res = await api.post(`/api/v1/admin/recharge/users/${userId}/credit`, payload);
  return unwrap<{ ok: boolean; wallet: WalletSummary }>(res);
}

export async function debitWallet(userId: number, payload: { amount_cents: number; remark?: string }) {
  const res = await api.post(`/api/v1/admin/recharge/users/${userId}/debit`, payload);
  return unwrap<{ ok: boolean; wallet: WalletSummary }>(res);
}

export async function listRechargeConfigs() {
  const res = await api.get('/api/v1/admin/recharge/configs');
  return unwrap<{ list: RechargeConfigItem[] }>(res);
}

export async function upsertRechargeConfigs(items: Array<{ config_key: string; config_value: string; config_type?: string; description?: string; status?: number }>) {
  const res = await api.put('/api/v1/admin/recharge/configs', { items });
  return unwrap<{ ok: boolean; updated: number }>(res);
}
