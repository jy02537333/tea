import api, { unwrapResponse } from './api';
import type { MeSummary } from './types';

// 优先使用统一聚合视图；如后端暂不支持则回退到 /api/v1/user/info 并做最小映射
export async function getMeSummary(): Promise<MeSummary> {
  try {
    const res = await api.get('/api/v1/users/me/summary');
    return unwrapResponse<MeSummary>(res);
  } catch (e) {
    try {
      const infoRes = await api.get('/api/v1/user/info');
      const u = unwrapResponse<any>(infoRes);
      return {
        user: {
          id: u?.id,
          uid: u?.uid,
          open_id: u?.open_id,
          nickname: u?.nickname,
          avatar: u?.avatar,
          phone: u?.phone,
          gender: u?.gender,
          balance: u?.balance,
          points: u?.points,
          role: u?.role,
        },
        wallet: { balance_cents: typeof u?.balance === 'number' ? Math.round(u.balance * 100) : undefined },
        points: { balance: u?.points },
      } as MeSummary;
    } catch (_ignored) {
      throw e;
    }
  }
}
