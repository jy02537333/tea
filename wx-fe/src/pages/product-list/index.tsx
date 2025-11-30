import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Thumbnail from '../../components/Thumbnail';
import { listProducts } from '../../services/products';
import { Product } from '../../services/types';

export default function ProductList() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetch();
  }, []);

  async function fetch() {
    setLoading(true);
    try {
      const res = await listProducts({ page: 1, limit: 20 });
      const maybe = res as any;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setProducts(items);
    } finally {
      setLoading(false);
    }
  }

  // 兼容 images 为 string[] 或 string
  function normalizeImages(images: any): string[] {
    if (!images) return [];
    if (Array.isArray(images)) return images;
    try {
      const arr = JSON.parse(images);
      if (Array.isArray(arr)) return arr;
    } catch {}
    if (typeof images === 'string') return [images];
    return [];
  }

  return (
    <View style={{ padding: 12 }}>
      {products.map((p) => (
        <View key={p.id} style={{ marginBottom: 12, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 8 }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{p.name}</Text>
          <Text>价格: {p.price}</Text>
          {/* 多图渲染 */}
          <View style={{ flexDirection: 'row', marginTop: 4 }}>
            {normalizeImages(p.images).map((url, i) => (
              <Thumbnail key={i} src={url} width={60} height={60} radius={4} lazyLoad={true} />
            ))}
          </View>
          <Button onClick={() => console.log('add', p.id)}>加入购物车</Button>
        </View>
      ))}
    </View>
  );
}
