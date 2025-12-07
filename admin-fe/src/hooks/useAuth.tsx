import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { devLogin as apiDevLogin, getUserInfo, login as apiLogin, PasswordLoginPayload } from '../services/auth';
import { User } from '../types/user';
import { setToken } from '../services/api';
import { getUserPermissions } from '../services/users';

interface AuthContextValue {
  token: string | null;
  user: User | null;
  permissions: string[];
  loading: boolean;
  login: (payload: PasswordLoginPayload) => Promise<void>;
  devLogin: (openid: string) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
  refreshPermissions: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setTokenState] = useState<string | null>(typeof window !== 'undefined' ? localStorage.getItem('token') : null);
  const [user, setUser] = useState<User | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPermissions = useCallback(async (userId: number) => {
    try {
      const list = await getUserPermissions(userId);
      setPermissions(list);
    } catch (error) {
      console.warn('无法获取权限列表', error);
      setPermissions([]);
    }
  }, []);

  useEffect(() => {
    async function bootstrap() {
      if (!token) {
        setLoading(false);
        setPermissions([]);
        return;
      }
      try {
        const me = await getUserInfo();
        setUser(me);
        if (me?.id) {
          await fetchPermissions(me.id);
        } else {
          setPermissions([]);
        }
      } catch (error) {
        setToken(null);
        setTokenState(null);
        setPermissions([]);
      } finally {
        setLoading(false);
      }
    }
    bootstrap();
  }, [token, fetchPermissions]);

  const logout = () => {
    setUser(null);
    setToken(null);
    setTokenState(null);
    setPermissions([]);
  };

  const login = useCallback(async (payload: PasswordLoginPayload) => {
    const res = await apiLogin(payload);
    if (res.token) {
      setToken(res.token);
      setTokenState(res.token);
      const me = await getUserInfo();
      setUser(me);
      if (me?.id) {
        await fetchPermissions(me.id);
      }
    }
  }, [fetchPermissions]);

  const devLogin = useCallback(async (openid: string) => {
    const res = await apiDevLogin(openid);
    if (res.token) {
      setToken(res.token);
      setTokenState(res.token);
      const me = await getUserInfo();
      setUser(me);
      if (me?.id) {
        await fetchPermissions(me.id);
      }
    }
  }, [fetchPermissions]);

  const hasPermission = useCallback(
    (permission: string) => {
      if (!permission) return true;
      if (user?.role === 'admin') return true;
      return permissions.includes(permission);
    },
    [permissions, user?.role]
  );

  const refreshPermissions = useCallback(async () => {
    if (!user?.id) {
      setPermissions([]);
      return;
    }
    await fetchPermissions(user.id);
  }, [fetchPermissions, user?.id]);

  const value = useMemo(
    () => ({ token, user, permissions, loading, login, devLogin, logout, hasPermission, refreshPermissions }),
    [token, user, permissions, loading, login, devLogin, logout, hasPermission, refreshPermissions]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuthContext() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuthContext 必须在 AuthProvider 内使用');
  }
  return ctx;
}
