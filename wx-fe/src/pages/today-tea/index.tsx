import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function TodayTeaPage() {
  return (
    <View className="page-today-tea">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">今日茶单</Text>
          </View>
          <Text className="nav-right">茶心阁 · 茶心门店</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="date-row">
            <View className="date-main">
              <Text className="tag-today">今日 · 周六</Text>
              <Text>桂花乌龙 / 白桃乌龙 / 岩茶水仙</Text>
            </View>
            <Text>一客一茶 · 现场现泡</Text>
          </View>

          <Text className="section-title">冷萃茶饮</Text>
          <View className="group-card">
            <View className="tea-row">
              <View className="tea-main">
                <Text className="tea-name">桂花乌龙 · 冷萃</Text>
                <Text className="tea-meta">冷萃 500ml ｜ 桂花香气清甜，适合夏日慢饮</Text>
                <Text className="tea-tag-row">招牌 · 推荐</Text>
              </View>
              <View className="tea-right">
                <Text className="tea-price">¥ 26</Text>
                <Text className="tea-hot">今日热销</Text>
              </View>
            </View>
            <View className="tea-row">
              <View className="tea-main">
                <Text className="tea-name">白桃乌龙 · 冷萃</Text>
                <Text className="tea-meta">冷萃 500ml ｜ 果香饱满 · 轻盈</Text>
              </View>
              <View className="tea-right">
                <Text className="tea-price">¥ 28</Text>
              </View>
            </View>
          </View>

          <Text className="section-title">热泡茶饮</Text>
          <View className="group-card">
            <View className="tea-row">
              <View className="tea-main">
                <Text className="tea-name">白桃乌龙 · 热泡</Text>
                <Text className="tea-meta">热泡 400ml ｜ 适合夜晚慢饮</Text>
              </View>
              <View className="tea-right">
                <Text className="tea-price">¥ 28</Text>
              </View>
            </View>
            <View className="tea-row">
              <View className="tea-main">
                <Text className="tea-name">岩茶 · 水仙</Text>
                <Text className="tea-meta">单品茶 ｜ 岩韵显，回甘持久</Text>
              </View>
              <View className="tea-right">
                <Text className="tea-price">¥ 88</Text>
              </View>
            </View>
          </View>

          <Text className="section-title">今日小食</Text>
          <View className="group-card">
            <View className="tea-row">
              <View className="tea-main">
                <Text className="tea-name">云也小点心</Text>
                <Text className="tea-meta">搭配任何茶饮皆宜</Text>
              </View>
              <View className="tea-right">
                <Text className="tea-price">¥ 30</Text>
              </View>
            </View>
          </View>
        </ScrollView>

        <View className="footer">
          <Text>茶心阁 · 小程序「今日茶单」页面静态稿</Text>
          <Text className="footer-sub">仅用于样式联调与布局对齐示意</Text>
        </View>
      </View>
    </View>
  );
}
