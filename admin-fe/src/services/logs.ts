import { api, PaginatedResult, unwrapPagination } from './api';

export interface OperationLog {
  id: number;
  user_id: number;
  module?: string;
  operation?: string;
  ip?: string;
  user_agent?: string;
  created_at?: string;
}

export interface AccessLog {
  id: number;
  user_id?: number;
  method?: string;
  path?: string;
  query?: string;
  status_code?: number;
  latency?: number;
  ip?: string;
  user_agent?: string;
  created_at?: string;
}

export interface OperationLogQuery {
  page?: number;
  limit?: number;
  module?: string;
  method?: string;
  operation?: string;
  order_id?: string;
  path?: string;
  user_id?: string;
  start?: string;
  end?: string;
}

export interface AccessLogQuery {
  page?: number;
  limit?: number;
  method?: string;
  path?: string;
  user_id?: string;
  status?: string;
  start?: string;
  end?: string;
}

export async function listOperationLogs(params: OperationLogQuery): Promise<PaginatedResult<OperationLog>> {
  const res = await api.get('/api/v1/admin/logs/operations', { params });
  return unwrapPagination<OperationLog>(res);
}

export async function exportOperationLogs(params: Omit<OperationLogQuery, 'page' | 'limit'> & { format?: 'csv' | 'xlsx' } = {}) {
  const res = await api.get('/api/v1/admin/logs/operations/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}

export async function listAccessLogs(params: AccessLogQuery): Promise<PaginatedResult<AccessLog>> {
  const res = await api.get('/api/v1/admin/logs/access', { params });
  return unwrapPagination<AccessLog>(res);
}

export async function exportAccessLogs(params: Omit<AccessLogQuery, 'page' | 'limit'> & { format?: 'csv' | 'xlsx' } = {}) {
  const res = await api.get('/api/v1/admin/logs/access/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}
