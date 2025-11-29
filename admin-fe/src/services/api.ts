import axios from 'axios';

const BASE_URL = (typeof (import.meta as any) !== 'undefined' && (import.meta as any).env?.VITE_API_BASE_URL) || process.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
});

// Token helpers
export function setToken(token: string | null) {
  if (token) {
    localStorage.setItem('token', token);
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    localStorage.removeItem('token');
    delete api.defaults.headers.common['Authorization'];
  }
}

// load token on startup
const saved = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
if (saved) {
  api.defaults.headers.common['Authorization'] = `Bearer ${saved}`;
}

// Request interceptor example
api.interceptors.request.use((config) => {
  // allow explicit override
  return config;
});

// Auth interceptor: attach Bearer token when available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers = config.headers || {};
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: simple 401 handling
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401) {
      // frontend should handle redirect to login
      setToken(null);
      // optionally dispatch an event
      window.dispatchEvent(new Event('unauthorized'));
    }
    return Promise.reject(err);
  }
);

// Response interceptor: redirect to login on 401
api.interceptors.response.use(
  (resp) => resp,
  (err) => {
    if (err?.response?.status === 401) {
      // preserve current path
      const from = encodeURIComponent(window.location.pathname + window.location.search);
      if (!window.location.pathname.includes('/login')) {
        window.location.href = `/login?from=${from}`;
      }
    }
    return Promise.reject(err);
  }
);

export default api;

// Helper to unwrap responses consistently and satisfy strict typing in services.
export function unwrapResponse<T>(res: any): T {
  // prefer { data: ... } wrapper
  if (res && res.data && typeof res.data === 'object' && 'data' in res.data) {
    return res.data.data as T;
  }
  // fallback when server responds without wrapper
  if (res && typeof res.data !== 'undefined') {
    return res.data as T;
  }
  return res as T;
}
