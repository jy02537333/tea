import axios from 'axios';

const runtimeConfig = typeof window !== 'undefined' ? (window as any).__TEA_RUNTIME_CONFIG__ : undefined;
const runtimeBaseUrl = runtimeConfig?.apiBaseUrl;
const envBaseUrl = (import.meta as any)?.env?.VITE_API_BASE_URL;
const inferredBaseUrl = typeof window !== 'undefined' ? `${window.location.protocol}//${window.location.hostname}:9292` : 'http://localhost:9292';
const BASE_URL = runtimeBaseUrl || envBaseUrl || inferredBaseUrl;

export const api = axios.create({
  baseURL: BASE_URL,
  timeout: 30000,
});

export function setToken(token: string | null) {
  if (token) {
    localStorage.setItem('token', token);
    api.defaults.headers.common.Authorization = `Bearer ${token}`;
  } else {
    localStorage.removeItem('token');
    delete api.defaults.headers.common.Authorization;
  }
}

const saved = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
if (saved) {
  setToken(saved);
}

api.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error?.response?.status === 401) {
      setToken(null);
      window.dispatchEvent(new CustomEvent('unauthorized'));
    }
    return Promise.reject(error);
  }
);

export function unwrap<T>(response: any): T {
  if (response?.data?.data) return response.data.data as T;
  return response?.data as T;
}

// 兼容旧代码中的 unwrapResponse 命名
export function unwrapResponse<T>(response: any): T {
  return unwrap<T>(response);
}

export interface PaginatedResult<T> {
  list: T[];
  total: number;
  page: number;
  limit: number;
}

export function unwrapPagination<T>(response: any): PaginatedResult<T> {
  const payload = response?.data ?? {};
  return {
    list: (payload.data as T[]) ?? [],
    total: payload.total ?? 0,
    page: payload.page ?? 1,
    limit: payload.limit ?? 20,
  };
}

// 兼容旧代码默认导出 api
export default api;
