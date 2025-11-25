import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import { getOrder } from '../../services/orders';
import { Order } from '../../services/types';

export default function OrderDetail({ id }: { id: number }) {
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  async function fetch() {
    setLoading(true);
    try {
      const res = await getOrder(id);
      setOrder(res as Order);
    } finally {
      setLoading(false);
    }
  }

  if (loading) return <Text>加载中...</Text>;
  if (!order) return null;

  return (
    <View style={{ padding: 12 }}>
      <Text>订单号: {order.order_no}</Text>
      <Text>金额: {order.pay_amount}</Text>
      <Text>状态: {order.status}</Text>
      <Text>收货: {order.address_info}</Text>
    </View>
  );
}
