import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCart, updateCartItem, removeCartItem } from '../../services/cart';
import { CartItem, Product, Store } from '../../services/types';
import { getProducts } from '../../services/products';
import { getStore } from '../../services/stores';

interface CartViewItem extends CartItem {
  product?: Product;
}

export default function CartPage() {
  const [items, setItems] = useState<CartViewItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentStore, setCurrentStore] = useState<Store | null>(null);

  useEffect(() => {
    void fetchCart();
    void loadCurrentStore();
  }, []);

  async function loadCurrentStore() {
    try {
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const storeId = storeIdRaw ? Number(storeIdRaw) : NaN;
      if (!Number.isNaN(storeId) && storeId > 0) {
        const s = await getStore(storeId);
        setCurrentStore(s as Store);
      }
    } catch (_) {
      // ignore
    }
  }

  async function fetchCart() {
    setLoading(true);
    try {
      const raw = await listCart();
      const cartItems: CartItem[] = (raw as any) || [];
      const productsRes = await getProducts({ page: 1, limit: 200 });
      const maybe: any = productsRes;
      const allProducts: Product[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      const merged: CartViewItem[] = cartItems.map((c) => ({
        ...c,
        product: allProducts.find((p) => p.id === c.product_id),
      }));
      setItems(merged);
    } catch (e) {
      console.error('load cart failed', e);
    } finally {
      setLoading(false);
    }
  }

  async function handleQuantityChange(id: number, quantity: number) {
    if (quantity <= 0) return;
    try {
      await updateCartItem(id, quantity);
      setItems((prev) => prev.map((it) => (it.id === id ? { ...it, quantity } : it)));
    } catch (e) {
      console.error('update cart failed', e);
    }
  }

  async function handleRemove(id: number) {
    try {
      await removeCartItem(id);
      setItems((prev) => prev.filter((it) => it.id !== id));
    } catch (e) {
      console.error('remove cart item failed', e);
    }
  }

  function calcTotal() {
    return items.reduce((sum, it) => {
      const price = Number((it.product?.price as any) ?? 0) || 0;
      return sum + price * it.quantity;
    }, 0);
  }

  function goCheckout() {
    if (!items.length) {
      Taro.showToast({ title: '购物车为空', icon: 'none', duration: 1500 });
      return;
    }
    Taro.navigateTo({ url: '/pages/checkout/index' });
  }

  return (
    <View style={{ padding: 12 }}>
      {currentStore && (
        <View style={{
          marginBottom: 8,
          padding: '6px 10px',
          borderWidth: 1,
          borderStyle: 'solid',
          borderColor: '#07c160',
          borderRadius: 16,
          display: 'inline-block',
          backgroundColor: '#f6ffed',
        }}>
          <Text style={{ color: '#389e0d' }}>当前门店：{currentStore.name}</Text>
        </View>
      )}
      {loading && <Text>加载中...</Text>}
      {!loading && !items.length && <Text>购物车为空</Text>}
      {items.map((it) => (
        <View
          key={it.id}
          style={{
            marginBottom: 12,
            borderBottomWidth: 1,
            borderStyle: 'solid',
            borderColor: '#eee',
            paddingBottom: 8,
          }}
        >
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{it.product?.name || `商品#${it.product_id}`}</Text>
          <Text> 数量: {it.quantity}</Text>
          <View style={{ marginTop: 4, display: 'flex', flexDirection: 'row' }}>
            <Button
              size="mini"
              onClick={() => handleQuantityChange(it.id, it.quantity - 1)}
            >
              -
            </Button>
            <Button
              size="mini"
              onClick={() => handleQuantityChange(it.id, it.quantity + 1)}
              style={{ marginLeft: 8 }}
            >
              +
            </Button>
            <Button
              size="mini"
              type="warn"
              style={{ marginLeft: 8 }}
              onClick={() => handleRemove(it.id)}
            >
              删除
            </Button>
          </View>
        </View>
      ))}

      {items.length > 0 && (
        <View style={{ marginTop: 12 }}>
          <Text>合计: {calcTotal()}</Text>
          <View style={{ marginTop: 8, display: 'flex', flexDirection: 'row' }}>
            <Button style={{ marginLeft: 8 }} type="primary" onClick={goCheckout}>
              去结算
            </Button>
          </View>
        </View>
      )}
    </View>
  );
}
