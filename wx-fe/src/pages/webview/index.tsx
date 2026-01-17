import React, { useMemo } from 'react';
import { View, WebView } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';

export default function WebviewPage() {
  const router = useRouter();
  const url = useMemo(() => {
    const raw = (router as any)?.params?.url;
    if (!raw) return '';
    try {
      return decodeURIComponent(String(raw));
    } catch (_e) {
      return String(raw);
    }
  }, [router]);

  // 小程序端用 WebView 承载外链；H5 端直接跳转
  if (process.env.TARO_ENV !== 'weapp') {
    if (url) {
      // eslint-disable-next-line no-restricted-globals
      (window as any).location.href = url;
    }
    return <View />;
  }

  return (
    <View style={{ minHeight: '100vh' }}>
      <WebView src={url} />
    </View>
  );
}
