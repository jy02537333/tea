import React, { useEffect, useState } from 'react';
import { ScrollView, Text, View } from '@tarojs/components';
import Taro, { useReachBottom, usePullDownRefresh } from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import type { Order } from '../../services/types';

function formatStatus(order: Order): string {
  const status = Number((order as any).status || (order as any).order_status || 0);
  const payStatus = Number((order as any).pay_status || 0);
  if (status === 5) return '已取消';
  if (payStatus === 2 && (status === 2 || status === 4)) return '已支付';
  if (status === 1 && payStatus === 1) return '待支付';
  return '处理中';
}

function formatAmount(v: any): string {
  const num = Number(v || 0);
  return num.toFixed(2);
}

export default function MembershipOrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    void reload();
  }, []);

  usePullDownRefresh(() => {
    void reload();
  });

  useReachBottom(() => {
    if (!loading && hasMore) {
      void loadMore();
    }
  });

  async function reload() {
    setPage(1);
    setHasMore(true);
    setOrders([]);
    await fetchPage(1, true);
  }

  async function loadMore() {
    const next = page + 1;
    await fetchPage(next, false);
  }

  async function fetchPage(p: number, replace: boolean) {
    setLoading(true);
    try {
      const res = await listOrders({ page: p, limit: 20, status: 0 });
      const list = Array.isArray(res.data) ? res.data : [];
      // 仅保留会员订单（order_type === 4）
      const membershipOrders = list.filter((o: any) => Number(o.order_type) === 4);
      const merged = replace ? membershipOrders : [...orders, ...membershipOrders];
      setOrders(merged);
      setPage(p);
      // 如果这一页返回的会员订单不足 20 条，暂时认为没有更多了
      setHasMore(membershipOrders.length >= 20);
    } catch (err) {
      console.error('load membership orders failed', err);
      Taro.showToast({ title: '加载会员订单失败', icon: 'none' });
    } finally {
      setLoading(false);
      Taro.stopPullDownRefresh();
    }
  }

  return (
    <View style={{ minHeight: '100vh', backgroundColor: '#f5f6f8' }}>
      <View style={{ padding: 16 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>我的会员订单</Text>
      </View>
      <ScrollView
        scrollY
        style={{ height: 'calc(100vh - 48px)', padding: 16 }}
        lowerThreshold={60}
      >
        {orders.length === 0 && !loading && (
          <Text style={{ color: '#999' }}>暂无会员订单</Text>
        )}
        {orders.map((o) => (
          <View
            key={o.id}
            style={{
              marginBottom: 12,
              padding: 12,
              borderRadius: 8,
              backgroundColor: '#fff',
              borderWidth: 1,
              borderColor: '#eee',
            }}
          >
            <Text style={{ fontSize: 14, color: '#333' }}>订单号：{(o as any).order_no}</Text>
            <Text style={{ marginTop: 4, color: '#666' }}>
              金额：￥{formatAmount((o as any).pay_amount)}
            </Text>
            <Text style={{ marginTop: 4, color: '#666' }}>
              状态：{formatStatus(o)}
            </Text>
            <Text style={{ marginTop: 4, color: '#999', fontSize: 12 }}>
              下单时间：{(o as any).created_at || ''}
            </Text>
          </View>
        ))}
        {loading && (
          <Text style={{ color: '#999' }}>加载中...</Text>
        )}
        {!loading && hasMore && orders.length > 0 && (
          <Text style={{ color: '#999' }}>上拉加载更多...</Text>
        )}
        {!loading && !hasMore && orders.length > 0 && (
          <Text style={{ color: '#999' }}>已无更多订单</Text>
        )}
      </ScrollView>
    </View>
  );
}
