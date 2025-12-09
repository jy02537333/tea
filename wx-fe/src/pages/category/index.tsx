import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView, Input, Button, Picker } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCategories } from '../../services/categories';
import { getProducts } from '../../services/products';
import { listStores } from '../../services/stores';
import { Category, Product, Store } from '../../services/types';

const PAGE_SIZE = 10;

export default function CategoryPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | undefined>(undefined);

  const [stores, setStores] = useState<Store[]>([]);
  const [storePickerIndex, setStorePickerIndex] = useState(0);
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);

  const [keyword, setKeyword] = useState('');

  const [products, setProducts] = useState<Product[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [initializing, setInitializing] = useState(true);

  useEffect(() => {
    void bootstrap();
  }, []);

  async function bootstrap() {
    setInitializing(true);
    try {
      await Promise.all([loadCategories(), loadStores()]);
      await fetchProducts({ reset: true });
    } finally {
      setInitializing(false);
    }
  }

  async function loadCategories() {
    try {
      const res = await listCategories({ status: 1 });
      const list = Array.isArray(res?.data)
        ? res.data
        : Array.isArray((res as any)?.items)
          ? (res as any).items
          : Array.isArray((res as any))
            ? (res as any)
            : [];
      setCategories(list);
    } catch (error) {
      console.error('load categories failed', error);
      Taro.showToast({ title: '加载分类失败', icon: 'none' });
    }
  }

  async function loadStores() {
    try {
      const res = await listStores({ page: 1, limit: 50 });
      const list = Array.isArray(res?.data)
        ? res.data
        : Array.isArray((res as any)?.items)
          ? (res as any).items
          : Array.isArray((res as any))
            ? (res as any)
            : [];
      setStores(list);
    } catch (error) {
      console.error('load stores failed', error);
      Taro.showToast({ title: '加载门店失败', icon: 'none' });
    }
  }

  type QueryOverrides = { category_id?: number; keyword?: string; store_id?: number };

  async function fetchProducts(options: { reset?: boolean; overrides?: QueryOverrides } = {}) {
    if (loadingProducts) return;
    const { reset = false, overrides } = options;
    const targetPage = reset ? 1 : page + 1;
    const categoryId = overrides?.category_id !== undefined ? overrides.category_id : selectedCategoryId;
    const storeId = overrides?.store_id !== undefined ? overrides.store_id : selectedStoreId;
    const kw = overrides?.keyword !== undefined ? overrides.keyword : keyword;

    setLoadingProducts(true);
    try {
      const response = await getProducts({
        page: targetPage,
        limit: PAGE_SIZE,
        category_id: categoryId,
        keyword: kw ? kw.trim() : undefined,
        store_id: storeId,
      });

      const list = Array.isArray(response?.data)
        ? response.data
        : Array.isArray((response as any)?.items)
          ? (response as any).items
          : Array.isArray((response as any))
            ? (response as any)
            : [];

      setProducts((prev) => (reset ? list : [...prev, ...list]));
      setPage(targetPage);

      const total = typeof response?.total === 'number' ? response.total : undefined;
      const limit = typeof response?.limit === 'number' ? response.limit : PAGE_SIZE;
      setHasMore(total !== undefined ? targetPage * limit < total : list.length === limit);
    } catch (error) {
      console.error('load products failed', error);
      Taro.showToast({ title: '加载商品失败', icon: 'none' });
    } finally {
      setLoadingProducts(false);
    }
  }

  function handleCategorySelect(categoryId?: number) {
    setSelectedCategoryId(categoryId);
    void fetchProducts({ reset: true, overrides: { category_id: categoryId } });
  }

  function handleStoreChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setStorePickerIndex(index);
    const storeId = index === 0 ? undefined : stores[index - 1]?.id;
    setSelectedStoreId(storeId);
    void fetchProducts({ reset: true, overrides: { store_id: storeId } });
  }

  function handleKeywordConfirm(e: any) {
    const value = (e?.detail?.value ?? keyword).trim();
    setKeyword(value);
    void fetchProducts({ reset: true, overrides: { keyword: value } });
  }

  function handleLoadMore() {
    if (!hasMore || loadingProducts) return;
    void fetchProducts();
  }

  const storePickerRange = useMemo(() => ['全部门店', ...stores.map((store) => store.name)], [stores]);

  return (
    <View style={{ padding: 12 }}>
      <View style={{ marginBottom: 12 }}>
        <Input
          type="text"
          placeholder="搜索商品关键字"
          value={keyword}
          onInput={(e) => setKeyword((e.detail as any).value)}
          onConfirm={handleKeywordConfirm}
        />
      </View>

      <View style={{ marginBottom: 16 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>选择门店</Text>
        <Picker mode="selector" range={storePickerRange} onChange={handleStoreChange} value={storePickerIndex}>
          <View
            style={{
              marginTop: 8,
              padding: 12,
              borderWidth: 1,
              borderStyle: 'solid',
              borderColor: '#ddd',
              borderRadius: 6,
            }}
          >
            <Text>{storePickerRange[storePickerIndex] || '全部门店'}</Text>
          </View>
        </Picker>
      </View>

      <View style={{ marginBottom: 12 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>商品分类</Text>
        <ScrollView
          scrollX
          style={{ whiteSpace: 'nowrap', marginTop: 8 }}
          showScrollbar={false}
        >
          <View style={{ display: 'flex', flexDirection: 'row' }}>
            <View
              style={{
                padding: '8px 16px',
                marginRight: 8,
                borderRadius: 16,
                backgroundColor: selectedCategoryId === undefined ? '#07c160' : '#f0f0f0',
              }}
              onClick={() => handleCategorySelect(undefined)}
            >
              <Text style={{ color: selectedCategoryId === undefined ? '#fff' : '#333' }}>全部</Text>
            </View>
            {categories.map((category) => {
              const active = category.id === selectedCategoryId;
              return (
                <View
                  key={category.id}
                  style={{
                    padding: '8px 16px',
                    marginRight: 8,
                    borderRadius: 16,
                    backgroundColor: active ? '#07c160' : '#f0f0f0',
                  }}
                  onClick={() => handleCategorySelect(category.id)}
                >
                  <Text style={{ color: active ? '#fff' : '#333' }}>{category.name}</Text>
                </View>
              );
            })}
          </View>
        </ScrollView>
      </View>

      <View>
        {initializing && <Text>加载中...</Text>}
        {!initializing && products.length === 0 && <Text>暂无商品</Text>}
        {products.map((product) => (
          <View
            key={product.id}
            style={{
              padding: 12,
              borderBottomWidth: 1,
              borderStyle: 'solid',
              borderColor: '#f2f2f2',
            }}
          >
            <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{product.name}</Text>
            <Text style={{ display: 'block', marginTop: 4 }}>现价：¥{product.price}</Text>
            {product.original_price && (
              <Text style={{ display: 'block', marginTop: 2, color: '#999', textDecoration: 'line-through' }}>
                原价：¥{product.original_price}
              </Text>
            )}
            <Button
              size="mini"
              style={{ marginTop: 8, width: 120 }}
              onClick={() => Taro.navigateTo({ url: `/pages/product-detail/index?id=${product.id}` })}
            >
              查看详情
            </Button>
          </View>
        ))}
      </View>

      <View style={{ marginTop: 16, textAlign: 'center' }}>
        {hasMore ? (
          <Button size="mini" loading={loadingProducts} onClick={handleLoadMore}>
            {loadingProducts ? '加载中...' : '加载更多'}
          </Button>
        ) : (
          <Text style={{ color: '#999' }}>已加载全部商品</Text>
        )}
      </View>
    </View>
  );
}
