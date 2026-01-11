import Taro from '@tarojs/taro';

export type ShareAttributionParams = {
  sharer_uid?: number;
  share_store_id?: number;
};

function readPositiveIntFromStorage(key: string): number | undefined {
  try {
    const raw = Taro.getStorageSync(key);
    if (raw === null || typeof raw === 'undefined') return undefined;
    const n = Number(raw);
    if (!Number.isFinite(n) || n <= 0) return undefined;
    return Math.floor(n);
  } catch (_) {
    return undefined;
  }
}

/**
 * Frontend adapter for backend share attribution:
 * - sharer_uid comes from stored referrer_id
 * - share_store_id must match order store_id when store-bound; platform orders use 0
 */
export function buildOrderShareAttributionParams(options?: {
  storeId?: number;
  requireStoreId?: boolean;
}): ShareAttributionParams {
  const sharerUid = readPositiveIntFromStorage('referrer_id');
  if (!sharerUid) return {};

  const storeId = options?.storeId;
  const hasStoreId = !!storeId && Number.isFinite(storeId) && storeId > 0;

  if (options?.requireStoreId) {
    if (!hasStoreId) return {};
    return { sharer_uid: sharerUid, share_store_id: storeId };
  }

  // Not store-bound: allow sharer_uid with share_store_id=0 (platform order)
  return { sharer_uid: sharerUid, share_store_id: hasStoreId ? storeId : 0 };
}
