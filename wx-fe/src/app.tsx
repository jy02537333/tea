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
      if (typeof query.store_id === 'string' || typeof query.store_id === 'number') {
        const n = Number(query.store_id);
        if (!Number.isNaN(n) && n > 0) sid = n;
      }
      // 尝试从 scene 中解析 store_id=xxx
      if (!sid && typeof query.scene === 'string' && query.scene) {
        const decoded = decodeURIComponent(query.scene);
        const match = decoded.match(/(?:^|&|\?)store_id=(\d+)/);
        if (match && match[1]) {
          const n = Number(match[1]);
          if (!Number.isNaN(n) && n > 0) sid = n;
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
      }
      if (sid) {
        Taro.setStorageSync('current_store_id', String(sid));
      }
    } catch (_) {
      // ignore
    }
  }, []);

  return children;
}

export default App;
