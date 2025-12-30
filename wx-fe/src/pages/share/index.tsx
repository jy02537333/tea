import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getMeSummary } from '../../services/me';
import type { MeSummary } from '../../services/types';

export default function SharePage() {
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
      Taro.showToast({ title: e?.message || '加载分享数据失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  const shareStats = summary?.share;

  function handleGenerateShare() {
    Taro.showToast({ title: '分享链接已生成', icon: 'none' });
  }

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>分享推广</Text>
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && (
        <View style={{ marginTop: 12 }}>
          <Text style={{ display: 'block' }}>直推人数：{shareStats?.direct_count ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>团队人数：{shareStats?.team_count ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>累计佣金（分）：{shareStats?.total_commission_cents ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>可提现佣金（分）：{shareStats?.available_commission_cents ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>冻结佣金（分）：{shareStats?.frozen_commission_cents ?? 0}</Text>

          <Button style={{ marginTop: 12 }} type="primary" size="mini" onClick={handleGenerateShare}>
            生成分享链接/海报
          </Button>
          <Text style={{ display: 'block', marginTop: 12, color: '#999' }}>分享物料与二维码将在后续版本完善</Text>
        </View>
      )}
    </View>
  );
}
