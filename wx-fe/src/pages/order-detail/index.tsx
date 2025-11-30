import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getOrder, cancelOrder, payOrder, receiveOrder } from '../../services/orders';
import { cacheOrder, getCachedOrder } from '../../store/orders';
import type { Order } from '../../services/types';

export default function OrderDetailPage() {
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const params = Taro.getCurrentInstance().router?.params || {} as any;
    const id = Number(params.id || 0);
    if (id) {
      const cached = getCachedOrder(id);
      if (cached) setOrder(cached);
      fetch(id);
    }
  }, []);

  async function fetch(id: number) {
    setLoading(true);
    try {
      const o = await getOrder(id);
      setOrder(o);
      cacheOrder(o);
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
          <View style={{ marginTop: 8 }}>
            <Text> 商品项：</Text>
            {(order.items && order.items.length > 0) ? (
              order.items.map((it, idx) => (
                <View key={idx} style={{ flexDirection: 'row', justifyContent: 'space-between', paddingTop: 4, paddingBottom: 4 }}>
                  <Text>#{idx + 1} 商品ID: {it.product_id}</Text>
                  <Text> 数量: {it.quantity}</Text>
                  {typeof (it as any).price !== 'undefined' && <Text> 价格: {(it as any).price}</Text>}
                </View>
              ))
            ) : (
              <Text>无订单项或未返回</Text>
            )}
          </View>
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
