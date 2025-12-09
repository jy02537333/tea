import React, { useEffect, useState } from 'react';
import { View, Text, Input, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCart } from '../../services/cart';
import { createOrderFromCart } from '../../services/orders';
import { CartItem } from '../../services/types';
import { formatAddress, loadDefaultAddress, saveDefaultAddress } from '../../utils/address';

export default function CheckoutPage() {
  const [items, setItems] = useState<CartItem[]>([]);
  const [address, setAddress] = useState('');
  const [remark, setRemark] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    void fetchCart();
    void preloadAddress();
  }, []);

  async function fetchCart() {
    try {
      const res = await listCart();
      setItems((res as any) || []);
    } catch (e) {
      console.error('load cart for checkout failed', e);
    }
  }

  async function preloadAddress() {
    try {
      const stored = await loadDefaultAddress();
      if (stored) {
        const fallback = stored.full || formatAddress(stored);
        if (fallback && !address) {
          setAddress(fallback);
        }
      }
    } catch (error) {
      console.error('load default address failed', error);
    }
  }

  function calcTotal() {
    // 这里只做展示用，真实金额以后从后端订单返回
    return items.length;
  }

  async function handleSubmit() {
    if (!items.length) {
      Taro.showToast({ title: '购物车为空', icon: 'none', duration: 1500 });
      return;
    }
    setSubmitting(true);
    try {
      const payload = {
        delivery_type: 2, // 简化：2 = 配送
        address_info: address || undefined,
        remark: remark || undefined,
      };
      const order = await createOrderFromCart(payload as any);
      if (address.trim()) {
        await saveDefaultAddress({
          full: address.trim(),
          detail: address.trim(),
          orderId: (order as any)?.id,
          orderNo: (order as any)?.order_no,
          updatedAt: new Date().toISOString(),
          timestamp: Date.now(),
        });
      }
      Taro.showToast({ title: '下单成功', icon: 'success', duration: 1500 });
      if ((order as any)?.id) {
        Taro.navigateTo({ url: `/pages/order-detail/index?id=${(order as any).id}` });
      }
    } catch (e) {
      console.error('create order failed', e);
      Taro.showToast({ title: '下单失败', icon: 'none', duration: 1500 });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <View style={{ padding: 12 }}>
      <Text>确认订单（共 {items.length} 件）</Text>
      <Text>金额（示意）: {calcTotal()}</Text>

      <View style={{ marginTop: 12 }}>
        <Text>收货地址</Text>
        <Input
          type="text"
          placeholder="请输入收货地址"
          value={address}
          onInput={(e) => setAddress((e.detail as any).value)}
        />
        <Text style={{ fontSize: 12, color: '#999' }}>可在“我的-收货地址”设置默认地址</Text>
      </View>

      <View style={{ marginTop: 12 }}>
        <Text>备注</Text>
        <Input
          type="text"
          placeholder="可填写口味、送达时间等"
          value={remark}
          onInput={(e) => setRemark((e.detail as any).value)}
        />
      </View>

      <View style={{ marginTop: 16 }}>
        <Button type="primary" disabled={submitting} onClick={handleSubmit}>
          {submitting ? '提交中...' : '提交订单'}
        </Button>
      </View>
    </View>
  );
}
