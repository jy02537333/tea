import axios from 'axios';
import Taro from '@tarojs/taro';

declare const WX_API_BASE_URL: string | undefined;

const DEFAULT_BASE_URL = (() => {
  try {
    const loc = (globalThis as any)?.location;
    const host = loc?.hostname;
    const protocol = loc?.protocol;
    if (typeof host === 'string' && host.trim()) {
      const proto = typeof protocol === 'string' && protocol.startsWith('https') ? 'https' : 'http';
      return `${proto}://${host}:9292`;
    }
  } catch (_) {}
  return 'http://127.0.0.1:9292';
})();

function normalizeBaseUrl(input: string): string {
  const raw = (input || '').trim();
  if (!raw) return DEFAULT_BASE_URL;

  try {
    const u = new URL(raw);
    // On Linux, `host.docker.internal` is often not resolvable from the browser.
    // If a build baked it in, rewrite to the current page hostname.
    if (u.hostname === 'host.docker.internal') {
      const loc = (globalThis as any)?.location;
      const pageHost = loc?.hostname;
      if (typeof pageHost === 'string' && pageHost.trim()) {
        u.hostname = pageHost;
      }
    }
    return u.toString();
  } catch (_) {
    return raw;
  }
}

const BASE_URL = normalizeBaseUrl(
  (typeof WX_API_BASE_URL !== 'undefined' && WX_API_BASE_URL) || DEFAULT_BASE_URL,
);

export const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
});

export function getToken(): string | null {
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

const existingToken = getToken();
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
