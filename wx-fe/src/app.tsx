import type { PropsWithChildren } from 'react';
import { useEffect } from 'react';
import Taro from '@tarojs/taro';
import './app.scss';

function App({ children }: PropsWithChildren<Record<string, unknown>>) {
  useEffect(() => {
    try {
      const opts = (Taro as any)?.getLaunchOptionsSync ? (Taro as any).getLaunchOptionsSync() : undefined;
      const query = opts?.query || {};
      let sid: number | undefined = undefined;
      let referrerId: number | undefined = undefined;
      if (typeof query.store_id === 'string' || typeof query.store_id === 'number') {
        const n = Number(query.store_id);
        if (!Number.isNaN(n) && n > 0) sid = n;
      }
      // 尝试从 scene 中解析 store_id=xxx
      if (!sid && typeof query.scene === 'string' && query.scene) {
        const decoded = decodeURIComponent(query.scene);
        const m1 = decoded.match(/(?:^|&|\?)store_id=(\d+)/);
        if (m1 && m1[1]) {
          const n = Number(m1[1]);
          if (!Number.isNaN(n) && n > 0) sid = n;
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
        const r = sp.get('referrer_id');
        if (r) {
          const n = Number(r);
          if (!Number.isNaN(n) && n > 0) referrerId = n;
        }
      }
      if (sid) {
        Taro.setStorageSync('current_store_id', String(sid));
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

  return children;
}

export default App;
