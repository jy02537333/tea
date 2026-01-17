import React, { useEffect, useState } from 'react';
import { View, Text, Button, Image } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
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
  const [filtering, setFiltering] = useState(false);

  useEffect(() => {
    void fetchCart();
    void loadCurrentStore();
  }, []);

  // 页面再次显示时，自动刷新购物车
  useDidShow(() => {
    void fetchCart();
  });

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
      const cartItemsAny: any[] = (raw as any) || [];
      // 直接使用后端在购物车条目里预载的 Product/Sku 与 effective_price，避免再去全量拉产品
      const merged: CartViewItem[] = cartItemsAny.map((c: any) => {
        const embeddedProduct: Product | undefined = (c && (c.Product || c.product)) as Product | undefined;
        return {
          ...(c as CartItem),
          product: embeddedProduct,
        } as CartViewItem;
      });
      setItems(merged);
    } catch (e) {
      console.error('load cart failed', e);
    } finally {
      setLoading(false);
    }
  }

  // 仅保留当前门店商品：
  // 1) 拉取当前门店的商品列表，构造允许的 product_id 集合
  // 2) 找出购物车中不属于本店的条目，确认后逐个移除（调用后端接口）
  async function filterToCurrentStore() {
    if (!currentStore) {
      Taro.showToast({ title: '未选择门店', icon: 'none', duration: 1500 });
      return;
    }
    if (!items.length) {
      Taro.showToast({ title: '购物车为空', icon: 'none', duration: 1500 });
      return;
    }
    setFiltering(true);
    try {
      const res = await getProducts({ page: 1, limit: 500, store_id: currentStore.id });
      const maybe: any = res;
      const storeProducts: Product[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      const allowed = new Set<number>(storeProducts.map((p: any) => Number(p.id)));
      const toRemove = items.filter((it) => !allowed.has(Number(it.product_id)));
      if (!toRemove.length) {
        Taro.showToast({ title: '已全部为本店商品', icon: 'success', duration: 1200 });
        return;
      }
      const mod = await Taro.showModal({
        title: '清理非本店商品',
        content: `将移除 ${toRemove.length} 件非本店/平台商品，是否继续？`,
        confirmText: '移除',
      } as any);
      if (!(mod as any)?.confirm) return;
      for (const it of toRemove) {
        try { await removeCartItem(it.id); } catch (e) { /* 单个失败继续 */ }
      }
      setItems((prev) => prev.filter((it) => allowed.has(Number(it.product_id))));
      Taro.showToast({ title: '已清理', icon: 'success', duration: 1200 });
    } catch (e) {
      console.error('filter to current store failed', e);
      Taro.showToast({ title: '操作失败', icon: 'none', duration: 1500 });
    } finally {
      setFiltering(false);
    }
  }

  async function handleQuantityChange(id: number, quantity: number) {
    if (quantity <= 0) return;
    try {
      await updateCartItem(id, quantity, currentStore?.id);
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
    return items.reduce((sum, it: any) => {
      const price = Number(it?.effective_price ?? (it.product?.price as any) ?? 0) || 0;
      return sum + price * it.quantity;
    }, 0);
  }

  function coverUrl(p?: Product): string {
    if (!p) return 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
    const raw = (p as any).image_url || (p as any).cover || p.images || '';
    const first = String(raw)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)[0];
    return first || 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
  }

  function coverUrlFromItem(it: any): string {
    const p = (it && (it.product || it.Product)) as Product | undefined;
    const byProduct = coverUrl(p);
    if (byProduct && !byProduct.includes('dummyimage')) return byProduct;
    const skuImg = (it && ((it.sku && it.sku.image) || (it.Sku && it.Sku.image))) || '';
    if (skuImg && String(skuImg).trim()) return String(skuImg).trim();
    return 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
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
      {currentStore && items.length > 0 && (
        <View style={{ marginTop: 8, display: 'flex', flexDirection: 'row' }}>
          <Button size="mini" disabled={filtering} onClick={filterToCurrentStore}>
            {filtering ? '处理中...' : `只保留本店（${currentStore.name}）`}
          </Button>
        </View>
      )}
      {items.map((it: any) => (
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
          <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
            <Image src={coverUrlFromItem(it)} mode="aspectFill" style={{ width: 56, height: 56, borderRadius: 4, background: '#f5f5f5' }} />
            <View style={{ marginLeft: 8, flex: 1 }}>
              <Text style={{ display: 'block', fontSize: 16, fontWeight: 'bold' }}>{it.product?.name || `商品#${it.product_id}`}</Text>
              <Text style={{ color: '#666' }}>单价：{Number(it?.effective_price ?? (it.product?.price as any) ?? 0) || 0}</Text>
            </View>
            <Text> 数量: {it.quantity}</Text>
          </View>
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
