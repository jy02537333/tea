import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Thumbnail from '../../components/Thumbnail';
import Taro from '@tarojs/taro';
import { getProduct } from '../../services/products';
import { addCartItem } from '../../services/cart';
import type { Product } from '../../services/types';

export default function ProductDetail() {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const params = Taro.getCurrentInstance().router?.params || {} as any;
    const id = Number(params.id || 0);
    if (id) fetchDetail(id);
  }, []);

  async function fetchDetail(id: number) {
    setLoading(true);
    try {
      const p = await getProduct(id);
      setProduct(p);
    } finally {
      setLoading(false);
    }
  }

  async function addToCart() {
    if (!product) return;
    try {
      await addCartItem(product.id, 0, 1);
      Taro.showToast({ title: '已加入购物车', icon: 'success' });
    } catch (e) {
      Taro.showToast({ title: '加入失败', icon: 'error' });
    }
  }

  function normalizeImages(images: any): string[] {
    if (!images) return [];
    if (Array.isArray(images)) return images;
    try { const arr = JSON.parse(images); if (Array.isArray(arr)) return arr; } catch {}
    if (typeof images === 'string') return [images];
    return [];
  }

  return (
    <View style={{ padding: 12 }}>
      {!product && <Text>加载中...</Text>}
      {product && (
        <View>
          <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{product.name}</Text>
          <Text>价格: {product.price}</Text>
          <View style={{ flexDirection: 'row', marginTop: 6 }}>
            {normalizeImages(product.images).map((url, i) => (
              <Thumbnail key={i} src={url} width={80} height={80} radius={6} lazyLoad={true} />
            ))}
          </View>
          <Button onClick={addToCart} disabled={loading}>加入购物车</Button>
          <Button onClick={() => Taro.navigateTo({ url: '/src/pages/cart/index' })}>去购物车</Button>
        </View>
      )}
    </View>
  );
}
