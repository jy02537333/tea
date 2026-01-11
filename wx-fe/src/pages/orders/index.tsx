import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import { Order, Store } from '../../services/types';
import { getStore } from '../../services/stores';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<number | undefined>(undefined);
  const [currentStore, setCurrentStore] = useState<Store | null>(null);

  useEffect(() => {
    void loadCurrentStore();
  }, []);

  useEffect(() => {
    void fetchOrders();
  }, [status, currentStore?.id]);

  async function loadCurrentStore() {
    try {
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const storeId = storeIdRaw ? Number(storeIdRaw) : NaN;
      if (!Number.isNaN(storeId) && storeId > 0) {
        const s = await getStore(storeId);
        setCurrentStore(s as Store);
      }
    } catch (_) {
      // ignore
    }
  }

  async function fetchOrders() {
    setLoading(true);
    try {
      const sid = currentStore?.id;
      const params: { page?: number; limit?: number; status?: number; store_id?: number } = { page: 1, limit: 20 };
      if (typeof status === 'number') params.status = status;
      if (sid && Number.isFinite(sid) && sid > 0) params.store_id = sid;
      const res = await listOrders(params);
      const maybe: any = res;
      const data: Order[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      setOrders(data);
    } catch (e) {
      console.error('load orders failed', e);
    } finally {
      setLoading(false);
    }
  }

  function changeStatus(s?: number) {
    setStatus(s);
  }

  function goDetail(id: number) {
    const sid = currentStore?.id;
    const url = sid && sid > 0
      ? `/pages/order-detail/index?id=${id}&store_id=${sid}`
      : `/pages/order-detail/index?id=${id}`;
    Taro.navigateTo({ url });
  }

  return (
    <View data-testid="page-orders" style={{ padding: 12 }}>
      {currentStore && (
        <View style={{
          marginBottom: 8,
          padding: '6px 10px',
          borderWidth: 1,
          borderStyle: 'solid',
          borderColor: '#07c160',
          borderRadius: 16,
          display: 'inline-block',
          backgroundColor: '#f6ffed',
        }}>
          <Text style={{ color: '#389e0d' }}>当前门店：{currentStore.name}</Text>
        </View>
      )}
      {/* 状态切换 Tab（简化版） */}
      <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row' }}>
        <Button size="mini" onClick={() => changeStatus(undefined)}>
          全部
        </Button>
        <Button size="mini" onClick={() => changeStatus(1)} style={{ marginLeft: 8 }}>
          待支付
        </Button>
        <Button size="mini" onClick={() => changeStatus(2)} style={{ marginLeft: 8 }}>
          已支付
        </Button>
        <Button size="mini" onClick={() => changeStatus(4)} style={{ marginLeft: 8 }}>
          已完成
        </Button>
      </View>

      {loading && <Text>加载中...</Text>}
      {!loading && !orders.length && <Text>暂无订单</Text>}
      {orders.map((o) => (
        <View
          key={o.id}
          style={{
            marginBottom: 12,
            borderBottomWidth: 1,
            borderStyle: 'solid',
            borderColor: '#eee',
            paddingBottom: 8,
          }}
        >
          <Text>订单号: {o.order_no}</Text>
          <Text> 金额: {o.pay_amount}</Text>
          <Text> 状态: {o.status}</Text>
          <Button size="mini" style={{ marginTop: 4 }} onClick={() => goDetail(o.id)}>
            查看详情
          </Button>
        </View>
      ))}
    </View>
  );
}
