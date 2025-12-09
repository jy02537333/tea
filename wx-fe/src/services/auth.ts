import api, { setToken, unwrapResponse } from './api';
import { AuthResponse, User } from './types';

export async function login(payload: { code?: string; username?: string; password?: string; openid?: string }): Promise<AuthResponse> {
  const res = await api.post('/api/v1/user/login', payload);
  const data = unwrapResponse<AuthResponse>(res);
  if (data?.token) setToken(data.token as string);
  return data;
}

export async function getUserInfo(): Promise<User> {
  const res = await api.get('/api/v1/user/info');
  return unwrapResponse<User>(res);
}

export async function updateUserInfo(updates: Partial<Pick<User, 'nickname' | 'avatar'>> & { gender?: number }): Promise<string | undefined> {
  const res = await api.put('/api/v1/user/info', updates);
  return unwrapResponse<string | undefined>(res);
}
