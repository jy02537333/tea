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

export interface OssPolicy {
  accessid: string;
  host: string;
  policy: string;
  signature: string;
  dir: string;
  expire: number;
}

export async function getOssPolicy(dir?: string): Promise<OssPolicy> {
  const res = await api.post('/api/v1/upload/oss/policy', dir ? { dir } : {});
  return unwrap<OssPolicy>(res);
}
