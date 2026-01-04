import React, { useEffect, useMemo, useState } from 'react';
import { Button, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh, useRouter } from '@tarojs/taro';
import { cancelOrder, confirmReceive, listOrders, payOrder } from '../../services/orders';
import type { Order, Refund } from '../../services/types';
import { formatAddress, parseAddressInfo } from '../../utils/address';
import { createTicket } from '../../services/tickets';
import { listMyRefunds } from '../../services/refunds';

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
  const router = useRouter();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [actioning, setActioning] = useState<{ id: number; type: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [expandedRefunds, setExpandedRefunds] = useState<Record<number, boolean>>({});
  const [refundsMap, setRefundsMap] = useState<Record<number, Refund[]>>({});
  const [refundsLoadingMap, setRefundsLoadingMap] = useState<Record<number, boolean>>({});
  const showMockRefund = useMemo(() => {
    const mock = router?.params?.mock_refund;
    return mock === '1' || mock === 'true';
  }, [router?.params?.mock_refund]);

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
      const list = resp?.data || [];
      setOrders(list);
      // 如果 mock_refund 开启，预展开第一条订单的退款时间线并填充示例
      if (showMockRefund && list.length) {
        const first = list[0];
        setExpandedRefunds((m) => ({ ...m, [first.id]: true }));
        setRefundsMap((m) => ({ ...m, [first.id]: [{
          id: 998001,
          order_id: first.id,
          payment_id: 0,
          refund_no: 'RF-MOCK-AFTERSALE-001',
          refund_amount: String(first.pay_amount ?? '0'),
          refund_reason: '示例退款用于截图展示（售后页）',
          status: 1,
          created_at: new Date().toISOString(),
        }] }));
      }
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

  async function toggleRefundTimeline(order: Order) {
    const opened = !!expandedRefunds[order.id];
    const next = { ...expandedRefunds, [order.id]: !opened };
    setExpandedRefunds(next);
    if (!opened) {
      // opening: fetch refunds if not loaded
      if (!refundsMap[order.id]) {
        setRefundsLoadingMap((m) => ({ ...m, [order.id]: true }));
        try {
          const resp = await listMyRefunds({ order_id: order.id, page: 1, limit: 20 });
          setRefundsMap((m) => ({ ...m, [order.id]: resp?.data || [] }));
        } catch (_) {
          // ignore
        } finally {
          setRefundsLoadingMap((m) => ({ ...m, [order.id]: false }));
        }
      }
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
                <>
                  <Button size="mini" type="warn" onClick={() => handleRefundInquiry(order)}>
                    退款进度咨询
                  </Button>
                  <Button size="mini" onClick={() => toggleRefundTimeline(order)}>
                    {expandedRefunds[order.id] ? '收起退款进度' : '查看退款进度'}
                  </Button>
                </>
              )}
            </View>
            {expandedRefunds[order.id] && (
              <View style={{ marginTop: 10, padding: 10, backgroundColor: '#fffbe6', borderRadius: 8, border: '1px solid #ffe58f' }}>
                <Text style={{ fontWeight: 'bold', display: 'block' }}>退款进度</Text>
                {refundsLoadingMap[order.id] && <Text style={{ color: '#8c8c8c' }}>加载中...</Text>}
                {!refundsLoadingMap[order.id] && (!refundsMap[order.id] || !refundsMap[order.id].length) && (
                  <Text style={{ color: '#8c8c8c' }}>暂无退款记录</Text>
                )}
                {!refundsLoadingMap[order.id] && refundsMap[order.id] && refundsMap[order.id].map((rf) => (
                  <View key={rf.id} style={{ marginTop: 8, padding: 8, backgroundColor: '#fff', borderRadius: 8 }}>
                    <Text style={{ display: 'block' }}>退款单号：{rf.refund_no}</Text>
                    <Text style={{ display: 'block', marginTop: 2 }}>退款金额：￥{String(rf.refund_amount)}</Text>
                    <Text style={{ display: 'block', marginTop: 2, color: '#8c8c8c', fontSize: 12 }}>
                      状态：{rf.status === 1 ? '申请中' : rf.status === 2 ? '退款成功' : '退款失败'}
                    </Text>
                    <Text style={{ display: 'block', marginTop: 2, color: '#8c8c8c', fontSize: 12 }}>申请时间：{rf.created_at || '-'}</Text>
                    {rf.refunded_at && (
                      <Text style={{ display: 'block', marginTop: 2, color: '#8c8c8c', fontSize: 12 }}>退款完成时间：{rf.refunded_at}</Text>
                    )}
                    {rf.refund_reason && (
                      <Text style={{ display: 'block', marginTop: 2, color: '#999' }}>原因：{rf.refund_reason}</Text>
                    )}
                  </View>
                ))}
                <View style={{ marginTop: 8 }}>
                  <Button size="mini" onClick={async () => {
                    setRefundsLoadingMap((m) => ({ ...m, [order.id]: true }));
                    try {
                      const resp = await listMyRefunds({ order_id: order.id, page: 1, limit: 20 });
                      setRefundsMap((m) => ({ ...m, [order.id]: resp?.data || [] }));
                      Taro.showToast({ title: '已刷新退款记录', icon: 'none' });
                    } catch (_) {
                      Taro.showToast({ title: '刷新失败', icon: 'none' });
                    } finally {
                      setRefundsLoadingMap((m) => ({ ...m, [order.id]: false }));
                    }
                  }}>刷新退款记录</Button>
                </View>
              </View>
            )}
          </View>
        );
      })}
    </View>
  );
}
