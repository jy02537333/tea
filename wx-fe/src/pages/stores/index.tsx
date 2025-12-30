import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listStores } from '../../services/stores';
import { Store } from '../../services/types';

export default function StoresPage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    void fetchStores();
  }, []);

  async function fetchStores() {
    setLoading(true);
    try {
      const res = await listStores({ page: 1, limit: 50 });
      const maybe: any = res;
      const items: Store[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      setStores(items);
    } catch (e) {
      console.error('load stores failed', e);
      Taro.showToast({ title: '加载门店失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  function goDetail(id: number) {
    Taro.navigateTo({ url: `/pages/store-detail/index?store_id=${id}` });
  }

  function setCurrent(id: number) {
    try { Taro.setStorageSync('current_store_id', String(id)); } catch (_) {}
    Taro.showToast({ title: '已设为当前门店', icon: 'success' });
  }

  return (
    <View style={{ padding: 12 }}>
      <Text style={{ fontSize: 16, fontWeight: 'bold' }}>门店列表</Text>
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && !stores.length && <Text style={{ display: 'block', marginTop: 8 }}>暂无门店</Text>}
      {stores.map((s) => (
        <View key={s.id} style={{ marginTop: 12, paddingBottom: 8, borderBottom: '1px solid #eee' }}>
          <Text style={{ display: 'block', fontSize: 16 }}>{s.name}</Text>
          {s.address && <Text style={{ display: 'block', color: '#666' }}>地址：{s.address}</Text>}
          <View style={{ marginTop: 8, display: 'flex', gap: 8 }}>
            <Button size="mini" type="primary" onClick={() => goDetail(s.id)}>查看详情</Button>
            <Button size="mini" onClick={() => setCurrent(s.id)}>设为当前门店</Button>
          </View>
        </View>
      ))}
    </View>
  );
}
