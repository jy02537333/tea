import React, { useMemo } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { mockOrders, mockStores } from '../../services/mockData';
import './index.scss';

export default function OrderDetailPage() {
  const router = Taro.getCurrentInstance().router;
  const orderIdParam = router?.params?.id;

  const order = useMemo(() => {
    if (orderIdParam) {
      const found = mockOrders.find((item) => item.id === String(orderIdParam));
      if (found) return found;
    }
    return mockOrders[0];
  }, [orderIdParam]);

  const store = useMemo(
    () => mockStores.find((item) => item.id === order.storeId) || mockStores[0],
    [order.storeId]
  );

  const statusLabel = order.status === 'pending'
    ? '待支付'
    : order.status === 'done'
      ? '已完成'
      : order.status === 'canceled'
        ? '已取消'
        : '制作中';

  const steps = ['已下单', '制作中', '待取餐', '已完成'];
  const activeStepIndex = order.status === 'done'
    ? 3
    : order.status === 'making'
      ? 1
      : order.status === 'pending'
        ? 0
        : -1;

  const itemTotal = order.items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <View className="page-order-detail">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">订单详情</Text>
          </View>
          <Text className="nav-right">#{order.id}</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="status-card">
            <View className="status-row">
              <Text className="status-title">{statusLabel}</Text>
              <Text className={`status-tag status-tag--${order.status}`}>{order.pickupType}</Text>
            </View>
            {activeStepIndex >= 0 && (
              <View className="status-steps">
                {steps.map((step, index) => (
                  <View className="step" key={step}>
                    <View className={`step-dot ${index <= activeStepIndex ? 'step-dot--active' : ''}`} />
                    <Text className={`step-label ${index <= activeStepIndex ? 'step-label--active' : ''}`}>{step}</Text>
                  </View>
                ))}
              </View>
            )}
            {order.status === 'canceled' && (
              <Text className="status-cancel">该订单已取消，请重新下单。</Text>
            )}
          </View>

          <View className="info-card">
            <View className="info-header">
              <Text className="info-title">门店信息</Text>
              <Text className="info-sub">Store</Text>
            </View>
            <Text className="info-line">{order.storeName}</Text>
            <Text className="info-line">{store.address}</Text>
            <Text className="info-line">{store.phone}</Text>
          </View>

          <View className="info-card">
            <View className="info-header">
              <Text className="info-title">商品信息</Text>
              <Text className="info-sub">Items</Text>
            </View>
            {order.items.map((item) => (
              <View className="item-row" key={item.id}>
                <Text className="item-name">{item.name}</Text>
                <Text className="item-qty">x {item.quantity}</Text>
              </View>
            ))}
            <Text className="item-total">共 {itemTotal} 件商品</Text>
          </View>

          <View className="info-card">
            <View className="info-header">
              <Text className="info-title">订单信息</Text>
              <Text className="info-sub">Summary</Text>
            </View>
            <View className="summary-row">
              <Text>下单时间</Text>
              <Text>{order.time}</Text>
            </View>
            <View className="summary-row">
              <Text>订单编号</Text>
              <Text>{order.id}</Text>
            </View>
            <View className="summary-row">
              <Text>{order.totalLabel}</Text>
              <Text className="summary-amount">{typeof order.totalValue === 'number' ? `¥ ${order.totalValue}` : '-'}</Text>
            </View>
          </View>
        </ScrollView>
      </View>
    </View>
  );
}
