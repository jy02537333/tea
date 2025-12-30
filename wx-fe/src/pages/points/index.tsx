import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getMeSummary } from '../../services/me';
import type { MeSummary } from '../../services/types';

export default function PointsPage() {
  const [summary, setSummary] = useState<MeSummary | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    void load();
  }, []);

  async function load() {
    setLoading(true);
    try {
      const s = await getMeSummary();
      setSummary(s);
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '加载积分失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  const points = summary?.points?.balance ?? 0;

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>积分中心</Text>
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && (
        <View style={{ marginTop: 12 }}>
          <Text style={{ display: 'block' }}>当前积分：{points}</Text>
          <Text style={{ display: 'block', marginTop: 12, color: '#999' }}>积分明细与规则可在后续版本完善</Text>
        </View>
      )}
    </View>
  );
}
