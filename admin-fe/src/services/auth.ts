import api, { setToken, unwrapResponse } from './api';
import { User, ApiResponse } from './types';

export interface LoginPayload {
  code?: string; // wechat code
  username?: string;
  password?: string;
  openid?: string; // dev
}

export interface LoginResponse {
  token: string;
  user?: User;
}

export async function login(payload: LoginPayload): Promise<LoginResponse> {
  const res = await api.post('/api/v1/user/login', payload);
  const data = unwrapResponse<LoginResponse>(res);
  if ((data as any)?.token) setToken((data as any).token);
  return data;
}

export async function devLogin(openid: string): Promise<LoginResponse> {
  const res = await api.post('/api/v1/user/dev-login', { openid });
  const data = unwrapResponse<LoginResponse>(res);
  if ((data as any)?.token) setToken((data as any).token);
  return data;
}

export async function getUserInfo(): Promise<User> {
  const res = await api.get('/api/v1/user/info');
  return unwrapResponse<User>(res);
}

export async function refreshToken(): Promise<{ token: string } | null> {
  const res = await api.post('/api/v1/user/refresh');
  const data = unwrapResponse<{ token: string }>(res);
  if (data?.token) setToken(data.token);
  return data;
}
