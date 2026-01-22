import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import './index.scss';

export default function RechargePage() {
  return (
    <View className="page-recharge">
      <View className="app">
        <View className="nav-bar">
          <Text className="nav-title">充值中心</Text>
          <Text className="nav-right">茶心阁 · 充值福利</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="balance-card">
            <View className="balance-top">
              <View>
                <Text className="balance-label">当前余额</Text>
                <Text className="balance-value">¥ 120.00</Text>
                <Text className="balance-sub">充值到账后可用于下单与茶席消费</Text>
              </View>
              <Text className="balance-badge">会员权益加成</Text>
            </View>
          </View>

          <Text className="plan-section-title">选择充值档位</Text>
          <View className="plan-grid">
            <View className="plan-card plan-hot">
              <Text className="plan-hot-label">热门</Text>
              <Text className="plan-main">¥ 100</Text>
              <Text className="plan-tag">到账 ¥ 110</Text>
              <Text className="plan-gift">赠茶币 10</Text>
            </View>
            <View className="plan-card">
              <Text className="plan-main">¥ 200</Text>
              <Text className="plan-tag">到账 ¥ 220</Text>
              <Text className="plan-gift">赠茶币 20</Text>
            </View>
            <View className="plan-card">
              <Text className="plan-main">¥ 500</Text>
              <Text className="plan-tag">到账 ¥ 560</Text>
              <Text className="plan-gift">赠茶币 60</Text>
            </View>
            <View className="plan-card">
              <Text className="plan-main">¥ 1000</Text>
              <Text className="plan-tag">到账 ¥ 1150</Text>
              <Text className="plan-gift">赠茶币 150</Text>
            </View>
          </View>

          <View className="method-card">
            <View className="method-row">
              <View className="method-left">
                <View className="method-icon">💳</View>
                <View>
                  <Text className="method-name">微信支付</Text>
                  <Text className="method-desc">推荐 · 到账更快</Text>
                </View>
              </View>
              <Text className="method-right">已选</Text>
            </View>
            <View className="method-row">
              <View className="method-left">
                <View className="method-icon">🏦</View>
                <View>
                  <Text className="method-name">银行卡</Text>
                  <Text className="method-desc">支持储蓄卡</Text>
                </View>
              </View>
              <Text className="method-right">可用</Text>
            </View>
          </View>
        </ScrollView>

        <View className="bottom-bar">
          <View className="bottom-info">
            <Text className="bottom-main">已选 ¥100 · <Text className="bottom-highlight">到账 ¥110</Text></Text>
            <Text className="bottom-main">含赠送茶币 10</Text>
          </View>
          <View className="bottom-btn">确认充值</View>
        </View>
      </View>
    </View>
  );
}
