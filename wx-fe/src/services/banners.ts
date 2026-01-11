import api, { unwrapResponse } from './api';

export interface Banner {
  id: number;
  title?: string;
  image_url: string;
  link_type?: number;
  link_url?: string;
  sort?: number;
  status?: number;
}

export async function listBanners(limit = 10): Promise<Banner[]> {
  const res = await api.get('/api/v1/banners', { params: { limit } });
  return unwrapResponse<Banner[]>(res);
}
