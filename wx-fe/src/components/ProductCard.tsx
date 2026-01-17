import React from 'react';
import { View, Text, Image } from '@tarojs/components';
import type { Product } from '../services/types';
import './ProductCard.scss';

export default function ProductCard(props: {
  product: Product;
  showCover?: boolean;
  onClick?: () => void;
  meta?: React.ReactNode;
  extra?: React.ReactNode;
}) {
  const { product, showCover = true, onClick, meta, extra } = props;

  function coverUrl(p: Product): string {
    const raw = (p as any).image_url || (p as any).cover || p.images || '';
    const first = String(raw).split(',').map((s) => s.trim()).filter(Boolean)[0];
    return first || 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
  }

  return (
    <View className="card-product" onClick={onClick}>
      {showCover && (
        <Image src={coverUrl(product)} mode="aspectFill" className="cover-sm" />
      )}
      <Text className="name">{product.name}</Text>
      <Text className="price">¥ {typeof product.price === 'string' ? product.price : Number(product.price).toFixed(2)}</Text>
      {meta && (
        <View className="meta" onClick={(e) => { e.stopPropagation?.(); }}>
          {meta}
        </View>
      )}
      {extra && (
        <View className="actions" onClick={(e) => { e.stopPropagation?.(); }}>
          {extra}
        </View>
      )}
    </View>
  );
}
