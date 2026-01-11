import api, { unwrapResponse } from './api';
import type { MeSummary, User } from './types';

// 优先使用统一聚合视图；如后端暂不支持则回退到 /api/v1/user/info 并做最小映射
export async function getMeSummary(): Promise<MeSummary> {
  try {
    const res = await api.get('/api/v1/users/me/summary');
    const raw = unwrapResponse<any>(res);
    // 兼容不同后端返回结构：
    // 1) 期望：{ user: {...}, wallet: {...}, points: {...}, coupons: {...}, share: {...} }
    // 2) 简版：{ user_id, nickname, wallet_balance, points, coupons, membership }
    if (raw && typeof raw === 'object') {
      // 完整结构直接返回
      if (raw.user && typeof raw.user === 'object') {
        return raw as MeSummary;
      }
      // 简版结构，尝试映射
      const uid = raw.user_id ?? raw.uid ?? raw.id;
      if (typeof uid !== 'undefined') {
        const user: User = {
          id: Number(uid) || 0,
          nickname: raw.nickname ?? raw.name ?? '',
          phone: raw.phone,
          avatar: raw.avatar,
          role: raw.role || 'user',
        };
        const wallet_balance = raw.wallet_balance;
        const wallet = { balance_cents: typeof wallet_balance === 'number' ? Math.round(wallet_balance * 100) : undefined };
        const points = { balance: typeof raw.points === 'number' ? raw.points : undefined };
        const coupons = { available_count: typeof raw.coupons === 'number' ? raw.coupons : undefined };
        return { user, wallet, points, coupons } as MeSummary;
      }
      // 无法解析用户字段，补打一条 info 获取用户对象
      try {
        const infoRes = await api.get('/api/v1/user/info');
        const u = unwrapResponse<any>(infoRes);
        const user: User = {
          id: Number(u.id ?? u.uid ?? 0) || 0,
          nickname: u.nickname ?? u.name ?? '',
          phone: u.phone,
          avatar: u.avatar,
          role: u.role || 'user',
        };
        return { user } as MeSummary;
      } catch (_) {}
    }
    return raw as MeSummary;
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
