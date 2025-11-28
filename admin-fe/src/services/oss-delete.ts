import api from './api';

// 删除阿里云 OSS 图片（支持批量）
export async function deleteOssFiles(urls: string[]) {
  return api.post('/api/v1/oss/delete', { urls });
}
