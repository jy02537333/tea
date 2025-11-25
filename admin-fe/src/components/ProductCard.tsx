import React from 'react';
import { Card, Button } from 'antd';
import { Product } from '../services/types';

export default function ProductCard({ product, onAdd }: { product: Product; onAdd?: (id: number) => void }) {
  return (
    <Card size="small" title={product.name} style={{ width: 240 }}>
      <div>价格: {product.price}</div>
      <div>库存: {product.stock}</div>
      <Button type="primary" onClick={() => onAdd && onAdd(product.id)} style={{ marginTop: 8 }}>
        加入购物车
      </Button>
    </Card>
  );
}
