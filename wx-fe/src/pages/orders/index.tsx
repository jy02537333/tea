import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { mockOrders } from '../../services/mockData';
import './index.scss';

export default function OrdersPage() {
  const [activeTab, setActiveTab] = useState<'all' | 'ongoing' | 'done' | 'canceled'>('all');
  const [currentStoreName, setCurrentStoreName] = useState('');

  useEffect(() => {
    try {
      const raw = Taro.getStorageSync('current_store_name');
      let saved = '';
      if (typeof raw === 'string') {
        saved = raw;
        if (raw.trim().startsWith('{')) {
          try {
            const parsed = JSON.parse(raw);
            if (parsed && typeof parsed.data === 'string') saved = parsed.data;
          } catch (_) {}
        }
      } else if (raw && typeof raw === 'object' && typeof (raw as any).data === 'string') {
        saved = String((raw as any).data);
      }
      saved = String(saved || '').trim();
      if (saved) setCurrentStoreName(saved.replace(' · 本店', ''));
    } catch (_) {}
  }, []);

  const filteredOrders = useMemo(() => {
    if (activeTab === 'all') return mockOrders;
    if (activeTab === 'ongoing') return mockOrders.filter((order) => order.status === 'making' || order.status === 'pending');
    if (activeTab === 'done') return mockOrders.filter((order) => order.status === 'done');
    return mockOrders.filter((order) => order.status === 'canceled');
  }, [activeTab]);

  function renderStatus(orderStatus: string) {
    if (orderStatus === 'pending') return { label: '待支付', className: 'order-status order-status--pending' };
    if (orderStatus === 'done') return { label: '已完成', className: 'order-status order-status--done' };
    if (orderStatus === 'canceled') return { label: '已取消', className: 'order-status order-status--canceled' };
    return { label: '制作中', className: 'order-status' };
  }

  function goOrderDetail(id: string) {
    Taro.navigateTo({ url: `/pages/order-detail/index?id=${encodeURIComponent(id)}` }).catch(() => {});
  }

  return (
    <View className="page-orders">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back">‹</View>
            <Text className="nav-title">订单列表</Text>
          </View>
          <Text className="nav-right">茶心君 · 我的订单</Text>
        </View>

        <View className="tabs">
          <View className="tabs-left">
            <View className={`tab ${activeTab === 'all' ? 'tab--active' : ''}`} onClick={() => setActiveTab('all')}>
              <Text>全部</Text>
            </View>
            <View className={`tab ${activeTab === 'ongoing' ? 'tab--active' : ''}`} onClick={() => setActiveTab('ongoing')}>
              <Text>进行中</Text>
            </View>
            <View className={`tab ${activeTab === 'done' ? 'tab--active' : ''}`} onClick={() => setActiveTab('done')}>
              <Text>已完成</Text>
            </View>
            <View className={`tab ${activeTab === 'canceled' ? 'tab--active' : ''}`} onClick={() => setActiveTab('canceled')}>
              <Text>已取消</Text>
            </View>
          </View>
          <Text className="tabs-right">近三个月</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          {filteredOrders.map((order) => {
            const status = renderStatus(order.status);
            const itemsText = order.items.map((item) => `${item.name} x${item.quantity}`).join(' ｜ ');
            const isCurrentStore = currentStoreName && order.storeName.includes(currentStoreName);
            return (
              <View className="order-card" key={order.id}>
                <View className="order-header">
                  <View className="order-store-row">
                    <Text className="order-store">{order.storeName}</Text>
                    {isCurrentStore && <Text className="order-store-badge">当前门店</Text>}
                  </View>
                  <Text className={status.className}>{status.label}</Text>
                </View>
                <View className="order-meta-row">
                  <Text>{order.time} · {order.pickupType}</Text>
                  <Text># {order.id}</Text>
                </View>
                <Text className="order-items">{itemsText}</Text>
                <View className="order-total-row">
                  <View>
                    <Text className="order-total-label">{order.totalLabel}</Text>
                    {typeof order.totalValue === 'number' && (
                      <Text className="order-total-value">¥ {order.totalValue}</Text>
                    )}
                  </View>
                  <View className="order-actions">
                    {order.actions.map((action) => (
                      <View
                        key={action}
                        className={action === '催一催' || action === '去支付' ? 'btn-primary' : 'btn-ghost'}
                        onClick={() => {
                          if (action === '订单详情') goOrderDetail(order.id);
                        }}
                      >
                        {action}
                      </View>
                    ))}
                  </View>
                </View>
              </View>
            );
          })}

          {filteredOrders.length === 0 && (
            <View className="empty-tip">
              <Text className="empty-main">没有更多订单啦～</Text>
              <Text>可前往首页继续下单体验新茶款</Text>
            </View>
          )}
        </ScrollView>

        <View className="tab-bar">
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/home/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🏠</Text></View>
              <Text className="tab-bar__label">首页</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/menu/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🍵</Text></View>
              <Text className="tab-bar__label">点单</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/discover/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🧭</Text></View>
              <Text className="tab-bar__label">发现</Text>
            </View>
          </View>
          <View className="tab-bar__item tab-bar__item--active">
            <View className="tab-bar__link">
              <View className="tab-bar__icon"><Text>📄</Text></View>
              <Text className="tab-bar__label">订单</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/profile/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>👤</Text></View>
              <Text className="tab-bar__label">我的</Text>
            </View>
          </View>
        </View>
      </View>
    </View>
  );
}
