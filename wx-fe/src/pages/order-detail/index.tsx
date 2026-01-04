import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Image, Textarea } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { cancelOrder, confirmReceive, getOrder, payOrder } from '../../services/orders';
import { Order, OrderItem, Store, Refund } from '../../services/types';
import { getStore } from '../../services/stores';
import { createTicket } from '../../services/tickets';
import { listMyRefunds } from '../../services/refunds';

type ActionKey = 'cancel' | 'pay' | 'confirm' | null;

const STATUS_TEXT: Record<number, string> = {
  1: '待支付',
  2: '已付款',
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

function toNumber(value?: number | string): number | undefined {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = parseInt(value, 10);
    return Number.isNaN(parsed) ? undefined : parsed;
  }
  return undefined;
}

function formatAmount(value?: number | string): string {
  if (value === undefined || value === null) return '--';
  const num = typeof value === 'string' ? parseFloat(value) : value;
  if (Number.isNaN(num)) return '--';
  return num.toFixed(2);
}

function parseAddress(info?: string | Record<string, any>): Record<string, any> | null {
  if (!info) return null;
  if (typeof info === 'object') return info;
  try {
    return JSON.parse(info);
  } catch (_) {
    return { raw: info };
  }
}

export default function OrderDetail({ id }: { id?: number }) {
  const router = useRouter();
  const paramId = router?.params?.id ?? router?.params?.orderId;
  const orderId = useMemo(() => {
    if (typeof id === 'number' && !Number.isNaN(id)) return id;
    const parsed = paramId ? Number(paramId) : NaN;
    return Number.isNaN(parsed) ? undefined : parsed;
  }, [id, paramId]);

  const [order, setOrder] = useState<Order | null>(null);
  const [items, setItems] = useState<OrderItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<ActionKey>(null);
  const [address, setAddress] = useState<Record<string, any> | null>(null);
  const [complaintContent, setComplaintContent] = useState('');
  const [complaintLoading, setComplaintLoading] = useState(false);
  const [polling, setPolling] = useState(true);
  const [lastRefreshedAt, setLastRefreshedAt] = useState<number | null>(null);
  const POLL_MS = 3000;
  const [currentStore, setCurrentStore] = useState<Store | null>(null);
  const [refunds, setRefunds] = useState<Refund[]>([]);
  const [refundsLoading, setRefundsLoading] = useState(false);

  const loadOrder = useCallback(async () => {
    if (!orderId) return;
    setLoading(true);
    try {
      const data = await getOrder(orderId);
      setOrder(data.order);
      setItems(Array.isArray(data.items) ? data.items : []);
      setAddress(parseAddress(data.order?.address_info));
      setLastRefreshedAt(Date.now());
    } catch (err: any) {
      console.error('load order failed', err);
      Taro.showToast({ title: err?.message || '加载订单失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => {
    if (orderId) void loadOrder();
  }, [orderId, loadOrder]);

  // 加载门店信息：优先使用订单的 store_id，其次尝试路由参数
  useEffect(() => {
    const fromOrder = toNumber(order?.store_id);
    const fromRouter = router?.params?.store_id ? Number(router.params.store_id) : undefined;
    const sid = fromOrder || (fromRouter && Number.isFinite(fromRouter) && fromRouter > 0 ? fromRouter : undefined);
    if (!sid || (currentStore && currentStore.id === sid)) return;
    (async () => {
      try {
        const s = await getStore(sid);
        setCurrentStore(s as Store);
      } catch (_) {
        // ignore store load error
      }
    })();
  }, [order?.store_id, router?.params?.store_id]);

  const numericStatus = useMemo(() => toNumber(order?.status), [order?.status]);
  const statusText = useMemo(() => {
    if (!numericStatus) return order?.status ? String(order.status) : '--';
    return STATUS_TEXT[numericStatus] || String(order?.status ?? '--');
  }, [numericStatus, order?.status]);

  const numericPayStatus = useMemo(() => {
    const v = order?.pay_status;
    if (typeof v === 'number') return v;
    if (typeof v === 'string') {
      const p = parseInt(v, 10);
      return Number.isNaN(p) ? undefined : p;
    }
    return undefined;
  }, [order?.pay_status]);

  // 拉取该订单的退款记录（当进入退款态时）
  useEffect(() => {
    if (!orderId) return;
    const ps = numericPayStatus;
    if (ps === 3 || ps === 4) {
      (async () => {
        setRefundsLoading(true);
        try {
          const resp = await listMyRefunds({ order_id: orderId, page: 1, limit: 20 });
          setRefunds(resp?.data || []);
        } catch (err) {
          // 静默失败，仅在 UI 保留基本提示
        } finally {
          setRefundsLoading(false);
        }
      })();
    } else {
      setRefunds([]);
    }
  }, [orderId, numericPayStatus]);

  const payStatusText = useMemo(() => {
    if (!numericPayStatus) return order?.pay_status ? String(order.pay_status) : '--';
    return PAY_STATUS_TEXT[numericPayStatus] || String(order?.pay_status ?? '--');
  }, [numericPayStatus, order?.pay_status]);

  const addressText = useMemo(() => {
    if (!address) {
      if (typeof order?.address_info === 'string') return order.address_info || '--';
      return '--';
    }
    const parts: string[] = [];
    if (address.contact || address.name) parts.push(address.contact || address.name);
    if (address.phone) parts.push(address.phone);
    if (address.full_address || address.detail || address.address) parts.push(address.full_address || address.detail || address.address);
    if (!parts.length && address.raw) parts.push(address.raw);
    return parts.join(' / ') || '--';
  }, [address, order?.address_info]);

  const showPay = numericStatus === 1;
  const showCancel = numericStatus === 1;
  const showConfirm = numericStatus === 3;

  const TERMINAL_STATUSES = useMemo(() => new Set([4, 5]), []);

  useEffect(() => {
    if (!polling || !orderId) return;
    if (TERMINAL_STATUSES.has((numericStatus ?? -1))) return;
    const timer = setInterval(() => {
      void loadOrder();
    }, POLL_MS);
    return () => clearInterval(timer);
  }, [polling, orderId, numericStatus, loadOrder, TERMINAL_STATUSES]);

  const runAction = async (key: Exclude<ActionKey, null>, fn: () => Promise<void>, successText: string) => {
    if (!orderId) return;
    setActionLoading(key);
    try {
      await fn();
      Taro.showToast({ title: successText, icon: 'success' });
      await loadOrder();
    } catch (err: any) {
      console.error(`${key} order failed`, err);
      Taro.showToast({ title: err?.message || '操作失败', icon: 'none' });
    } finally {
      setActionLoading(null);
    }
  };

  const handlePay = () => runAction('pay', () => payOrder(orderId!), '支付成功');

  const handleCancel = async () => {
    const result = await Taro.showModal({ title: '取消订单', content: '确认要取消该订单吗？' });
    if (!result.confirm || !orderId) return;
    await runAction('cancel', () => cancelOrder(orderId, '用户取消'), '已取消');
  };

  const handleConfirm = async () => {
    const result = await Taro.showModal({ title: '确认收货', content: '请确认已经收到商品并无售后问题。' });
    if (!result.confirm || !orderId) return;
    await runAction('confirm', () => confirmReceive(orderId), '已确认');
  };

  const handleSubmitComplaint = async () => {
    if (!orderId) return;
    if (!complaintContent.trim()) {
      Taro.showToast({ title: '请先填写投诉内容', icon: 'none' });
      return;
    }
    setComplaintLoading(true);
    try {
      await createTicket({
        type: 'order',
        source: 'miniapp_order',
        order_id: orderId,
        title: `订单投诉 ${order?.order_no || ''}`.trim(),
        content: complaintContent.trim(),
      });
      Taro.showToast({ title: '投诉已提交，我们会尽快处理', icon: 'none' });
      setComplaintContent('');
    } catch (err: any) {
      console.error('create complaint ticket failed', err);
      Taro.showToast({ title: err?.message || '提交失败，请稍后再试', icon: 'none' });
    } finally {
      setComplaintLoading(false);
    }
  };

  if (!orderId) {
    return (
      <View style={{ padding: 16 }}>
        <Text>未获取到有效的订单编号</Text>
      </View>
    );
  }

  if (loading && !order) {
    return (
      <View style={{ padding: 16 }}>
        <Text>订单加载中...</Text>
      </View>
    );
  }

  if (!order) {
    return (
      <View style={{ padding: 16 }}>
        <Text>未找到订单信息</Text>
      </View>
    );
  }

  return (
    <View style={{ padding: 16 }}>
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
      <View style={{ marginBottom: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{statusText}</Text>
        <Text style={{ display: 'block', marginTop: 8 }}>订单号：{order.order_no}</Text>
        <Text style={{ display: 'block', marginTop: 4 }}>下单时间：{order.created_at ?? '--'}</Text>
        <Text style={{ display: 'block', marginTop: 4, color: '#8c8c8c', fontSize: 12 }}>支付状态：{payStatusText}</Text>
        <View style={{ display: 'flex', gap: 8, alignItems: 'center', marginTop: 8 }}>
          <Text style={{ color: '#666', fontSize: 12 }}>
            自动刷新：{polling && !TERMINAL_STATUSES.has((numericStatus ?? -1)) ? '开启' : '关闭'}
            {TERMINAL_STATUSES.has((numericStatus ?? -1)) ? '（已到终态）' : ''}
          </Text>
          <Button size="mini" onClick={() => setPolling((p) => !p)}>
            {polling ? '停止自动刷新' : '开启自动刷新'}
          </Button>
          <Button size="mini" onClick={() => void loadOrder()}>
            刷新状态
          </Button>
          {lastRefreshedAt && (
            <Text style={{ color: '#999', fontSize: 12 }}>
              最近刷新：{new Date(lastRefreshedAt).toLocaleTimeString()}
            </Text>
          )}
        </View>
      </View>

      <View style={{ padding: 12, backgroundColor: '#f7f7f7', borderRadius: 8, marginBottom: 16 }}>
        <Text style={{ fontWeight: 'bold' }}>收货信息</Text>
        <Text style={{ display: 'block', marginTop: 6 }}>{addressText}</Text>
        <Text style={{ display: 'block', marginTop: 4 }}>配送方式：{order.delivery_type === 2 ? '配送' : '到店自取'}</Text>
      </View>

      <View style={{ marginBottom: 16 }}>
        <Text style={{ fontWeight: 'bold' }}>商品信息</Text>
        {items.map((item) => (
          <View
            key={item.id}
            style={{
              marginTop: 12,
              display: 'flex',
              flexDirection: 'row',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 12,
            }}
          >
            {item.image ? <Image src={item.image} style={{ width: 60, height: 60, borderRadius: 4 }} mode="aspectFill" /> : <View style={{ width: 60, height: 60, backgroundColor: '#f2f2f2', borderRadius: 4 }} />}
            <View style={{ flex: 1 }}>
              <Text style={{ display: 'block' }}>{item.product_name || `商品#${item.product_id}`}</Text>
              {item.sku_name && <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>规格：{item.sku_name}</Text>}
            </View>
            <View style={{ alignItems: 'flex-end', display: 'flex', flexDirection: 'column' }}>
              <Text>￥{formatAmount(item.price || item.amount)}</Text>
              <Text style={{ color: '#666', fontSize: 12 }}>x{item.quantity}</Text>
            </View>
          </View>
        ))}
        {!items.length && <Text style={{ display: 'block', marginTop: 8 }}>暂无商品信息</Text>}
      </View>

      <View style={{ padding: 12, backgroundColor: '#f7f7f7', borderRadius: 8, marginBottom: 16 }}>
        <Text style={{ display: 'block' }}>商品金额：￥{formatAmount(order.total_amount)}</Text>
        <Text style={{ display: 'block', marginTop: 4 }}>优惠：-￥{formatAmount(order.discount_amount)}</Text>
        <Text style={{ display: 'block', marginTop: 4 }}>配送费：￥{formatAmount(order.delivery_fee)}</Text>
        <Text style={{ display: 'block', marginTop: 8, fontSize: 16, fontWeight: 'bold' }}>实付金额：￥{formatAmount(order.pay_amount)}</Text>
      </View>

      <View style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
        {showCancel && (
          <Button size="mini" loading={actionLoading === 'cancel'} onClick={handleCancel} type="warn">
            取消订单
          </Button>
        )}
        {showPay && (
          <Button size="mini" loading={actionLoading === 'pay'} onClick={handlePay} type="primary">
            立即支付
          </Button>
        )}
        {showConfirm && (
          <Button size="mini" loading={actionLoading === 'confirm'} onClick={handleConfirm} type="primary">
            确认收货
          </Button>
        )}
      </View>

      {(numericPayStatus === 3 || numericPayStatus === 4) && (
        <View style={{ marginTop: 16, padding: 12, backgroundColor: '#fffbe6', borderRadius: 8, border: '1px solid #ffe58f' }}>
          <Text style={{ fontWeight: 'bold', display: 'block' }}>退款进度</Text>
          <Text style={{ display: 'block', marginTop: 6, color: '#ad8b00' }}>{numericPayStatus === 3 ? '退款处理中，请耐心等待' : '已退款完成'}</Text>
          {order.paid_at && <Text style={{ display: 'block', marginTop: 4, color: '#8c8c8c', fontSize: 12 }}>支付时间：{order.paid_at}</Text>}
          <View style={{ marginTop: 8 }}>
            {refundsLoading && <Text style={{ color: '#8c8c8c' }}>退款记录加载中...</Text>}
            {!refundsLoading && !refunds.length && (
              <Text style={{ color: '#8c8c8c' }}>暂无退款记录</Text>
            )}
            {!refundsLoading && refunds.map((rf) => (
              <View key={rf.id} style={{ marginTop: 8, padding: 8, backgroundColor: '#fff', borderRadius: 8 }}>
                <Text style={{ display: 'block' }}>退款单号：{rf.refund_no}</Text>
                <Text style={{ display: 'block', marginTop: 2 }}>退款金额：￥{formatAmount(rf.refund_amount)}</Text>
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
                if (!orderId) return;
                setRefundsLoading(true);
                try {
                  const resp = await listMyRefunds({ order_id: orderId, page: 1, limit: 20 });
                  setRefunds(resp?.data || []);
                  Taro.showToast({ title: '已刷新退款记录', icon: 'none' });
                } catch (_) {
                  Taro.showToast({ title: '刷新失败', icon: 'none' });
                } finally {
                  setRefundsLoading(false);
                }
              }}>刷新退款记录</Button>
            </View>
          </View>
        </View>
      )}

      <View style={{ marginTop: 20 }}>
        <Text style={{ fontWeight: 'bold', display: 'block', marginBottom: 8 }}>订单有问题？提交投诉</Text>
        <Textarea
          style={{
            minHeight: 100,
            padding: 8,
            borderRadius: 8,
            border: '1px solid #ddd',
          }}
          placeholder="请简单描述您遇到的问题，我们会安排客服跟进"
          value={complaintContent}
          onInput={(e) => setComplaintContent((e.detail as any).value)}
        />
        <Button
          style={{ marginTop: 12 }}
          size="mini"
          type="warn"
          loading={complaintLoading}
          disabled={complaintLoading}
          onClick={handleSubmitComplaint}
        >
          提交投诉
        </Button>
      </View>
    </View>
  );
}
