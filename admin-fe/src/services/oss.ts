import api from './api';

// 获取阿里云 OSS 上传签名
export async function getOssSignature(params?: { dir?: string }) {
  const res = await api.get('/api/v1/oss/signature', { params });
  return res.data.data;
}
