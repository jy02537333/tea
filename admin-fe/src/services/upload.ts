import { api, unwrap } from './api';

export interface UploadResponse {
  url: string;
}

export async function uploadMedia(file: File): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append('file', file);
  const res = await api.post('/api/v1/admin/uploads', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return unwrap<UploadResponse>(res);
}
