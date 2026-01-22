import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import { mockMenuCategories, mockMenuItems, mockMenuRecommendations, mockStores } from '../../services/mockData';
import './index.scss';

export default function MenuPage() {
  const [storeName, setStoreName] = useState('隽也YUYE茶馆');
  const [storeNameRaw, setStoreNameRaw] = useState('');
  const [activeCategory, setActiveCategory] = useState(mockMenuCategories[0]?.id ?? 'membership');

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
      if (saved) {
        setStoreNameRaw(saved);
        setStoreName(saved.replace(' · 本店', ''));
      }
    } catch (_) {}
  }

  useEffect(() => {
    syncStoreName();
  }, []);

  useDidShow(() => {
    syncStoreName();
  });

  const filteredItems = useMemo(
    () => mockMenuItems.filter((item) => item.categoryId === activeCategory),
    [activeCategory]
  );

  const storeDetailUrl = useMemo(() => {
    const matched = mockStores.find((store) => store.name === storeNameRaw)
      || mockStores.find((store) => store.name.includes(storeName));
    const params: string[] = [];
    if (matched?.id) params.push(`store_id=${matched.id}`);
    const name = storeNameRaw || storeName;
    if (name) params.push(`store_name=${encodeURIComponent(name)}`);
    return `/pages/store-detail/index${params.length ? `?${params.join('&')}` : ''}`;
  }, [storeName, storeNameRaw]);

  return (
    <View className="page-menu">
      <View className="app">
        <View className="app-header">
          <View className="app-header-top">
            <View className="header-info">
              <Text className="store-name">{storeName}</Text>
              <Text className="store-meta">距离您 1671.6km · 商家暂无公告</Text>
            </View>
            <View className="pickup-toggle">
              <View className="pickup-btn pickup-btn--active"><Text>自提</Text></View>
              <View className="pickup-btn"><Text>外卖</Text></View>
            </View>
          </View>
          <View className="store-distance-row">
            <View className="store-distance-main">
              <View className="store-distance-dot" />
              <Text>当前门店距离您约 <Text className="store-distance-strong">1671.6km</Text></Text>
            </View>
            <View className="store-distance-actions">
              <View className="store-switch-btn" onClick={() => Taro.navigateTo({ url: '/pages/stores/index' }).catch(() => {})}>
                <Text>切换门店</Text>
              </View>
              <View className="store-switch-btn store-detail-btn" onClick={() => Taro.navigateTo({ url: storeDetailUrl }).catch(() => {})}>
                <Text>门店详情</Text>
              </View>
            </View>
          </View>
          <View className="product-source-toggle">
            <View className="product-source-btn product-source-btn--active"><Text>门店商品</Text></View>
            <View className="product-source-btn" onClick={() => Taro.navigateTo({ url: '/pages/menu-platform/index?category=featured' }).catch(() => {})}>
              <Text>平台商品</Text>
            </View>
          </View>
        </View>

        <View className="recommend-section">
          <View className="section-title-row">
            <Text className="section-title">商家推荐</Text>
            <Text className="section-sub">为你精选的茶饮</Text>
          </View>
          <ScrollView className="recommend-row" scrollX showScrollbar={false} enhanced>
              {mockMenuRecommendations.map((item) => (
                <View className="recommend-card" key={item.id}>
                  <View className="recommend-img" />
                  <View className="recommend-body">
                    <Text className="recommend-name">{item.name}</Text>
                    <View className="recommend-price-row">
                      <Text className="price"><Text className="price-unit">¥</Text>{item.price}</Text>
                      <View className="btn-plus-circle"><Text>+</Text></View>
                    </View>
                  </View>
                </View>
              ))}
          </ScrollView>
        </View>

        <View className="menu-wrap">
          <View className="category-list">
            {mockMenuCategories.map((category) => (
              <Text
                key={category.id}
                className={`category-item ${activeCategory === category.id ? 'category-item--active' : ''}`}
                onClick={() => setActiveCategory(category.id)}
              >
                {category.name}
              </Text>
            ))}
          </View>
          <View className="goods-list">
            {filteredItems.map((item) => (
              <View className="goods-card" key={item.id}>
                <View className="goods-thumb" />
                <View className="goods-main">
                  <View>
                    <Text className="goods-name">{item.name}</Text>
                    <Text className="goods-desc">{item.desc}</Text>
                  </View>
                  <View className="goods-bottom">
                    <Text className="goods-tag">{item.tag}</Text>
                    <View className="goods-price-row">
                      <Text className="price"><Text className="price-unit">¥</Text>{item.price}</Text>
                      {item.useQtyControl ? (
                        <View className="goods-qty-row">
                          <View className="qty-btn"><Text>-</Text></View>
                          <View className="qty-value"><Text>1</Text></View>
                          <View className="qty-btn"><Text>+</Text></View>
                        </View>
                      ) : (
                        <View className="btn-plus-circle"><Text>+</Text></View>
                      )}
                    </View>
                  </View>
                  <View className="goods-action-row">
                    <View className="share-btn"><Text>分销</Text></View>
                  </View>
                </View>
              </View>
            ))}
          </View>
        </View>

        <View className="tab-bar">
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/home/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🏠</Text></View>
              <Text className="tab-bar__label">首页</Text>
            </View>
          </View>
          <View className="tab-bar__item tab-bar__item--active">
            <View className="tab-bar__link">
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
