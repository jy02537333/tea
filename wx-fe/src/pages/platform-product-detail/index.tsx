import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { mockPlatformCategories, mockPlatformProducts } from '../../services/mockData';
import './index.scss';

export default function PlatformProductDetailPage() {
  const router = Taro.getCurrentInstance().router;
  const productIdParam = router?.params?.id;
  const categoryParam = router?.params?.category;
  const product = useMemo(() => {
    if (productIdParam) {
      const found = mockPlatformProducts.find((item) => item.id === String(productIdParam));
      if (found) return found;
    }
    return mockPlatformProducts[0];
  }, [productIdParam]);

  const [selectedSkuId, setSelectedSkuId] = useState(product.defaultSkuId);

  const categoryLabel = useMemo(() => {
    if (!categoryParam) return '';
    const matched = mockPlatformCategories.find((item) => item.id === String(categoryParam));
    return matched?.name ?? '';
  }, [categoryParam]);

  useEffect(() => {
    setSelectedSkuId(product.defaultSkuId);
  }, [product.defaultSkuId]);
  const selectedSku = useMemo(
    () => product.skus.find((sku) => sku.id === selectedSkuId) ?? product.skus[0],
    [product.skus, selectedSkuId]
  );

  const statusText = selectedSku.status === 'available'
    ? '可购买'
    : selectedSku.status === 'presale'
      ? '预售中'
      : '已售罄';
  const statusClass = selectedSku.status === 'soldout' ? 'sku-status sku-status--soldout' : 'sku-status';
  const fallbackListUrl = categoryParam
    ? `/pages/menu-platform/index?category=${encodeURIComponent(String(categoryParam))}`
    : '/pages/menu-platform/index';

  return (
    <View className="page-platform-product-detail">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View
              className="nav-back"
              onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {
                Taro.navigateTo({ url: fallbackListUrl }).catch(() => {});
              })}
            >
              ‹
            </View>
            <Text className="nav-title">商品详情</Text>
          </View>
          <Text className="nav-right">平台发货</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="hero-img-wrap">
            <View className="hero-img" />
            <Text className="hero-tag">{product.heroTag}</Text>
            <Text className="hero-badge">{product.heroBadge}</Text>
          </View>

          <View className="section-card">
            <View className="title-row">
              <View className="title-main">
                <Text className="prod-name">{product.name}</Text>
                <View className="prod-tags">
                  {categoryLabel && (
                    <Text className="prod-tag-pill prod-tag-pill--active">{categoryLabel}</Text>
                  )}
                  {product.tags.map((tag) => (
                    <Text className="prod-tag-pill" key={tag}>{tag}</Text>
                  ))}
                  <Text className={statusClass}>{statusText}</Text>
                </View>
              </View>
              <View className="price-col">
                <Text className="price-main"><Text className="price-unit">¥</Text>{selectedSku.price}</Text>
                <Text className="price-sub">约可获积分 {selectedSku.points}</Text>
              </View>
            </View>
            <Text className="desc-text">{product.desc}</Text>
            <View className="sku-row">
              <Text className="sku-label">规格选择</Text>
              <View className="sku-list">
                {product.skus.map((sku) => (
                  <Text
                    key={sku.id}
                    className={`sku-pill ${selectedSkuId === sku.id ? 'sku-pill--active' : ''} ${sku.status === 'soldout' ? 'sku-pill--disabled' : ''}`}
                    onClick={() => setSelectedSkuId(sku.id)}
                  >
                    {sku.name}
                  </Text>
                ))}
              </View>
              <Text className="sku-meta">库存：{typeof selectedSku.stock === 'number' ? selectedSku.stock : '-'} · 状态：{statusText}</Text>
            </View>
            <Text className="meta-line"><Text className="meta-label">净含量</Text>{product.netWeight}</Text>
            <Text className="meta-line"><Text className="meta-label">保质期</Text>{product.shelfLife}</Text>
          </View>

          <View className="section-card">
            <View className="section-title-row">
              <Text className="section-title">发货与配送</Text>
              <Text className="section-sub">Shipping</Text>
            </View>
            <View className="bullet-list">
              {product.shippingTips.map((tip) => (
                <Text key={tip}>• {tip}</Text>
              ))}
            </View>
          </View>

          <View className="section-card">
            <View className="section-title-row">
              <Text className="section-title">冲泡建议</Text>
              <Text className="section-sub">Brewing Tips</Text>
            </View>
            <View className="bullet-list">
              {product.brewTips.map((tip) => (
                <Text key={tip}>• {tip}</Text>
              ))}
            </View>
          </View>

          <View className="section-card">
            <View className="section-title-row">
              <Text className="section-title">售后说明</Text>
              <Text className="section-sub">After Sales</Text>
            </View>
            <View className="bullet-list">
              {product.afterSales.map((tip) => (
                <Text key={tip}>• {tip}</Text>
              ))}
            </View>
          </View>
        </ScrollView>

        <View className="bottom-bar">
          <View className="bottom-left">
            <Text>当前选择：<Text className="bottom-strong">{selectedSku.name}</Text></Text>
            <Text>合计金额：<Text className="bottom-strong">¥{selectedSku.price}</Text></Text>
          </View>
          <View className={`btn-primary ${selectedSku.status === 'soldout' ? 'btn-primary--disabled' : ''}`}>
            {selectedSku.status === 'soldout' ? '已售罄' : selectedSku.status === 'presale' ? '预售下单' : '加入购物车'}
          </View>
        </View>
      </View>
    </View>
  );
}
