import type { PropsWithChildren } from 'react';
import { useEffect } from 'react';
import Taro from '@tarojs/taro';
import { getToken, setToken } from './services/api';
import './app.scss';

const PUBLIC_PAGES = new Set(['pages/index/index', 'pages/login/index']);

function normalizeRoutePath(path: string | undefined | null): string {
  const raw = (path || '').trim();
  if (!raw) return '';
  const noQuery = raw.split('?')[0] || '';
  const noHash = noQuery.startsWith('#') ? noQuery.slice(1) : noQuery;
  return noHash.replace(/^\//, '');
}

function parseH5HashToPathAndQuery(hash: string): { path: string; query: string } {
  const h = (hash || '').replace(/^#/, '');
  // Typical Taro H5: /pages/xxx/index?x=1
  const [p, q] = h.split('?');
  return { path: normalizeRoutePath(p), query: q ? `?${q}` : '' };
}

function getCurrentRouteInfo(): { path: string; fullPath: string } {
  try {
    const pages = (Taro as any)?.getCurrentPages?.() as any[] | undefined;
    const last = pages && pages.length ? pages[pages.length - 1] : undefined;
    const route = normalizeRoutePath(last?.route);
    const options = last?.options || last?.$taroParams || {};
    if (route) {
      const qs = new URLSearchParams();
      Object.entries(options).forEach(([k, v]) => {
        if (typeof v === 'undefined' || v === null) return;
        qs.set(k, String(v));
      });
      const query = qs.toString();
      const fullPath = `/${route}${query ? `?${query}` : ''}`;
      return { path: route, fullPath };
    }
  } catch (_) {
    // ignore
  }

  if (typeof window !== 'undefined') {
    const { path, query } = parseH5HashToPathAndQuery(window.location.hash);
    if (path) return { path, fullPath: `/${path}${query}` };
  }

  return { path: '', fullPath: '' };
}

function safeInternalRedirectTarget(raw: string | undefined | null): string {
  const v = (raw || '').trim();
  if (!v) return '';

  const decoded = (() => {
    try {
      return decodeURIComponent(v);
    } catch (_) {
      return v;
    }
  })();

  const cleaned = decoded.startsWith('/') ? decoded : `/${decoded}`;
  if (!cleaned.startsWith('/pages/')) return '';
  return cleaned;
}

function App({ children }: PropsWithChildren<Record<string, unknown>>) {
  useEffect(() => {
    try {
      const opts = (Taro as any)?.getLaunchOptionsSync ? (Taro as any).getLaunchOptionsSync() : undefined;
      const query = opts?.query || {};
      let sid: number | undefined = undefined;
      let referrerId: number | undefined = undefined;
      let tableId: number | undefined = undefined;
      let tableNo: string | undefined = undefined;
      if (typeof query.store_id === 'string' || typeof query.store_id === 'number') {
        const n = Number(query.store_id);
        if (!Number.isNaN(n) && n > 0) sid = n;
      }
      if (typeof query.table_id === 'string' || typeof query.table_id === 'number') {
        const n = Number(query.table_id);
        if (!Number.isNaN(n) && n > 0) tableId = n;
      }
      if (typeof query.table_no === 'string' || typeof query.table_no === 'number') {
        const v = String(query.table_no).trim();
        if (v) tableNo = v;
      }
      // 尝试从 scene 中解析 store_id=xxx
      if (!sid && typeof query.scene === 'string' && query.scene) {
        const decoded = decodeURIComponent(query.scene);
        const m1 = decoded.match(/(?:^|&|\?)store_id=(\d+)/);
        if (m1 && m1[1]) {
          const n = Number(m1[1]);
          if (!Number.isNaN(n) && n > 0) sid = n;
        }
        const mtid = decoded.match(/(?:^|&|\?)table_id=(\d+)/);
        if (!tableId && mtid && mtid[1]) {
          const n = Number(mtid[1]);
          if (!Number.isNaN(n) && n > 0) tableId = n;
        }
        const mtno = decoded.match(/(?:^|&|\?)table_no=([^&]+)/);
        if (!tableNo && mtno && mtno[1]) {
          const v = String(mtno[1]).trim();
          if (v) tableNo = v;
        }
        const m2 = decoded.match(/(?:^|&|\?)referrer_id=(\d+)/);
        if (m2 && m2[1]) {
          const n = Number(m2[1]);
          if (!Number.isNaN(n) && n > 0) referrerId = n;
        }
      }
      // H5 场景：从当前 URL 查询参数兜底解析
      if (!sid && typeof window !== 'undefined') {
        const sp = new URLSearchParams(window.location.search);
        const v = sp.get('store_id');
        if (v) {
          const n = Number(v);
          if (!Number.isNaN(n) && n > 0) sid = n;
        }
        if (!tableId) {
          const tid = sp.get('table_id');
          if (tid) {
            const n = Number(tid);
            if (!Number.isNaN(n) && n > 0) tableId = n;
          }
        }
        if (!tableNo) {
          const tno = sp.get('table_no');
          if (tno) {
            const vv = String(tno).trim();
            if (vv) tableNo = vv;
          }
        }
        const r = sp.get('referrer_id');
        if (r) {
          const n = Number(r);
          if (!Number.isNaN(n) && n > 0) referrerId = n;
        }
        const tk = sp.get('tk');
        if (tk) {
          try {
            // Keep storage + axios header + localStorage consistent
            setToken(tk);
          } catch (_) {
            try {
              Taro.setStorageSync('token', tk);
            } catch (_e) {}
          }
        }
      }
      if (sid) {
        Taro.setStorageSync('current_store_id', String(sid));
      }
      if (tableId) {
        Taro.setStorageSync('current_table_id', String(tableId));
      } else {
        try {
          Taro.removeStorageSync('current_table_id');
        } catch (_) {}
      }
      if (tableNo) {
        Taro.setStorageSync('current_table_no', String(tableNo));
      } else {
        try {
          Taro.removeStorageSync('current_table_no');
        } catch (_) {}
      }
      if (referrerId) {
        Taro.setStorageSync('referrer_id', String(referrerId));
        // 非阻塞：尝试上报绑定关系
        try {
          // 动态导入以避免循环依赖
          const mod = require('./services/share');
          if (mod && typeof mod.bindReferral === 'function') {
            mod.bindReferral(referrerId);
          }
        } catch (_) {}
      }
    } catch (_) {
      // ignore
    }
  }, []);

  useEffect(() => {
    let redirecting = false;
    let lastRedirectAt = 0;

    const ensureAuthed = () => {
      const { path, fullPath } = getCurrentRouteInfo();
      if (!path) return;
      if (PUBLIC_PAGES.has(path)) return;

      const token = getToken();
      if (token) return;

      const now = Date.now();
      if (redirecting || now - lastRedirectAt < 800) return;
      redirecting = true;
      lastRedirectAt = now;

      const target = safeInternalRedirectTarget(fullPath) || `/${path}`;
      const url = `/pages/login/index?redirect=${encodeURIComponent(target)}`;

      try {
        Taro.redirectTo({ url });
      } catch (_) {
        // ignore
      } finally {
        setTimeout(() => {
          redirecting = false;
        }, 800);
      }
    };

    const scheduleEnsure = () => setTimeout(ensureAuthed, 0);

    // Initial check
    scheduleEnsure();

    // Mini-program style hook (if available)
    const onAppRoute = (Taro as any)?.onAppRoute;
    if (typeof onAppRoute === 'function') {
      try {
        onAppRoute(scheduleEnsure);
      } catch (_) {
        // ignore
      }
    }

    // Taro internal router change event (works across some runtimes)
    try {
      (Taro as any)?.eventCenter?.on?.('__taroRouterChange', scheduleEnsure);
    } catch (_) {
      // ignore
    }

    // H5 fallback
    if (typeof window !== 'undefined') {
      window.addEventListener('hashchange', scheduleEnsure);
      window.addEventListener('popstate', scheduleEnsure);
    }

    return () => {
      try {
        (Taro as any)?.eventCenter?.off?.('__taroRouterChange', scheduleEnsure);
      } catch (_) {
        // ignore
      }
      if (typeof window !== 'undefined') {
        window.removeEventListener('hashchange', scheduleEnsure);
        window.removeEventListener('popstate', scheduleEnsure);
      }
    };
  }, []);

  return children;
}

export default App;
