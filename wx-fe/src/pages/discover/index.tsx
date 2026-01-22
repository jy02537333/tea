import React, { useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function DiscoverPage() {
  const [activeTab, setActiveTab] = useState<'activity' | 'news'>('activity');

  return (
    <View className="page-discover">
      <View className="app">
        <View className="nav-bar">
          <Text className="nav-title">发现</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="tabs" role="tablist">
            <View
              className={`tab ${activeTab === 'activity' ? 'tab--active' : ''}`}
              role="tab"
              aria-selected={activeTab === 'activity'}
              onClick={() => setActiveTab('activity')}
            >
              <Text>活动</Text>
            </View>
            <View
              className={`tab ${activeTab === 'news' ? 'tab--active' : ''}`}
              role="tab"
              aria-selected={activeTab === 'news'}
              onClick={() => setActiveTab('news')}
            >
              <Text>资讯</Text>
            </View>
          </View>

          <View className={`tab-panel ${activeTab === 'activity' ? '' : 'tab-panel--hidden'}`} data-panel="activity">
            <Text className="section-title">活动</Text>
            <View className="card">
              <Text className="card-title">茶心阁 · 春日品鉴会</Text>
              <Text className="card-meta">本周六 14:00 · 杭州西湖店 · 报名中</Text>
              <View className="tag-row">
                <Text className="tag">线下活动</Text>
                <Text className="tag">需报名</Text>
              </View>
            </View>
            <View className="card">
              <Text className="card-title">会员日专享 · 买一送一</Text>
              <Text className="card-meta">每月 18 日 · 指定茶品</Text>
              <View className="tag-row">
                <Text className="tag">会员专属</Text>
                <Text className="tag">进行中</Text>
              </View>
            </View>
          </View>

          <View className={`tab-panel ${activeTab === 'news' ? '' : 'tab-panel--hidden'}`} data-panel="news">
            <Text className="section-title">资讯</Text>
            <View className="card">
              <Text className="card-title">茶山新季到货 · 桂花乌龙上新</Text>
              <Text className="card-meta">今日上新 · 了解产地与风味故事</Text>
            </View>
            <View className="card">
              <Text className="card-title">茶席礼仪小课堂</Text>
              <Text className="card-meta">本周精选 · 一客一茶的礼仪与体验</Text>
            </View>
          </View>
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
          <View className="tab-bar__item tab-bar__item--active">
            <View className="tab-bar__link">
              <View className="tab-bar__icon"><Text>🧭</Text></View>
              <Text className="tab-bar__label">发现</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {})}>
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
