import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function WithdrawPage() {
  return (
    <View className="page-withdraw">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">提现</Text>
          </View>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="form-card">
            <View className="form-row">
              <Text className="form-label">提现方式</Text>
              <Text className="form-value">银行卡</Text>
            </View>
            <View className="form-row">
              <Text className="form-label">到账银行卡</Text>
              <Text className="form-value">招商银行（尾号 8899）</Text>
            </View>
            <View className="form-row">
              <Text className="form-label">可提现余额</Text>
              <Text className="form-value">¥ 120.00</Text>
            </View>
            <View>
              <Text className="form-label">提现金额</Text>
              <Text className="amount">¥ 100.00</Text>
            </View>
            <View className="btn-primary">确认提现</View>
            <Text className="hint">提现预计 1-3 个工作日到账，具体到账时间以银行为准。</Text>
          </View>
        </ScrollView>

        <View className="footer">茶心阁 · 小程序「提现」页面静态稿</View>
      </View>
    </View>
  );
}
