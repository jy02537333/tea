import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import { mockCart, mockPlatformCategories, mockPlatformItems } from '../../services/mockData';
import './index.scss';

export default function MenuPlatformPage() {
  const [activeCategory, setActiveCategory] = useState(mockPlatformCategories[0]?.id ?? 'featured');
  const router = Taro.getCurrentInstance().router;

  function syncCategoryFromRoute() {
    const categoryParam = router?.params?.category;
    if (categoryParam && mockPlatformCategories.some((item) => item.id === String(categoryParam))) {
      setActiveCategory(String(categoryParam));
      try {
        Taro.setStorageSync('platform_category', String(categoryParam));
      } catch (_) {}
      return;
    }
    try {
      const cached = Taro.getStorageSync('platform_category');
      if (cached && mockPlatformCategories.some((item) => item.id === String(cached))) {
        setActiveCategory(String(cached));
      }
    } catch (_) {}
  }

  useEffect(() => {
    syncCategoryFromRoute();
  }, []);

  useDidShow(() => {
    syncCategoryFromRoute();
  });

  useEffect(() => {
    try {
      Taro.setStorageSync('platform_category', activeCategory);
    } catch (_) {}
  }, [activeCategory]);
  const filteredItems = useMemo(
    () => mockPlatformItems.filter((item) => {
      if (activeCategory === 'featured') return true;
      return item.categoryId === activeCategory;
    }),
    [activeCategory]
  );
  const cartCount = useMemo(
    () => mockCart.items.reduce((sum, item) => sum + item.quantity, 0),
    []
  );
  const cartTotal = useMemo(
    () => mockCart.items.reduce((sum, item) => sum + item.price * item.quantity, 0),
    []
  );

  function goCheckout() {
    const storeId = mockCart.storeId;
    const query = storeId ? `?store_id=${encodeURIComponent(storeId)}` : '';
    Taro.navigateTo({ url: `/pages/checkout/index${query}` }).catch(() => {});
  }

  return (
    <View className="page-menu-platform">
      <View className="app">
        <View className="app-header">
          <View className="app-header-top">
            <View>
              <Text className="store-name">隽也YUYE茶馆</Text>
              <Text className="store-meta">平台商品 · 由茶心阁直发</Text>
            </View>
          </View>
          <View className="product-source-toggle">
            <View className="product-source-btn" onClick={() => Taro.navigateTo({ url: '/pages/menu/index' }).catch(() => {})}>门店商品</View>
            <View className="product-source-btn product-source-btn--active">平台商品</View>
          </View>
        </View>

        <View className="menu-wrap">
          <View className="category-list">
            {mockPlatformCategories.map((category) => (
              <Text
                key={category.id}
                className={`category-item ${activeCategory === category.id ? 'category-item--active' : ''}`}
                onClick={() => setActiveCategory(category.id)}
              >
                {category.name}
              </Text>
            ))}
          </View>
          <ScrollView className="goods-list" scrollY enhanced>
            {filteredItems.map((item) => (
              <View
                className="goods-card"
                key={item.id}
                onClick={() => Taro.navigateTo({ url: `/pages/platform-product-detail/index?id=${item.id}&category=${activeCategory}` }).catch(() => {})}
              >
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
                      <View className="btn-plus-circle">+</View>
                    </View>
                  </View>
                  <View className="goods-action-row">
                    <View className="share-btn">分销</View>
                  </View>
                </View>
              </View>
            ))}
          </ScrollView>
        </View>

        <View className="cart-bar">
          <View className="cart-left">
            <View className="cart-bag">
              👜
              <View className="cart-badge">{cartCount}</View>
            </View>
            <Text className="cart-amount"><Text className="cart-amount__unit">¥</Text>{cartTotal}</Text>
          </View>
          <View className="cart-actions">
            <View className="cart-btn cart-btn--ghost" onClick={() => Taro.navigateTo({ url: '/pages/cart/index' }).catch(() => {})}>购物车</View>
            <View className="cart-btn" onClick={goCheckout}>去结算</View>
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
