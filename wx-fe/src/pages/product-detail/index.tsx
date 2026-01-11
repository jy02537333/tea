import React, { useEffect, useState } from 'react';
import { View, Text, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getProduct } from '../../services/products';
import { addCartItem } from '../../services/cart';
import { Product, Store } from '../../services/types';
import { getStore } from '../../services/stores';
import './index.scss';

export default function ProductDetail() {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [currentStore, setCurrentStore] = useState<Store | null>(null);

  useEffect(() => {
    void fetchDetail();
  }, []);

  async function fetchDetail() {
    const router = Taro.getCurrentInstance().router;
    const idParam = router?.params?.id;
    const storeIdParam = router?.params?.store_id;
    if (!idParam) return;
    const id = Number(idParam);
    const store_id = storeIdParam ? Number(storeIdParam) : undefined;

    setLoading(true);
    try {
      if (store_id && Number.isFinite(store_id) && store_id > 0) {
        try { Taro.setStorageSync('current_store_id', String(store_id)); } catch (_) {}
      }
      const res = await getProduct(id, store_id);
      setProduct(res as Product);
      if (store_id) {
        try {
          const s = await getStore(store_id);
          setCurrentStore(s as Store);
        } catch (e) {
          // ignore store fetch errors
        }
      }
    } catch (e) {
      console.error('load product failed', e);
    } finally {
      setLoading(false);
    }
  }

  async function handleAddToCart() {
    if (!product) return;
    setSubmitting(true);
    try {
      await addCartItem(product.id, null, 1);
      Taro.showToast({ title: '已加入购物车', icon: 'success', duration: 1500 });
    } catch (e) {
      console.error('add to cart failed', e);
      Taro.showToast({ title: '加入失败', icon: 'none', duration: 1500 });
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) return <Text>加载中...</Text>;
  if (!product) return <Text>未找到商品</Text>;

  return (
    <View className="page-product-detail">
      {currentStore && (
        <View className="store-badge">
          <Text className="store-text">当前门店：{currentStore.name}</Text>
        </View>
      )}
      {product.images && (
        <Image className="cover" src={product.images} mode="aspectFill" />
      )}
      <View className="content">
        <Text className="title">{product.name}</Text>
        <View className="price-row">
          <Text className="price-now">¥ {typeof product.price === 'string' ? product.price : Number(product.price).toFixed(2)}</Text>
          {product.original_price && (
            <Text className="price-origin">¥ {product.original_price}</Text>
          )}
        </View>
      </View>

      <View className="action-bar">
        <Button className="btn-primary" disabled={submitting} onClick={handleAddToCart}>
          {submitting ? '提交中...' : '加入购物车'}
        </Button>
        <Button className="btn-secondary" onClick={() => Taro.navigateTo({ url: '/pages/cart/index' })}>去购物车</Button>
      </View>
    </View>
  );
}
