import axios from 'axios';
import Taro from '@tarojs/taro';

declare const WX_API_BASE_URL: string | undefined;

const DEFAULT_BASE_URL = 'http://127.0.0.1:8082';
const BASE_URL =
  (typeof WX_API_BASE_URL !== 'undefined' && WX_API_BASE_URL) ||
  process.env.WX_API_BASE_URL ||
  process.env.VITE_API_BASE_URL ||
  DEFAULT_BASE_URL;

export const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
});

function readToken(): string | null {
  try {
    const value = Taro.getStorageSync('token');
    if (value) return value;
  } catch (error) {
    // ignore storage errors in non-mini-program environments
  }
  if (typeof localStorage !== 'undefined') {
    return localStorage.getItem('token');
  }
  return null;
}

export function setToken(token: string | null) {
  try {
    if (token) Taro.setStorageSync('token', token);
    else Taro.removeStorageSync('token');
  } catch (error) {
    // ignore storage errors when running outside Taro runtime
  }

  if (typeof localStorage !== 'undefined') {
    if (token) localStorage.setItem('token', token);
    else localStorage.removeItem('token');
  }

  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    delete api.defaults.headers.common['Authorization'];
  }
}

const existingToken = readToken();
if (existingToken) {
  api.defaults.headers.common['Authorization'] = `Bearer ${existingToken}`;
}

export default api;

export function unwrapResponse<T>(res: any): T {
  if (res && res.data && typeof res.data === 'object' && 'data' in res.data) {
    return res.data.data as T;
  }
  if (res && typeof res.data !== 'undefined') {
    return res.data as T;
  }
  return res as T;
}
