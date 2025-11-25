import React from 'react';
import { View, Text, Button } from '@tarojs/components';
import { Product } from '../services/types';

export default function ProductCard({ product, onAdd }: { product: Product; onAdd?: (id: number) => void }) {
  return (
    <View style={{ borderWidth: 1, borderColor: '#eee', padding: 8, borderRadius: 4 }}>
      <Text style={{ fontSize: 16 }}>{product.name}</Text>
      <Text>价格: {product.price}</Text>
      <Button onClick={() => onAdd && onAdd(product.id)}>加入购物车</Button>
    </View>
  );
}
