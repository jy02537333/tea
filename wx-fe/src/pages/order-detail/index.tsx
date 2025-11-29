import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getOrder, cancelOrder, payOrder, receiveOrder } from '../../services/orders';
import type { Order } from '../../services/types';

export default function OrderDetailPage() {
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const params = Taro.getCurrentInstance().router?.params || {} as any;
    const id = Number(params.id || 0);
    if (id) fetch(id);
  }, []);

  async function fetch(id: number) {
    setLoading(true);
    try {
      const o = await getOrder(id);
      setOrder(o);
    } finally {
      setLoading(false);
    }
  }

  async function onCancel() {
    if (!order) return;
    try {
      await cancelOrder(order.id, '用户取消');
      Taro.showToast({ title: '已取消', icon: 'success' });
      fetch(order.id);
    } catch { Taro.showToast({ title: '取消失败', icon: 'error' }); }
  }
  async function onPay() {
    if (!order) return;
    try {
      await payOrder(order.id);
      Taro.showToast({ title: '已支付', icon: 'success' });
      fetch(order.id);
    } catch { Taro.showToast({ title: '支付失败', icon: 'error' }); }
  }
  async function onReceive() {
    if (!order) return;
    try {
      await receiveOrder(order.id);
      Taro.showToast({ title: '已收货', icon: 'success' });
      fetch(order.id);
    } catch { Taro.showToast({ title: '操作失败', icon: 'error' }); }
  }

  return (
    <View style={{ padding: 12 }}>
      {!order && <Text>加载中...</Text>}
      {order && (
        <View>
          <Text>订单号: {order.order_no || order.id}</Text>
          <Text> 金额: {order.pay_amount}</Text>
          <Text> 状态: {order.status}</Text>
          <View style={{ flexDirection: 'row', marginTop: 8 }}>
            <Button onClick={onCancel} disabled={loading}>取消订单</Button>
            <Button onClick={onPay} disabled={loading}>支付</Button>
            <Button onClick={onReceive} disabled={loading}>确认收货</Button>
          </View>
        </View>
      )}
    </View>
  );
}
