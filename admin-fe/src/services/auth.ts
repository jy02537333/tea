import { api, setToken, unwrap } from './api';
import { User } from '../types/user';

interface LoginResponse {
  token: string;
  user?: User;
}

export interface PasswordLoginPayload {
  username: string;
  password: string;
  captcha_id: string;
  captcha_code: string;
}

export interface CaptchaResponse {
  id: string;
  image: string;
}

export async function login(payload: PasswordLoginPayload) {
  const res = await api.post('/api/v1/user/login', payload);
  const body = res?.data ?? {};
  // 后端统一返回形态：{ code, message, data }
  if (typeof body.code === 'number' && body.code !== 0) {
    throw new Error(body.message || '登录失败');
  }
  const data = (body.data ?? unwrap<LoginResponse>(res)) as LoginResponse;
  if (data?.token) setToken(data.token);
  return data;
}

export async function devLogin(openid: string) {
  const res = await api.post('/api/v1/user/dev-login', { openid });
  const body = res?.data ?? {};
  if (typeof body.code === 'number' && body.code !== 0) {
    throw new Error(body.message || '登录失败');
  }
  const data = (body.data ?? unwrap<LoginResponse>(res)) as LoginResponse;
  if (data?.token) setToken(data.token);
  return data;
}

export async function getUserInfo() {
  const res = await api.get('/api/v1/user/info');
  return unwrap<User>(res);
}

export async function fetchCaptcha() {
  const res = await api.get('/api/v1/auth/captcha');
  return unwrap<CaptchaResponse>(res);
}
