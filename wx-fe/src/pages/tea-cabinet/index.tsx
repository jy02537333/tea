import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function TeaCabinetPage() {
  return (
    <View className="page-tea-cabinet">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">我的茶柜</Text>
          </View>
          <Text className="nav-right">当前共 4 款寄存</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="summary-card">
            <View className="summary-top">
              <View>
                <Text className="summary-main">茶心君的专属茶柜</Text>
                <Text className="summary-sub">门店：茶心阁 · 茶心门店（茶心君坐店）</Text>
              </View>
              <Text className="summary-tag">到店出茶请提前备注</Text>
            </View>
            <Text className="summary-sub">本柜位暂存茶叶 4 款 · 茶器 1 套</Text>
          </View>

          <Text className="cabinet-section-title">寄存茶叶</Text>
          <View className="cabinet-card">
            <View className="cabinet-row">
              <View className="cabinet-main">
                <Text className="cabinet-name">桂花乌龙</Text>
                <Text className="cabinet-meta">剩余约 40g ｜ 预计可出杯 6-8 次</Text>
                <Text className="cabinet-tag-row">上次出茶：2 天前</Text>
              </View>
              <View className="cabinet-right">
                <Text>已开封</Text>
                <Text className="cabinet-status">可随时冲泡</Text>
              </View>
            </View>
            <View className="cabinet-row">
              <View className="cabinet-main">
                <Text className="cabinet-name">白桃乌龙</Text>
                <Text className="cabinet-meta">剩余约 30g ｜ 建议 1 个月内饮完</Text>
              </View>
              <View className="cabinet-right">
                <Text>未开封</Text>
              </View>
            </View>
          </View>

          <Text className="cabinet-section-title">寄存茶器</Text>
          <View className="cabinet-card">
            <View className="cabinet-row">
              <View className="cabinet-main">
                <Text className="cabinet-name">个人茶席茶具套装</Text>
                <Text className="cabinet-meta">盖碗 / 公道杯 / 品茗杯 4 只</Text>
              </View>
              <View className="cabinet-right">
                <Text>状态：完好</Text>
                <Text className="cabinet-status">仅供本人使用</Text>
              </View>
            </View>
          </View>

          <Text className="empty-tip">如需调整寄存内容或取回，请在到店后与茶心君确认～</Text>
        </ScrollView>

        <View className="footer">
          <Text>茶心阁 · 小程序「我的茶柜」页面静态稿</Text>
          <Text className="footer-sub">仅用于样式联调与布局对齐示意</Text>
        </View>
      </View>
    </View>
  );
}
