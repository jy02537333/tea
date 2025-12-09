import Taro from '@tarojs/taro';
import { fetchDefaultAddress as fetchDefaultAddressApi, updateDefaultAddress as updateDefaultAddressApi } from '../services/address';

export const DEFAULT_ADDRESS_STORAGE_KEY = 'tea_default_address';

export interface ParsedAddress {
  contact?: string;
  phone?: string;
  detail?: string;
  full?: string;
}

export interface StoredAddress extends ParsedAddress {
  orderId?: number;
  orderNo?: string;
  updatedAt?: string;
  timestamp?: number;
  source?: 'local' | 'server';
}

export function parseAddressInfo(addressInfo?: unknown): ParsedAddress | null {
  if (!addressInfo) return null;
  let payload: any = addressInfo;
  if (typeof payload === 'string') {
    const trimmed = payload.trim();
    if (!trimmed) return null;
    try {
      payload = JSON.parse(trimmed);
    } catch {
      return { detail: trimmed, full: trimmed };
    }
  }

  if (typeof payload === 'object' && payload !== null) {
    const contact =
      payload.name ||
      payload.contact ||
      payload.receiver ||
      payload.consignee ||
      payload.userName;
    const phone = payload.phone || payload.mobile || payload.tel || payload.telNumber;
    const detailPieces = [payload.province, payload.city, payload.county, payload.area, payload.street, payload.detail, payload.detailInfo, payload.address, payload.full]
      .filter((str) => typeof str === 'string' && str.trim().length > 0)
      .map((str) => (str as string).trim());
    const detail = payload.detail || payload.detailInfo || payload.address || detailPieces.join('');
    const full = payload.full || detail || detailPieces.join('');

    return {
      contact,
      phone,
      detail: detail || full,
      full: full || detail,
    };
  }

  return null;
}

export function formatAddress(address?: ParsedAddress | null): string {
  if (!address) return '';
  const chunks = [address.contact, address.phone, address.detail || address.full].filter((item) => !!item);
  return chunks.join(' · ');
}

async function loadLocalDefaultAddress(): Promise<StoredAddress | null> {
  try {
    const res = await Taro.getStorage({ key: DEFAULT_ADDRESS_STORAGE_KEY });
    return res.data as StoredAddress;
  } catch {
    return null;
  }
}

async function saveLocalDefaultAddress(address: StoredAddress): Promise<void> {
  await Taro.setStorage({ key: DEFAULT_ADDRESS_STORAGE_KEY, data: address });
}

function toStoredAddress(payload: unknown, metadata?: { updatedAt?: string; source?: StoredAddress['source'] }): StoredAddress | null {
  const parsed = parseAddressInfo(payload);
  if (!parsed) return null;
  const base: StoredAddress = {
    ...parsed,
    full: parsed.full || parsed.detail,
    detail: parsed.detail || parsed.full,
    updatedAt: metadata?.updatedAt,
    timestamp: Date.now(),
    source: metadata?.source,
  };
  if (payload && typeof payload === 'object') {
    const obj = payload as Record<string, any>;
    if (obj.orderId) base.orderId = Number(obj.orderId);
    if (obj.orderNo) base.orderNo = String(obj.orderNo);
  }
  return base;
}

async function fetchRemoteDefaultAddress(): Promise<StoredAddress | null> {
  const resp = await fetchDefaultAddressApi();
  if (!resp?.address) return null;
  return toStoredAddress(resp.address, { updatedAt: resp.updated_at || undefined, source: 'server' });
}

export async function loadDefaultAddress(options?: { refreshRemote?: boolean }): Promise<StoredAddress | null> {
  if (options?.refreshRemote !== false) {
    try {
      const remote = await fetchRemoteDefaultAddress();
      if (remote) {
        await saveLocalDefaultAddress(remote);
        return remote;
      }
    } catch (error) {
      console.error('load remote default address failed', error);
    }
  }
  return loadLocalDefaultAddress();
}

export async function saveDefaultAddress(address: StoredAddress, options?: { syncRemote?: boolean }): Promise<void> {
  const payload: StoredAddress = {
    ...address,
    timestamp: Date.now(),
  };
  await saveLocalDefaultAddress(payload);
  if (options?.syncRemote === false) {
    return;
  }
  try {
    await updateDefaultAddressApi({
      contact: payload.contact,
      phone: payload.phone,
      detail: payload.detail,
      full: payload.full,
      orderId: payload.orderId,
      orderNo: payload.orderNo,
      updatedAt: payload.updatedAt || new Date().toISOString(),
    });
  } catch (error) {
    console.error('sync default address to server failed', error);
  }
}
