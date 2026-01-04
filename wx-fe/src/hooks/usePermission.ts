import { useEffect, useState, useCallback } from 'react';
import { getMeSummary } from '../services/me';
import type { User } from '../services/types';

export interface PermissionState {
  user: User | null;
  role: string;
  loading: boolean;
  isAdmin: boolean;
  allowedStoreMgmt: boolean;
  allowedStoreAccounts: boolean;
  allowedStoreFinance: boolean;
  canManageStores: (u?: User | null) => boolean;
  canManageStoreAccounts: (u?: User | null) => boolean;
  canManageStoreFinance: (u?: User | null) => boolean;
  refresh: () => Promise<void>;
}

export default function usePermission(): PermissionState {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const s = await getMeSummary();
      setUser(s?.user || null);
    } catch (e) {
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const role = (user?.role || '').toLowerCase();
  const isAdmin = !!role && role.includes('admin');
  const allowedStoreAccounts = !!role && (role.includes('admin') || role.includes('store:accounts') || role.includes('store'));
  const allowedStoreFinance = !!role && (role.includes('admin') || role.includes('store:finance') || role.includes('store'));
  const allowedStoreMgmt = allowedStoreAccounts || allowedStoreFinance;

  function canManageStores(u?: User | null) {
    const r = (u?.role || role || '').toLowerCase();
    return !!r && (r.includes('admin') || r.includes('store'));
  }

  function canManageStoreAccounts(u?: User | null) {
    const r = (u?.role || role || '').toLowerCase();
    return !!r && (r.includes('admin') || r.includes('store:accounts') || r.includes('store'));
  }

  function canManageStoreFinance(u?: User | null) {
    const r = (u?.role || role || '').toLowerCase();
    return !!r && (r.includes('admin') || r.includes('store:finance') || r.includes('store'));
  }

  return {
    user,
    role,
    loading,
    isAdmin,
    allowedStoreMgmt,
    allowedStoreAccounts,
    allowedStoreFinance,
    canManageStores,
    canManageStoreAccounts,
    canManageStoreFinance,
    refresh: load,
  };
}
