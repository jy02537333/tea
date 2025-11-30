import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import { listStores } from '../../services/stores';
import { cacheOrders, getCachedOrders } from '../../store/orders';
import type { Order } from '../../services/types';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [statusFilter, setStatusFilter] = useState<number | null>(null);
  const [storeFilter, setStoreFilter] = useState<number | null>(null);
  const [stores, setStores] = useState<{ label: string; value: number }[]>([]);

  const STATUS_OPTIONS: { label: string; value: number | null }[] = [
    { label: '全部', value: null },
    { label: '待支付', value: 1 },
    { label: '已支付', value: 2 },
    { label: '已发货', value: 3 },
    { label: '已完成', value: 4 },
    { label: '已取消', value: 9 },
  ];

  useEffect(() => { resetAndFetch(); }, [statusFilter, storeFilter]);
  useEffect(() => { fetchStores(); }, []);
    async function fetchStores() {
      try {
        const res = await listStores({ page: 1, limit: 100 });
        const maybe: any = res;
        const list = maybe?.data || maybe?.items || [];
        setStores(list.map((s: any) => ({ label: s.name, value: s.id })));
      } catch {}
    }
  useEffect(() => {
    // 优先显示缓存的列表以减少闪烁
    const cached = getCachedOrders();
    if (cached && cached.length > 0) {
      setOrders(cached);
    }
  }, []);

  function resetAndFetch() {
    setPage(1);
    setOrders([]);
    fetchOrders(1, true);
  }

  async function fetchOrders(nextPage = page, replace = false) {
    if (loading) return;
    setLoading(true);
    try {
      const params: any = { page: nextPage, limit: 20 };
      if (statusFilter !== null) params.status = statusFilter;
      if (storeFilter !== null) params.store_id = storeFilter;
      const res = await listOrders(params);
      const maybe = res as any;
      const list: Order[] = maybe?.data || maybe?.items || [];
      setHasMore(list.length === 20); // 简单判断还有更多
      setOrders(replace ? list : [...orders, ...list]);
      cacheOrders(replace ? list : [...orders, ...list]);
      setPage(nextPage);
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={{ padding: 12 }}>
      <View style={{ flexDirection: 'row', flexWrap: 'wrap', marginBottom: 8 }}>
        {STATUS_OPTIONS.map(opt => (
          <Button
            key={opt.label}
            size="mini"
            onClick={() => setStatusFilter(opt.value)}
            style={{ marginRight: 4, backgroundColor: statusFilter === opt.value ? '#1677ff' : '#ddd', color: statusFilter === opt.value ? '#fff' : '#000' }}
          >{opt.label}</Button>
        ))}
        <View style={{ marginLeft: 8 }}>
          <Text>门店：</Text>
          <View style={{ flexDirection: 'row', flexWrap: 'wrap' }}>
            {stores.map(s => (
              <Button
                key={s.value}
                size="mini"
                onClick={() => setStoreFilter(s.value)}
                style={{ marginRight: 4, backgroundColor: storeFilter === s.value ? '#52c41a' : '#ddd', color: storeFilter === s.value ? '#fff' : '#000' }}
              >{s.label}</Button>
            ))}
            <Button size="mini" onClick={() => setStoreFilter(null)} style={{ marginLeft: 4 }}>清除门店</Button>
          </View>
        </View>
      </View>
      {loading && <Text>加载中...</Text>}
      {!loading && orders.length === 0 && <Text>暂无订单</Text>}
      {orders.map((o) => (
        <View key={o.id} style={{ marginBottom: 10, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 6 }}>
          <Text>订单号: {o.order_no || o.id}</Text>
          <Text> 金额: {o.pay_amount}</Text>
          <Text> 状态: {renderStatus(o.status)}</Text>
          <Button onClick={() => Taro.navigateTo({ url: `/src/pages/order-detail/index?id=${o.id}` })}>查看详情</Button>
        </View>
      ))}
      <Button onClick={() => resetAndFetch()} disabled={loading}>刷新</Button>
      {hasMore && <Button onClick={() => fetchOrders(page + 1)} disabled={loading}>加载更多</Button>}
    </View>
  );
}

function renderStatus(s: any): string {
  const map: Record<string | number, string> = {
    0: '待处理',
    1: '待支付',
    2: '已支付',
    3: '已发货',
    4: '已完成',
    9: '已取消',
    pending: '待处理',
    paid: '已支付',
    shipped: '已发货',
    completed: '已完成',
    canceled: '已取消',
  };
  return map[s as any] || String(s ?? '未知');
}
