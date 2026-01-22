import React, { useEffect, useState } from 'react';
import { View, Text, ScrollView, Image } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import './index.scss';
import top1 from '../../assets/home/top1.jpg';
import top2 from '../../assets/home/top2.jpg';

export default function HomePage() {
  const [storeName, setStoreName] = useState('茶心阁 · 茶心门店（茶心君坐店）');

  function syncStoreName() {
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
      if (saved) setStoreName(saved);
    } catch (_) {}
  }

  useEffect(() => {
    syncStoreName();
  }, []);

  useDidShow(() => {
    syncStoreName();
  });

  return (
    <View className="page-home">
      <View className="app">
        <View className="app-header">
          <View className="header-left">
            <View className="header-store">
              <View
                className="header-store-link"
                onClick={() => Taro.navigateTo({ url: '/pages/stores/index' }).catch(() => {})}
              >
                <View className="header-store-dot" />
                <Text className="header-store-name">{storeName}</Text>
                <Text className="header-store-arrow">﹀</Text>
              </View>
            </View>
            <Text className="header-sub">距离您约 1671.6km · 今日营业 10:00-22:00</Text>
          </View>
          <View className="header-right">
            <View className="header-avatar">
              <Text>云也</Text>
            </View>
          </View>
        </View>

        <View className="ad-carousel">
          <ScrollView className="ad-track" scrollX showScrollbar={false} enhanced>
            <View className="ad-item">
              <View className="ad-badge">广告</View>
              <Image className="ad-img" src={top1} mode="aspectFill" />
            </View>
            <View className="ad-item">
              <View className="ad-badge">广告</View>
              <Image className="ad-img" src={top2} mode="aspectFill" />
            </View>
          </ScrollView>
          <View className="ad-dots">
            <View className="ad-dot ad-dot--active" />
            <View className="ad-dot" />
          </View>
        </View>

        <View className="notice">
          <View className="notice-left">
            <View className="notice-dot"><Text>!</Text></View>
            <Text>温馨提示：本店一客一茶，如需拼台请备注说明～</Text>
          </View>
          <Text className="notice-arrow">›</Text>
        </View>

        <View className="entry-row">
          <View
            className="entry-card entry-card--tall"
            onClick={() => Taro.navigateTo({ url: '/pages/menu-platform/index?category=featured' }).catch(() => {})}
          >
            <View className="entry-icon"><Text>购</Text></View>
            <Text className="entry-title">平台商品</Text>
            <Text className="entry-sub">礼盒 · 茶器 · 年卡</Text>
            <View className="entry-footer">
              <Text className="entry-footer__label">进入平台商品</Text>
              <View className="entry-dot-list">
                <View className="entry-dot" />
                <View className="entry-dot" />
                <View className="entry-dot" />
              </View>
            </View>
          </View>
          <View className="entry-col">
            <View
              className="entry-card"
              onClick={() => Taro.navigateTo({ url: '/pages/today-tea/index' }).catch(() => {})}
            >
              <Text className="entry-title">今日茶单</Text>
              <Text className="entry-sub">当日上新 · 店主推荐</Text>
              <View className="entry-footer">
                <Text className="entry-footer__label">探索今日杯中之茶</Text>
                <View className="entry-dot-list">
                  <View className="entry-dot" />
                  <View className="entry-dot" />
                  <View className="entry-dot" />
                </View>
              </View>
            </View>
            <View
              className="entry-card"
              onClick={() => Taro.navigateTo({ url: '/pages/tea-reservation/index' }).catch(() => {})}
            >
              <Text className="entry-title">茶间预约</Text>
              <Text className="entry-sub">预约茶席 · 包场</Text>
              <View className="entry-footer">
                <Text className="entry-footer__label">选择时段与人数</Text>
                <View className="entry-dot-list">
                  <View className="entry-dot" />
                  <View className="entry-dot" />
                  <View className="entry-dot" />
                </View>
              </View>
            </View>
          </View>
        </View>

        <View className="section">
          <View className="section-header">
            <Text className="section-title">商城分区</Text>
            <Text className="section-link">全部商品 ›</Text>
          </View>
          <View className="partition-row">
            <View className="partition-card">
              <Text className="partition-tag">热销</Text>
              <Text className="partition-main">本月 Top · 茶饮</Text>
              <Text className="partition-extra">桂花乌龙 / 白桃乌龙</Text>
            </View>
            <View className="partition-card">
              <Text className="partition-tag">饮品</Text>
              <Text className="partition-main">现制冷萃 / 热泡</Text>
              <Text className="partition-extra">单杯 / 套餐</Text>
            </View>
          </View>
          <View className="partition-row">
            <View className="partition-card">
              <Text className="partition-tag">茶叶</Text>
              <Text className="partition-main">单品茶 · 礼盒</Text>
              <Text className="partition-extra">适合自饮与送礼</Text>
            </View>
            <View className="partition-card">
              <Text className="partition-tag">茶器</Text>
              <Text className="partition-main">盖碗 / 公道杯 / 茶杯</Text>
              <Text className="partition-extra">为好茶找到好器</Text>
            </View>
          </View>
        </View>

        <View className="section">
          <View className="section-header">
            <Text className="section-title">今日推荐</Text>
            <Text className="section-link">更多推荐 ›</Text>
          </View>

          <View className="product-row">
            <View className="product-info">
              <Text className="product-name">桂花乌龙 · 冷萃</Text>
              <Text className="product-meta">冷萃 500ml ｜ 桂花香气清甜，入口顺滑</Text>
            </View>
            <View>
              <Text className="product-price"><Text className="product-price__unit">¥</Text> 26</Text>
              <View className="product-btn">＋ 加入茶盘</View>
            </View>
          </View>

          <View className="product-row">
            <View className="product-info">
              <Text className="product-name">白桃乌龙 · 热泡</Text>
              <Text className="product-meta">热泡 400ml ｜ 果香饱满，适合慢饮</Text>
            </View>
            <View>
              <Text className="product-price"><Text className="product-price__unit">¥</Text> 28</Text>
              <View className="product-btn">＋ 加入茶盘</View>
            </View>
          </View>

          <View className="product-row">
            <View className="product-info">
              <Text className="product-name">岩茶·水仙</Text>
              <Text className="product-meta">茶叶 50g ｜ 岩韵显，适合回甘爱好者</Text>
            </View>
            <View>
              <Text className="product-price"><Text className="product-price__unit">¥</Text> 88</Text>
              <View className="product-btn">＋ 加入茶盘</View>
            </View>
          </View>
        </View>

        <View className="footer">
          <Text>茶心阁 · 小程序首页静态稿</Text>
          <Text className="footer-sub">仅用于样式联调与布局对齐示意</Text>
        </View>

        <View className="tab-bar">
          <View className="tab-bar__item tab-bar__item--active">
            <View className="tab-bar__link">
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
