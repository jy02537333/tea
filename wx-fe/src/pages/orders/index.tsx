import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import type { Order } from '../../services/types';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => { fetchOrders(); }, []);

  async function fetchOrders() {
    setLoading(true);
    try {
      const res = await listOrders({ page: 1, limit: 20 });
      const maybe = res as any;
      setOrders(maybe?.data || maybe?.items || []);
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={{ padding: 12 }}>
      {orders.map((o) => (
        <View key={o.id} style={{ marginBottom: 10, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 6 }}>
          <Text>订单号: {o.order_no || o.id}</Text>
          <Text> 金额: {o.pay_amount}</Text>
          <Button onClick={() => Taro.navigateTo({ url: `/src/pages/order-detail/index?id=${o.id}` })}>查看详情</Button>
        </View>
      ))}
      <Button onClick={fetchOrders} disabled={loading}>刷新</Button>
    </View>
  );
}
