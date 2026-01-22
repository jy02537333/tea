import React, { useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { CartItem, mockCart, mockStores } from '../../services/mockData';
import './index.scss';

type CartSelection = {
  storeId?: string;
  storeName?: string;
  items: CartItem[];
  groups?: Array<{ storeId: string; storeName: string; items: CartItem[] }>;
};

export default function CheckoutPage() {
  const router = Taro.getCurrentInstance().router;
  const storeIdParam = router?.params?.store_id;
  const [currentStoreId] = useState(() => {
    if (storeIdParam) return String(storeIdParam);
    try {
      const cached = Taro.getStorageSync('current_store_id');
      if (cached) return String(cached);
    } catch (_) {}
    return mockCart.storeId;
  });

  const selection = useMemo<CartSelection>(() => {
    try {
      const raw = Taro.getStorageSync('cart_selection');
      if (!raw) return { items: [] };
      if (typeof raw === 'string') {
        const parsed = JSON.parse(raw);
        if (parsed && Array.isArray(parsed.items)) {
          return {
            storeId: parsed.storeId,
            storeName: parsed.storeName,
            items: parsed.items as CartItem[],
            groups: Array.isArray(parsed.groups) ? parsed.groups : undefined
          };
        }
      }
      if (raw && typeof raw === 'object' && Array.isArray((raw as any).items)) {
        return {
          storeId: (raw as any).storeId,
          storeName: (raw as any).storeName,
          items: (raw as any).items as CartItem[],
          groups: Array.isArray((raw as any).groups) ? (raw as any).groups : undefined
        };
      }
    } catch (_) {}
    return { items: [] };
  }, []);

  const checkoutItems = selection.items.length > 0 ? selection.items : mockCart.items;
  const checkoutStoreId = selection.storeId || currentStoreId || mockCart.storeId;
  const checkoutGroups = useMemo(() => {
    if (selection.groups && selection.groups.length > 0) return selection.groups;
    return [{ storeId: checkoutStoreId, storeName: selection.storeName || mockCart.storeName, items: checkoutItems }];
  }, [selection.groups, checkoutItems, checkoutStoreId, selection.storeName]);

  const store = useMemo(
    () => mockStores.find((item) => item.id === checkoutStoreId) || mockStores[0],
    [checkoutStoreId]
  );

  const totals = useMemo(() => {
    let count = 0;
    let subtotal = 0;
    let discount = 0;
    let serviceFee = 0;
    checkoutGroups.forEach((group) => {
      const groupSubtotal = group.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
      const groupCount = group.items.reduce((sum, item) => sum + item.quantity, 0);
      const rule = (mockStores.find((item) => item.id === group.storeId) || store).feeRule;
      const groupDiscount = groupSubtotal >= rule.couponThreshold ? rule.couponDiscount : 0;
      count += groupCount;
      subtotal += groupSubtotal;
      discount += groupDiscount;
      serviceFee += rule.serviceFee;
    });
    const total = subtotal + serviceFee - discount;
    return { count, subtotal, discount, serviceFee, total };
  }, [checkoutGroups]);

  return (
    <View className="page-checkout">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back">‹</View>
            <Text className="nav-title">提交订单</Text>
          </View>
          <Text className="nav-right">{store.name}</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="card">
            <View className="card-header">
              <Text className="card-header-title">收货信息</Text>
              <Text className="card-header-extra">选择地址 ›</Text>
            </View>
            <View className="address-main">
              <Text className="address-tag">{mockCart.pickupType === 'self' ? '到店自取' : '店内享用'}</Text>
              <View>
                <Text className="address-line">{selection.storeName || store.name}</Text>
                <Text className="address-phone">{store.address} ｜ {store.phone}</Text>
              </View>
            </View>
          </View>

          {checkoutGroups.map((group) => {
            const groupSubtotal = group.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
            const groupCount = group.items.reduce((sum, item) => sum + item.quantity, 0);
            const groupStore = mockStores.find((item) => item.id === group.storeId) || store;
            const rule = groupStore.feeRule;
            const groupDiscount = groupSubtotal >= rule.couponThreshold ? rule.couponDiscount : 0;
            return (
              <View key={group.storeId}>
                <View className="card">
                  <View className="card-header">
                    <Text className="card-header-title">商品信息</Text>
                    <Text className="card-header-extra">共 {groupCount} 件</Text>
                  </View>
                  <View className="group-store-line">
                    <Text className="group-store-name">{group.storeName}</Text>
                    <Text className="group-store-meta">{groupStore.address}</Text>
                  </View>
                  {group.items.map((item) => (
                    <View className="item-row" key={item.id}>
                      <View className="item-main">
                        <Text className="item-name">{item.name}</Text>
                        <Text className="item-meta">{item.spec}</Text>
                      </View>
                      <View className="item-ops">
                        <Text className="item-price">¥ {item.price}</Text>
                        <Text className="item-count">x {item.quantity}</Text>
                      </View>
                    </View>
                  ))}
                </View>

                <View className="card">
                  <View className="card-header">
                    <Text className="card-header-title">优惠与活动</Text>
                    <Text className="card-header-extra">去选择优惠券 ›</Text>
                  </View>
                  <View className="fee-row">
                    <View>
                      <Text>商品小计</Text>
                    </View>
                    <Text className="fee-value">¥ {groupSubtotal}</Text>
                  </View>
                  <View className="fee-row">
                    <View>
                      <Text>优惠券</Text>
                      <Text className="fee-label-sub">{rule.couponLabel}</Text>
                    </View>
                    <Text className="fee-value-highlight">- ¥ {groupDiscount}</Text>
                  </View>
                  <View className="fee-row">
                    <Text>茶席服务费</Text>
                    <Text className="fee-value">¥ {rule.serviceFee}</Text>
                  </View>
                </View>
              </View>
            );
          })}

          <View className="card">
            <View className="card-header">
              <Text className="card-header-title">备注</Text>
              <Text className="card-header-extra">选填</Text>
            </View>
            <View className="remark-input">例如：不放蔗糖 / 少一点桂花 / 提前 10 分钟出杯等…</View>
          </View>
        </ScrollView>

        <View className="bottom-bar">
          <View className="bottom-info">
            <Text className="bottom-total">
              <Text className="bottom-total__label">合计：</Text>¥ {totals.total}
            </Text>
            <Text className="bottom-sub">已优惠 ¥ {totals.discount} ｜ 共 {totals.count} 件商品</Text>
          </View>
          <View className="bottom-btn">微信支付</View>
        </View>
      </View>
    </View>
  );
}
