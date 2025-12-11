import React, { useEffect, useState } from 'react';
import { View, Text, Input, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCart } from '../../services/cart';
import { createOrderFromCart, getAvailableCouponsForOrder } from '../../services/orders';
import { CartItem, UserCoupon } from '../../services/types';
import { formatAddress, loadDefaultAddress, saveDefaultAddress } from '../../utils/address';

export default function CheckoutPage() {
  const [items, setItems] = useState<CartItem[]>([]);
  const [totalAmount, setTotalAmount] = useState(0);
  const [address, setAddress] = useState('');
  const [remark, setRemark] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [couponSummary, setCouponSummary] = useState('暂不使用优惠券');
  const [selectedUserCouponId, setSelectedUserCouponId] = useState<number | undefined>(undefined);
  const [availableCoupons, setAvailableCoupons] = useState<UserCoupon[]>([]);
  const [showCouponList, setShowCouponList] = useState(false);

  useEffect(() => {
    void fetchCart();
    void preloadAddress();
  }, []);

  async function fetchCart() {
    try {
      const res = await listCart();
      const nextItems = (res as any) || [];
      setItems(nextItems);
      const sum = Array.isArray(nextItems)
        ? nextItems.reduce((acc: number, it: any) => {
            const price = Number((it.price as any) ?? 0);
            const quantity = Number((it.quantity as any) ?? 0);
            if (!Number.isFinite(price) || !Number.isFinite(quantity)) return acc;
            return acc + price * quantity;
          }, 0)
        : 0;
      setTotalAmount(sum);
      void refreshAvailableCoupons(sum);
    } catch (e) {
      console.error('load cart for checkout failed', e);
    }
  }

  async function refreshAvailableCoupons(sum: number) {
    try {
      if (!sum || sum <= 0) {
        setCouponSummary('暂无可用优惠券');
        setSelectedUserCouponId(undefined);
        setAvailableCoupons([]);
        return;
      }
      const payload = { order_amount: String(sum) };
      const data = await getAvailableCouponsForOrder(payload);
      const list = data.available || [];
      setAvailableCoupons(list);
      if (list.length > 0) {
        const first = list[0];
        const label = buildCouponLabel(first);
        setSelectedUserCouponId(first.id);
        setCouponSummary(`已自动选择：${label}`);
      } else {
        setSelectedUserCouponId(undefined);
        setCouponSummary('暂无可用优惠券');
      }
    } catch (error) {
      console.error('load available coupons for order failed', error);
      setCouponSummary('优惠券加载失败，可稍后重试');
      setSelectedUserCouponId(undefined);
      setAvailableCoupons([]);
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
    return totalAmount || items.length;
  }

  function buildCouponLabel(userCoupon: UserCoupon): string {
    const coupon = userCoupon.coupon;
    let label = coupon.name;
    if (coupon.type === 1 && coupon.amount) {
      label = `${coupon.name} - 满减¥${coupon.amount}`;
    } else if (coupon.type === 2 && coupon.discount) {
      label = `${coupon.name} - 折扣${coupon.discount}`;
    }
    return label;
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
        user_coupon_id: selectedUserCouponId,
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

      <View style={{ marginTop: 8 }}>
        <Text>优惠券：{couponSummary}</Text>
        <Text style={{ fontSize: 12, color: '#999', marginTop: 2 }}>
          可点击下方按钮查看本单可用优惠券并手动选择
        </Text>
        <View style={{ marginTop: 4, display: 'flex', flexDirection: 'row', flexWrap: 'wrap', gap: 4 }}>
          {availableCoupons.length > 0 && (
            <Button
              size="mini"
              onClick={() => setShowCouponList((prev) => !prev)}
              type="primary"
            >
              {showCouponList
                ? '收起本单可用优惠券'
                : `本单可用优惠券（${availableCoupons.length}）`}
            </Button>
          )}
          {selectedUserCouponId ? (
            <Button
              size="mini"
              onClick={() => {
                setSelectedUserCouponId(undefined);
                setCouponSummary('暂不使用优惠券');
              }}
            >
              不使用优惠券
            </Button>
          ) : null}
        </View>

        {showCouponList && availableCoupons.length > 0 && (
          <View style={{ marginTop: 8 }}>
            {availableCoupons.map((uc) => {
              const label = buildCouponLabel(uc);
              const isSelected = uc.id === selectedUserCouponId;
              return (
                <Button
                  key={uc.id}
                  size="mini"
                  type={isSelected ? 'primary' : 'default'}
                  style={{ marginRight: 8, marginBottom: 4 }}
                  onClick={() => {
                    setSelectedUserCouponId(uc.id);
                    setCouponSummary(`已选择：${label}`);
                    setShowCouponList(false);
                  }}
                >
                  {label}
                </Button>
              );
            })}
          </View>
        )}
      </View>

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
