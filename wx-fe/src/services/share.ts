import api, { unwrapResponse } from './api';

export async function bindReferral(referrerId: number): Promise<void> {
  try {
    const res = await api.post('/api/v1/referrals/bind', { referrer_id: referrerId });
    unwrapResponse(res);
  } catch (_) {
    // 非关键操作，忽略错误（如未登录或后端尚未实现）
  }
}

export function buildShareLink(params: { referrerId: number; storeId?: number }): string {
  const base = typeof window !== 'undefined' ? window.location.origin : '';
  const qs = new URLSearchParams();
  qs.set('referrer_id', String(params.referrerId));
  if (params.storeId) qs.set('store_id', String(params.storeId));
  // 最小版：落到首页
  return `${base}/pages/index/index?${qs.toString()}`;
}

export async function getWxaCode(params: { scene: string; page: string; width?: number; is_hyaline?: boolean }): Promise<string> {
  const res = await api.post('/api/v1/wx/wxacode', params);
  const data = unwrapResponse<{ image_base64?: string }>(res);
  return data?.image_base64 || '';
}
