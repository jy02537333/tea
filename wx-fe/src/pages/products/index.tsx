import React, { useMemo, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { mockCart, mockProductCategories, mockProductItems } from '../../services/mockData';
import './index.scss';

export default function ProductsPage() {
  const [activeCategory, setActiveCategory] = useState(mockProductCategories[0]?.id ?? 'hot');
  const filteredItems = useMemo(
    () => mockProductItems.filter((item) => item.categoryId === activeCategory),
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
    <View className="page-products">
      <View className="app">
        <View className="app-header">
          <View className="app-header-top">
            <View>
              <Text className="store-name">隽也YUYE茶馆</Text>
              <Text className="store-meta">距离您 1671.6km · 商家暂无公告</Text>
            </View>
          </View>
          <Text className="sub-title">商城分区 · 全部商品</Text>
        </View>

        <View className="menu-wrap">
          <View className="category-list">
            {mockProductCategories.map((category) => (
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
                      <View className="btn-plus-circle"><Text>+</Text></View>
                    </View>
                  </View>
                </View>
              </View>
            ))}
          </View>
        </View>

        <View className="cart-bar">
          <View className="cart-left">
            <View className="cart-bag">
              <Text>👜</Text>
              <View className="cart-badge"><Text>{cartCount}</Text></View>
            </View>
            <Text className="cart-amount"><Text className="cart-amount__unit">¥</Text>{cartTotal}</Text>
          </View>
          <View className="cart-btn" onClick={goCheckout}><Text>去结算</Text></View>
        </View>
      </View>
    </View>
  );
}
