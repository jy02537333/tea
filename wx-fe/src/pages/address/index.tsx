import React, { useEffect, useState } from 'react';
import { Button, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import type { Order } from '../../services/types';
import { formatAddress, loadDefaultAddress, parseAddressInfo, saveDefaultAddress, type StoredAddress } from '../../utils/address';

export default function AddressManagerPage() {
  const [addresses, setAddresses] = useState<StoredAddress[]>([]);
  const [defaultAddress, setDefaultAddress] = useState<StoredAddress | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void bootstrap();
  }, []);

  usePullDownRefresh(() => {
    void Promise.all([loadAddresses(true), refreshDefaultAddress()]);
  });

  async function bootstrap() {
    await Promise.all([loadAddresses(), refreshDefaultAddress()]);
  }

  async function refreshDefaultAddress() {
    const stored = await loadDefaultAddress({ refreshRemote: true });
    setDefaultAddress(stored);
  }

  async function loadAddresses(force = false) {
    if (loading && !force) return;
    setLoading(true);
    try {
      const resp = await listOrders({ page: 1, limit: 50 });
      const list = (resp?.data || []) as Order[];
      const parsed = list
        .map((order) => {
          const parsedInfo = parseAddressInfo(order.address_info);
          if (!parsedInfo) return null;
          return {
            ...parsedInfo,
            full: parsedInfo.full || formatAddress(parsedInfo),
            orderId: order.id,
            orderNo: order.order_no,
            updatedAt: (order as any)?.updated_at || (order as any)?.created_at,
          } as StoredAddress;
        })
        .filter(Boolean) as StoredAddress[];

      const dedupMap = new Map<string, StoredAddress>();
      parsed.forEach((addr) => {
        const key = `${addr.full || addr.detail || ''}-${addr.phone || ''}`;
        if (!dedupMap.has(key)) {
          dedupMap.set(key, addr);
        }
      });

      const deduped = Array.from(dedupMap.values());
      setAddresses(deduped);
      setError(deduped.length ? null : '暂无可回收的地址，完成一次配送订单即可沉淀地址信息');
    } catch (err: any) {
      console.error('load addresses failed', err);
      setError(err?.message || '加载失败');
    } finally {
      setLoading(false);
      Taro.stopPullDownRefresh();
    }
  }

  async function handleSetDefault(address: StoredAddress) {
    try {
      await saveDefaultAddress(address);
      await refreshDefaultAddress();
      Taro.showToast({ title: '默认地址已更新', icon: 'success' });
    } catch (err) {
      console.error('save default address failed', err);
      Taro.showToast({ title: '保存失败，请稍后再试', icon: 'none' });
    }
  }

  async function handleCopy(address: StoredAddress) {
    try {
      const text = address.full || formatAddress(address);
      if (!text) {
        Taro.showToast({ title: '地址为空', icon: 'none' });
        return;
      }
      await Taro.setClipboardData({ data: text });
      Taro.showToast({ title: '已复制', icon: 'success' });
    } catch (err) {
      console.error('copy address failed', err);
    }
  }

  return (
    <View style={{ padding: 16, backgroundColor: '#f5f6f8', minHeight: '100vh' }}>
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>默认收货地址</Text>
        {defaultAddress ? (
          <View style={{ marginTop: 12 }}>
            <Text style={{ display: 'block', fontSize: 16 }}>{formatAddress(defaultAddress)}</Text>
            <Text style={{ display: 'block', color: '#999', marginTop: 4 }}>最近同步：{defaultAddress.updatedAt || '未知'}</Text>
          </View>
        ) : (
          <Text style={{ display: 'block', color: '#999', marginTop: 12 }}>尚未设置默认地址，可从下方历史地址选择</Text>
        )}
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <View style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text style={{ fontSize: 18, fontWeight: 'bold' }}>历史地址</Text>
          <Button size="mini" loading={loading} onClick={() => loadAddresses(true)}>
            刷新
          </Button>
        </View>
        {error && <Text style={{ color: '#ff4d4f', display: 'block', marginTop: 12 }}>{error}</Text>}
        {!error && !addresses.length && <Text style={{ color: '#999', display: 'block', marginTop: 12 }}>暂无数据</Text>}
        {addresses.map((addr) => (
          <View key={`${addr.orderId}-${addr.detail}`} style={{ marginTop: 12, borderTop: '1px solid #f0f0f0', paddingTop: 12 }}>
            <Text style={{ display: 'block', fontSize: 16 }}>{formatAddress(addr)}</Text>
            <Text style={{ color: '#999', fontSize: 12 }}>来源订单 #{addr.orderNo}</Text>
            <View style={{ display: 'flex', marginTop: 8, gap: 8 }}>
              <Button size="mini" onClick={() => handleSetDefault(addr)}>
                设为默认
              </Button>
              <Button size="mini" onClick={() => handleCopy(addr)}>
                复制
              </Button>
            </View>
          </View>
        ))}
      </View>

      <View style={{ marginTop: 24 }}>
          <Button type="primary" onClick={() => Taro.navigateTo({ url: '/pages/checkout/index' })}>
          去下单填写新地址
        </Button>
      </View>
    </View>
  );
}
