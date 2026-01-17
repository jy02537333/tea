import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getMeSummary } from '../../services/me';
import type { MeSummary } from '../../services/types';

export default function WalletPage() {
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
      Taro.showToast({ title: e?.message || '加载钱包失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  const balanceCents = summary?.wallet?.balance_cents ?? 0;
  const teaCoins = summary?.wallet?.tea_coins ?? 0;
  const frozenCents = summary?.wallet?.frozen_amount_cents ?? 0;

  function centsToYuan(c: number | undefined) {
    if (!c || c <= 0) return '0.00';
    return (c / 100).toFixed(2);
  }

  return (
    <View data-testid="page-wallet" style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>钱包</Text>
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && (
        <View style={{ marginTop: 12 }}>
          <Text style={{ display: 'block' }}>余额：¥{centsToYuan(balanceCents)}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>冻结金额：¥{centsToYuan(frozenCents)}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>茶币：{teaCoins}</Text>
          <Text style={{ display: 'block', marginTop: 12, color: '#999' }}>提现入口与明细可在后续版本接入</Text>
        </View>
      )}
    </View>
  );
}
