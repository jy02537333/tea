import api from './api';

// 获取 OSS 文件列表（支持前缀、分页）
export async function listOssFiles(params?: { prefix?: string; marker?: string; limit?: number }) {
  const res = await api.get('/api/v1/oss/list', { params });
  return res.data;
}

// 批量删除 OSS 文件
export async function deleteOssFiles(urls: string[]) {
  return api.post('/api/v1/oss/delete', { urls });
}
