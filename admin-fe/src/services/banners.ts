import { api, unwrap } from './api';

export interface Banner {
  id: number;
  title?: string;
  image_url: string;
  link_type?: number;
  link_url?: string;
  sort?: number;
  status?: number;
  created_at?: string;
}

export interface BannerQuery {
  keyword?: string;
  status?: number;
}

export interface BannerPayload {
  title?: string;
  image_url: string;
  link_type?: number;
  link_url?: string;
  sort?: number;
  status?: number;
}

export async function listBanners(params: BannerQuery = {}) {
  const res = await api.get('/api/v1/admin/banners', { params });
  return unwrap<Banner[]>(res);
}

export async function createBanner(payload: BannerPayload) {
  const res = await api.post('/api/v1/admin/banners', payload);
  return unwrap<Banner>(res);
}

export async function updateBanner(id: number, payload: BannerPayload) {
  const res = await api.put(`/api/v1/admin/banners/${id}`, payload);
  return unwrap<Banner>(res);
}

export async function deleteBanner(id: number) {
  const res = await api.delete(`/api/v1/admin/banners/${id}`);
  return unwrap<{ ok: boolean }>(res);
}
