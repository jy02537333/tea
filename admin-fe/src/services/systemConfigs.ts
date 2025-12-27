import { api, unwrap } from './api';

export interface SystemConfig {
  id: number;
  config_key: string;
  config_value: string;
  config_type?: string;
  description?: string;
  status?: number;
}

export async function listSystemConfigs(params: { keys?: string[]; prefix?: string } = {}) {
  const res = await api.get('/api/v1/admin/system/configs', {
    params: {
      keys: params.keys?.length ? params.keys.join(',') : undefined,
      prefix: params.prefix,
    },
  });
  return unwrap<{ list: SystemConfig[] }>(res);
}

export type UpsertSystemConfigItem = {
  config_key: string;
  config_value: string;
  config_type?: string;
  description?: string;
  status?: number;
};

export async function upsertSystemConfigs(items: UpsertSystemConfigItem[]) {
  const res = await api.put('/api/v1/admin/system/configs', { items });
  return unwrap<{ ok: boolean; updated: number }>(res);
}
