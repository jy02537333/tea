import React, { useEffect, useMemo, useState } from 'react';
import { Button, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { cancelOrder, confirmReceive, listOrders, payOrder } from '../../services/orders';
import type { Order } from '../../services/types';
import { formatAddress, parseAddressInfo } from '../../utils/address';
import { createTicket } from '../../services/tickets';

const STATUS_TEXT: Record<number, string> = {
  1: '待付款',
  2: '待发货',
  3: '配送中',
  4: '已完成',
  5: '已取消',
};

const PAY_STATUS_TEXT: Record<number, string> = {
  1: '未付款',
  2: '已付款',
  3: '退款中',
  4: '已退款',
};

export default function AfterSalePage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [actioning, setActioning] = useState<{ id: number; type: string } | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void fetchOrders();
  }, []);

  usePullDownRefresh(() => {
    void fetchOrders();
  });

  async function fetchOrders() {
    setLoading(true);
    try {
      const resp = await listOrders({ page: 1, limit: 20 });
      setOrders(resp?.data || []);
      setError(null);
    } catch (err: any) {
      console.error('load orders failed', err);
      setError(err?.message || '订单加载失败');
    } finally {
      setLoading(false);
      Taro.stopPullDownRefresh();
    }
  }

  function getStatusText(status?: number) {
    return STATUS_TEXT[status || 0] || `状态${status ?? '-'}`;
  }

  function getPayStatusText(payStatus?: number) {
    return PAY_STATUS_TEXT[payStatus || 0] || `支付状态${payStatus ?? '-'}`;
  }

  function canPay(order: Order) {
    return Number(order.status) === 1;
  }

  function canCancel(order: Order) {
    return Number(order.status) === 1;
  }

  function canConfirm(order: Order) {
    return Number(order.status) === 3;
  }

  async function handleAction(order: Order, type: 'pay' | 'cancel' | 'receive') {
    setActioning({ id: order.id, type });
    try {
      if (type === 'pay') {
        await payOrder(order.id);
      } else if (type === 'cancel') {
        await cancelOrder(order.id, '用户申请取消');
      } else {
        await confirmReceive(order.id);
      }
      Taro.showToast({ title: '操作成功', icon: 'success' });
      await fetchOrders();
    } catch (err) {
      console.error('after-sale action failed', err);
      Taro.showToast({ title: '操作失败', icon: 'none' });
    } finally {
      setActioning(null);
    }
  }

  async function handleRefundInquiry(order: Order) {
    try {
      await createTicket({
        type: 'refund',
        source: 'miniapp_order',
        order_id: order.id,
        title: `退款进度咨询 ${order.order_no || ''}`.trim(),
        content: '请协助查询该订单的退款进度。',
      });
      Taro.showToast({ title: '已提交退款咨询', icon: 'none' });
    } catch (err: any) {
      console.error('refund inquiry failed', err);
      Taro.showToast({ title: err?.message || '提交失败，请稍后再试', icon: 'none' });
    }
  }

  return (
    <View style={{ padding: 16, backgroundColor: '#f5f6f8', minHeight: '100vh' }}>
      <View style={{ marginBottom: 12 }}>
        <Text style={{ fontSize: 16, color: '#666' }}>可对最近订单执行取消、支付、确认收货等操作。</Text>
      </View>
      {error && <Text style={{ color: '#ff4d4f', marginBottom: 12 }}>{error}</Text>}
      {!orders.length && !error && (
        <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16 }}>
          <Text style={{ color: '#999' }}>{loading ? '加载中...' : '暂无订单'}</Text>
        </View>
      )}
      {orders.map((order) => {
        const parsedAddress = parseAddressInfo(order.address_info);
        return (
          <View key={order.id} style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16, marginBottom: 12 }}>
            <View style={{ display: 'flex', justifyContent: 'space-between' }}>
              <Text style={{ fontWeight: 'bold' }}>订单号：{order.order_no}</Text>
              <Text style={{ color: '#1677ff' }}>{getStatusText(Number(order.status))}</Text>
            </View>
            <View style={{ marginTop: 6 }}>
              <Text style={{ color: '#8c8c8c', fontSize: 12 }}>支付状态：{getPayStatusText(Number(order.pay_status as any))}</Text>
            </View>
            {parsedAddress && <Text style={{ color: '#666', marginTop: 6 }}>{formatAddress(parsedAddress)}</Text>}
            {order.remark && <Text style={{ color: '#999', fontSize: 12 }}>备注：{order.remark}</Text>}
            <View style={{ display: 'flex', gap: 8, marginTop: 12, flexWrap: 'wrap' }}>
              {canPay(order) && (
                <Button
                  size="mini"
                  loading={actioning?.id === order.id && actioning?.type === 'pay'}
                  onClick={() => handleAction(order, 'pay')}
                >
                  去支付
                </Button>
              )}
              {canCancel(order) && (
                <Button
                  size="mini"
                  loading={actioning?.id === order.id && actioning?.type === 'cancel'}
                  onClick={() => handleAction(order, 'cancel')}
                >
                  取消订单
                </Button>
              )}
              {canConfirm(order) && (
                <Button
                  size="mini"
                  loading={actioning?.id === order.id && actioning?.type === 'receive'}
                  onClick={() => handleAction(order, 'receive')}
                >
                  确认收货
                </Button>
              )}
              <Button size="mini" onClick={() => Taro.navigateTo({ url: `/pages/order-detail/index?id=${order.id}` })}>
                查看详情
              </Button>
              {(Number(order.pay_status as any) === 3 || Number(order.pay_status as any) === 4) && (
                <Button size="mini" type="warn" onClick={() => handleRefundInquiry(order)}>
                  退款进度咨询
                </Button>
              )}
            </View>
          </View>
        );
      })}
    </View>
  );
}
