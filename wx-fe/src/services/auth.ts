import api, { setToken, unwrapResponse } from './api';
import { AuthResponse, User } from './types';

export async function login(payload: { code?: string; username?: string; password?: string; openid?: string }): Promise<AuthResponse> {
  // 新接口（优先）：统一鉴权入口 /api/v1/auth/login
  try {
    const res = await api.post('/api/v1/auth/login', payload);
    const data = unwrapResponse<AuthResponse>(res);
    if (data?.token) setToken(data.token as string);
    return data;
  } catch (_e) {
    // 兼容旧接口：/api/v1/user/login
    const res = await api.post('/api/v1/user/login', payload);
    const data = unwrapResponse<AuthResponse>(res);
    if (data?.token) setToken(data.token as string);
    return data;
  }
}

export async function getUserInfo(): Promise<User> {
  // 新接口（优先）：用户聚合视图 /api/v1/users/me/summary
  try {
    const res = await api.get('/api/v1/users/me/summary');
    const summary: any = unwrapResponse<any>(res);
    const user: User | undefined = (summary?.user as User) || (summary?.data?.user as User);
    if (user) return user;
  } catch (_ignored) {
    // ignore and fallback
  }
  // 兼容旧接口：/api/v1/user/info
  const res = await api.get('/api/v1/user/info');
  return unwrapResponse<User>(res);
}

export async function updateUserInfo(updates: Partial<Pick<User, 'nickname' | 'avatar'>> & { gender?: number }): Promise<string | undefined> {
  const res = await api.put('/api/v1/user/info', updates);
  return unwrapResponse<string | undefined>(res);
}
