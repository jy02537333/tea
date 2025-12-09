import api, { unwrapResponse } from './api';

export interface DefaultAddressApiResponse<T = any> {
  address?: T;
  raw?: string;
  updated_at?: string;
}

export async function fetchDefaultAddress(): Promise<DefaultAddressApiResponse | null> {
  const res = await api.get('/api/v1/user/default-address');
  return unwrapResponse<DefaultAddressApiResponse | null>(res);
}

export async function updateDefaultAddress(address: any): Promise<string | undefined> {
  const res = await api.put('/api/v1/user/default-address', { address });
  return unwrapResponse<string | undefined>(res);
}
