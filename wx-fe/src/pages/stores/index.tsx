import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listStores } from '../../services/stores';
import { Store } from '../../services/types';
import usePermission from '../../hooks/usePermission';
import { PERM_HINT_STORE_MGMT_READONLY_PAGE, PERM_HINT_STORE_MGMT_READONLY_TOAST, PERM_TOAST_NO_STORE_FINANCE } from '../../constants/permission';

export default function StoresPage() {
  const perm = usePermission();
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
    const canUseFinance = perm.allowedStoreFinance;
    if (perm.allowedStoreMgmt) {
      Taro.showToast({ title: canUseFinance ? '已设为当前门店，可在财务页使用' : '已设为当前门店（财务页需权限）', icon: canUseFinance ? 'success' : 'none' });
    } else {
      Taro.showToast({ title: '已设为当前门店（仅用于前端上下文），财务页需权限', icon: 'none' });
    }
  }

  return (
    <View style={{ padding: 12 }}>
      <Text style={{ fontSize: 16, fontWeight: 'bold' }}>门店列表</Text>
      {!perm.allowedStoreMgmt && (
        <Text style={{ display: 'block', marginTop: 6, color: '#999' }}>{PERM_HINT_STORE_MGMT_READONLY_PAGE}</Text>
      )}
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && !stores.length && <Text style={{ display: 'block', marginTop: 8 }}>暂无门店</Text>}
      {stores.map((s) => (
        <View key={s.id} style={{ marginTop: 12, paddingBottom: 8, borderBottom: '1px solid #eee' }}>
          <Text style={{ display: 'block', fontSize: 16 }}>{s.name}</Text>
          {s.address && <Text style={{ display: 'block', color: '#666' }}>地址：{s.address}</Text>}
          <View style={{ marginTop: 8, display: 'flex', gap: 8 }}>
            <Button size="mini" type="primary" onClick={() => goDetail(s.id)}>查看详情</Button>
            <Button size="mini" onClick={() => setCurrent(s.id)}>设为当前门店</Button>
            <Button
              size="mini"
              onClick={() => {
                if (!perm.allowedStoreFinance) {
                  Taro.showToast({ title: PERM_TOAST_NO_STORE_FINANCE, icon: 'none' });
                  return;
                }
                Taro.navigateTo({ url: `/pages/store-finance/index?store_id=${s.id}` });
              }}
            >
              财务流水
            </Button>
            {!perm.allowedStoreFinance && (
              <Text style={{ color: '#999', fontSize: 12 }}>（需权限）</Text>
            )}
          </View>
        </View>
      ))}
    </View>
  );
}
