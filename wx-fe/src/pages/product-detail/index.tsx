import React, { useEffect, useState } from 'react';
import { View, Text, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getProduct } from '../../services/products';
import { addCartItem } from '../../services/cart';
import { Product } from '../../services/types';

export default function ProductDetail() {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

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
      const res = await getProduct(id, store_id);
      setProduct(res as Product);
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
    <View style={{ padding: 12 }}>
      {product.images && (
        <Image
          src={product.images}
          mode="aspectFill"
          style={{ width: '100%', height: 200, marginBottom: 12 }}
        />
      )}
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{product.name}</Text>
      <View style={{ marginTop: 8, marginBottom: 12 }}>
        <Text>价格: {product.price}</Text>
        {product.original_price && (
          <Text> 原价: {product.original_price}</Text>
        )}
      </View>
      <Button disabled={submitting} onClick={handleAddToCart}>
        {submitting ? '提交中...' : '加入购物车'}
      </Button>
    </View>
  );
}
