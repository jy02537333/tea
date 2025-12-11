import React, { useEffect, useState } from 'react';
import { View, Text, Input, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getProducts } from '../../services/products';
import { listStores } from '../../services/stores';
import { Product, Store } from '../../services/types';

export default function IndexPage() {
  const [keyword, setKeyword] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [stores, setStores] = useState<Store[]>([]);
  const [currentStoreId, setCurrentStoreId] = useState<number | undefined>(undefined);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // 初始化加载附近门店和商品列表
    void loadStoresAndProducts();
  }, []);

  async function loadStoresAndProducts() {
    setLoading(true);
    try {
      await Promise.all([fetchStores(), fetchProducts()]);
    } finally {
      setLoading(false);
    }
  }

  async function fetchStores() {
    try {
      const res = await listStores({ page: 1, limit: 20 });
      const maybe: any = res;
      let items: Store[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setStores(items);
      if (!currentStoreId && items.length > 0) setCurrentStoreId(items[0].id);
    } catch (e) {
      console.error('load stores failed', e);
    }
  }

  async function fetchProducts(params?: { keyword?: string; store_id?: number }) {
    try {
      const searchKeyword = params?.keyword !== undefined ? params.keyword : keyword || undefined;
      const storeId = params?.store_id !== undefined ? params.store_id : currentStoreId;
      const res = await getProducts({
        page: 1,
        limit: 20,
        keyword: searchKeyword,
        store_id: storeId,
      });
      const maybe: any = res;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setProducts(items);
    } catch (e) {
      console.error('load products failed', e);
    }
  }

  function handleSearchConfirm(e: any) {
    const value: string = e?.detail?.value ?? keyword;
    setKeyword(value);
    void fetchProducts({ keyword: value, store_id: currentStoreId });
  }

  function handleStoreChange(id: number) {
    setCurrentStoreId(id);
    void fetchProducts({ store_id: id });
  }

  function goStoreFinance() {
    if (!currentStoreId) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' });
      return;
    }
    Taro.navigateTo({ url: `/pages/store-finance/index?store_id=${currentStoreId}` });
  }

  function goStoreAccounts() {
    if (!currentStoreId) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' });
      return;
    }
    Taro.navigateTo({ url: `/pages/store-accounts/index?store_id=${currentStoreId}` });
  }

  function goStoreActivities() {
    if (!currentStoreId) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' });
      return;
    }
    Taro.navigateTo({ url: `/pages/activities/index?store_id=${currentStoreId}` });
  }

  return (
    <View style={{ padding: 12 }}>
      {/* 搜索栏 */}
      <View style={{ marginBottom: 12 }}>
        <Input
          type="text"
          placeholder="搜索商品"
          value={keyword}
          onInput={(e) => setKeyword((e.detail as any).value)}
          onConfirm={handleSearchConfirm}
        />
      </View>

      <View style={{ marginBottom: 12 }}>
        <Button size="mini" onClick={() => Taro.navigateTo({ url: '/pages/category/index' })}>
          进入分类 / 商品列表
        </Button>
      </View>

      <View style={{ marginBottom: 12 }}>
        <Button size="mini" onClick={goStoreActivities}>
          查看门店活动
        </Button>
      </View>

      <View style={{ marginBottom: 12 }}>
        <Button size="mini" onClick={goStoreFinance}>
          查看门店财务流水
        </Button>
      </View>

      <View style={{ marginBottom: 12 }}>
        <Button size="mini" onClick={goStoreAccounts}>
          门店收款账户设置
        </Button>
      </View>

      {/* 门店横向选择（简单文本版） */}
      <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row', flexWrap: 'wrap' }}>
        {stores.map((store) => (
          <View
            key={store.id}
            style={{
              padding: 8,
              marginRight: 8,
              marginBottom: 8,
              borderRadius: 4,
              backgroundColor: store.id === currentStoreId ? '#007aff' : '#f0f0f0',
            }}
            onClick={() => handleStoreChange(store.id)}
          >
            <Text style={{ color: store.id === currentStoreId ? '#fff' : '#333' }}>{store.name}</Text>
          </View>
        ))}
      </View>

      {/* 商品列表 */}
      <View>
        {loading && <Text>加载中...</Text>}
        {!loading && products.length === 0 && <Text>暂无商品</Text>}
        {products.map((p) => (
          <View
            key={p.id}
            style={{
              marginBottom: 12,
              borderBottomWidth: 1,
              borderStyle: 'solid',
              borderColor: '#eee',
              paddingBottom: 8,
            }}
          >
            <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{p.name}</Text>
            <Text> 价格: {p.price}</Text>
          </View>
        ))}
      </View>
    </View>
  );
}
