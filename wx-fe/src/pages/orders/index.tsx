import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import { Order } from '../../services/types';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<number | undefined>(undefined);

  useEffect(() => {
    void fetchOrders();
  }, [status]);

  async function fetchOrders() {
    setLoading(true);
    try {
      const res = await listOrders({ page: 1, limit: 20, status });
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
    Taro.navigateTo({ url: `/pages/order-detail/index?id=${id}` });
  }

  return (
    <View style={{ padding: 12 }}>
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
        <Button size="mini" onClick={() => changeStatus(3)} style={{ marginLeft: 8 }}>
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
