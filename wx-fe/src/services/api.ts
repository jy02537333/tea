import axios from 'axios';

// In mini-program runtime `wx` global exists; declare to satisfy TypeScript during local checks
declare const wx: any;

const BASE_URL = (typeof (import.meta as any) !== 'undefined' && (import.meta as any).env?.VITE_API_BASE_URL) || process.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
});

export function setToken(token: string | null) {
  if (typeof wx !== 'undefined' && (wx as any).setStorageSync) {
    try { (wx as any).setStorageSync('token', token); } catch (e) {}
  } else if (typeof localStorage !== 'undefined') {
    if (token) localStorage.setItem('token', token); else localStorage.removeItem('token');
  }
  if (token) api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
}

// load token
try {
  const t = (typeof wx !== 'undefined' && (wx as any).getStorageSync) ? (wx as any).getStorageSync('token') : (typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null);
  if (t) api.defaults.headers.common['Authorization'] = `Bearer ${t}`;
} catch (e) {}

export default api;
// Helper similar to admin to unwrap ApiResponse<T>
export function unwrapResponse<T>(res: any): T {
  if (res && res.data && typeof res.data === 'object' && 'data' in res.data) {
    return res.data.data as T;
  }
  if (res && typeof res.data !== 'undefined') {
    return res.data as T;
  }
  return res as T;
}
