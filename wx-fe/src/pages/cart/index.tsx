import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCart, updateCartItem, removeCartItem } from '../../services/cart';
import { createOrderFromCart } from '../../services/orders';
// types removed for Babel compatibility during Taro build

export default function CartPage() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => { fetchCart(); }, []);

  async function fetchCart() {
    setLoading(true);
    try {
      const res = await listCart();
      setItems(res || []);
    } finally {
      setLoading(false);
    }
  }

  function total() {
    return items.reduce((sum, it) => sum + (Number((it && it.price) || 0) || 0) * (it && it.quantity || 1), 0);
  }

  async function inc(it) {
    await updateCartItem(it.id, (it.quantity || 1) + 1);
    fetchCart();
  }
  async function dec(it) {
    const q = (it.quantity || 1) - 1;
    if (q <= 0) return;
    await updateCartItem(it.id, q);
    fetchCart();
  }
  async function del(it) {
    await removeCartItem(it.id);
    fetchCart();
  }

  async function checkout() {
    try {
      const order = await createOrderFromCart({ delivery_type: 2 });
      Taro.showToast({ title: '下单成功', icon: 'success' });
      console.log('order', order);
    } catch (e) {
      Taro.showToast({ title: '下单失败', icon: 'error' });
    }
  }

  return (
    <View style={{ padding: 12 }}>
      {items.map((it) => (
        <View key={it.id} style={{ marginBottom: 8, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 6 }}>
          <Text>{(it && it.name) || it.product_id}</Text>
          <Text> 数量: {it.quantity}</Text>
          <View style={{ flexDirection: 'row', marginTop: 4 }}>
            <Button onClick={() => dec(it)}> - </Button>
            <Button onClick={() => inc(it)}> + </Button>
            <Button onClick={() => del(it)} type="warn">删除</Button>
          </View>
        </View>
      ))}
      <Text>总价: {total()}</Text>
      <Button onClick={checkout} disabled={loading}>下单</Button>
    </View>
  );
}
