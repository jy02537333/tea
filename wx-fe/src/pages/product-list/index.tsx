import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
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

  return (
    <View style={{ padding: 12 }}>
      {products.map((p) => (
        <View key={p.id} style={{ marginBottom: 12, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 8 }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{p.name}</Text>
          <Text>价格: {p.price}</Text>
          <Button onClick={() => console.log('add', p.id)}>加入购物车</Button>
        </View>
      ))}
    </View>
  );
}
