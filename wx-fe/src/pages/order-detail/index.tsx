import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Image, Textarea } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { cancelOrder, confirmReceive, getOrder, payOrder } from '../../services/orders';
import { Order, OrderItem } from '../../services/types';
import { createTicket } from '../../services/tickets';

type ActionKey = 'cancel' | 'pay' | 'confirm' | null;

const STATUS_TEXT: Record<number, string> = {
  1: '待支付',
  2: '已付款',
  3: '配送中',
  4: '已完成',
  5: '已取消',
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

  const loadOrder = useCallback(async () => {
    if (!orderId) return;
    setLoading(true);
    try {
      const data = await getOrder(orderId);
      setOrder(data.order);
      setItems(Array.isArray(data.items) ? data.items : []);
      setAddress(parseAddress(data.order?.address_info));
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

  const numericStatus = useMemo(() => toNumber(order?.status), [order?.status]);
  const statusText = useMemo(() => {
    if (!numericStatus) return order?.status ? String(order.status) : '--';
    return STATUS_TEXT[numericStatus] || String(order?.status ?? '--');
  }, [numericStatus, order?.status]);

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
      <View style={{ marginBottom: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{statusText}</Text>
        <Text style={{ display: 'block', marginTop: 8 }}>订单号：{order.order_no}</Text>
        <Text style={{ display: 'block', marginTop: 4 }}>下单时间：{order.created_at ?? '--'}</Text>
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
