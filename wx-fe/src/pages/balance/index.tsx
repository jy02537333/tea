import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function BalancePage() {
  return (
    <View className="page-balance">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">账户余额</Text>
          </View>
          <Text className="nav-right">余额明细</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="balance-card">
            <Text className="balance-title">当前余额</Text>
            <Text className="balance-amount">¥ 120.00</Text>
            <Text className="balance-meta">当前 VIP 等级权限：黄金VIP · 可免费提现 1 次/月</Text>
            <View className="action-row">
              <View className="btn">查看明细</View>
              <View className="btn btn-primary" onClick={() => Taro.navigateTo({ url: '/pages/withdraw/index' }).catch(() => {})}>提现</View>
            </View>
          </View>

          <View className="list-card">
            <View className="list-row">
              <Text className="list-label">本月可免费提现次数</Text>
              <Text className="list-right">剩余 1 次</Text>
            </View>
            <View className="list-row">
              <Text className="list-label">提现规则</Text>
              <Text className="list-right">查看 ›</Text>
            </View>
          </View>
        </ScrollView>

        <View className="footer">茶心阁 · 小程序「账户余额」页面静态稿</View>
      </View>
    </View>
  );
}
