import api, { unwrapResponse } from './api';

export interface SystemConfig {
  id: number;
  config_key: string;
  config_value: string;
  config_type?: string;
  description?: string;
  status?: number;
}

export async function listSiteConfigs(params: { keys?: string[]; prefix?: string } = {}): Promise<SystemConfig[]> {
  const query: Record<string, string> = {};
  if (params.keys && params.keys.length > 0) query.keys = params.keys.join(',');
  if (params.prefix) query.prefix = params.prefix;

  const res = await api.get('/api/v1/site/configs', { params: query });
  const data = unwrapResponse<{ list: SystemConfig[] }>(res);
  return Array.isArray(data?.list) ? data.list : [];
}
