import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView, Input, Button, Picker } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { listCategories } from '../../services/categories';
import { getProducts } from '../../services/products';
import { listStores } from '../../services/stores';
import { Category, Product, Store } from '../../services/types';

const PAGE_SIZE = 10;

export default function CategoryPage() {
  const router = useRouter();
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | undefined>(undefined);

  const [stores, setStores] = useState<Store[]>([]);
  const [storePickerIndex, setStorePickerIndex] = useState(0);
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);

  const [keyword, setKeyword] = useState('');

  // 筛选与排序（产地/包装/价格区间/排序）
  const originRange = ['不限', '华北', '华东', '华南', '西南', '西北', '其他'];
  const packagingRange = ['不限', '散装', '袋装', '罐装', '礼盒'];
  const sortRange = ['默认', '价格升序', '价格降序', '销量优先'];
  const [originPickerIndex, setOriginPickerIndex] = useState(0);
  const [packagingPickerIndex, setPackagingPickerIndex] = useState(0);
  const [sortPickerIndex, setSortPickerIndex] = useState(0);
  const [priceMin, setPriceMin] = useState<string>('');
  const [priceMax, setPriceMax] = useState<string>('');

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
      // 路由入参接管门店（在门店列表加载后设置 picker 状态）
      const paramSidRaw = router?.params?.store_id;
      const paramSid = paramSidRaw ? Number(paramSidRaw) : NaN;
      if (!Number.isNaN(paramSid) && paramSid > 0) {
        setSelectedStoreId(paramSid);
        try { Taro.setStorageSync('current_store_id', String(paramSid)); } catch (_) {}
        const index = stores.findIndex((s) => s.id === paramSid);
        setStorePickerIndex(index >= 0 ? index + 1 : 0);
      }
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

  type QueryOverrides = {
    category_id?: number;
    keyword?: string;
    store_id?: number;
    origin?: string;
    packaging?: string;
    min_price?: number;
    max_price?: number;
    sort?: string;
  };

  async function fetchProducts(options: { reset?: boolean; overrides?: QueryOverrides } = {}) {
    if (loadingProducts) return;
    const { reset = false, overrides } = options;
    const targetPage = reset ? 1 : page + 1;
    const categoryId = overrides?.category_id !== undefined ? overrides.category_id : selectedCategoryId;
    const storeId = overrides?.store_id !== undefined ? overrides.store_id : selectedStoreId;
    const kw = overrides?.keyword !== undefined ? overrides.keyword : keyword;
    const origin = overrides?.origin !== undefined ? overrides.origin : originPickerIndex > 0 ? originRange[originPickerIndex] : undefined;
    const packaging = overrides?.packaging !== undefined ? overrides.packaging : packagingPickerIndex > 0 ? packagingRange[packagingPickerIndex] : undefined;
    const minPrice = overrides?.min_price !== undefined ? overrides.min_price : priceMin ? Number(priceMin) : undefined;
    const maxPrice = overrides?.max_price !== undefined ? overrides.max_price : priceMax ? Number(priceMax) : undefined;
    const sort = overrides?.sort !== undefined ? overrides.sort : (sortPickerIndex === 1 ? 'price_asc' : sortPickerIndex === 2 ? 'price_desc' : sortPickerIndex === 3 ? 'sales_desc' : undefined);

    setLoadingProducts(true);
    try {
      const response = await getProducts({
        page: targetPage,
        limit: PAGE_SIZE,
        category_id: categoryId,
        keyword: kw ? kw.trim() : undefined,
        store_id: storeId,
        origin,
        packaging,
        min_price: minPrice,
        max_price: maxPrice,
        sort,
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
    try {
      if (storeId) Taro.setStorageSync('current_store_id', String(storeId));
      else Taro.removeStorageSync('current_store_id');
    } catch (_) {}
    void fetchProducts({ reset: true, overrides: { store_id: storeId } });
  }

  function handleKeywordConfirm(e: any) {
    const value = (e?.detail?.value ?? keyword).trim();
    setKeyword(value);
    void fetchProducts({ reset: true, overrides: { keyword: value } });
  }

  function handleOriginChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setOriginPickerIndex(index);
    const origin = index > 0 ? originRange[index] : undefined;
    void fetchProducts({ reset: true, overrides: { origin } });
  }

  function handlePackagingChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setPackagingPickerIndex(index);
    const packaging = index > 0 ? packagingRange[index] : undefined;
    void fetchProducts({ reset: true, overrides: { packaging } });
  }

  function handleSortChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setSortPickerIndex(index);
    const sort = index === 1 ? 'price_asc' : index === 2 ? 'price_desc' : index === 3 ? 'sales_desc' : undefined;
    void fetchProducts({ reset: true, overrides: { sort } });
  }

  function handlePriceMinConfirm(e: any) {
    const value = String(e?.detail?.value ?? priceMin).trim();
    setPriceMin(value);
    const min_price = value ? Number(value) : undefined;
    void fetchProducts({ reset: true, overrides: { min_price } });
  }

  function handlePriceMaxConfirm(e: any) {
    const value = String(e?.detail?.value ?? priceMax).trim();
    setPriceMax(value);
    const max_price = value ? Number(value) : undefined;
    void fetchProducts({ reset: true, overrides: { max_price } });
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

      {/* 筛选与排序区域 */}
      <View style={{ marginBottom: 16 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>筛选与排序</Text>
        <View style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 8 }}>
          {/* 产地 */}
          <View>
            <Text style={{ fontSize: 12, color: '#888' }}>产地</Text>
            <Picker mode="selector" range={originRange} onChange={handleOriginChange} value={originPickerIndex}>
              <View style={{ marginTop: 6, padding: 10, borderWidth: 1, borderStyle: 'solid', borderColor: '#ddd', borderRadius: 6 }}>
                <Text>{originRange[originPickerIndex] || '不限'}</Text>
              </View>
            </Picker>
          </View>

          {/* 包装 */}
          <View>
            <Text style={{ fontSize: 12, color: '#888' }}>包装</Text>
            <Picker mode="selector" range={packagingRange} onChange={handlePackagingChange} value={packagingPickerIndex}>
              <View style={{ marginTop: 6, padding: 10, borderWidth: 1, borderStyle: 'solid', borderColor: '#ddd', borderRadius: 6 }}>
                <Text>{packagingRange[packagingPickerIndex] || '不限'}</Text>
              </View>
            </Picker>
          </View>

          {/* 价格区间 */}
          <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center', gap: 8 }}>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 12, color: '#888' }}>最低价</Text>
              <Input type="number" value={priceMin} placeholder="例如 10" onConfirm={handlePriceMinConfirm} onInput={(e) => setPriceMin(String((e.detail as any).value))} />
            </View>
            <Text style={{ marginTop: 18 }}>-</Text>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 12, color: '#888' }}>最高价</Text>
              <Input type="number" value={priceMax} placeholder="例如 100" onConfirm={handlePriceMaxConfirm} onInput={(e) => setPriceMax(String((e.detail as any).value))} />
            </View>
          </View>

          {/* 排序 */}
          <View>
            <Text style={{ fontSize: 12, color: '#888' }}>排序</Text>
            <Picker mode="selector" range={sortRange} onChange={handleSortChange} value={sortPickerIndex}>
              <View style={{ marginTop: 6, padding: 10, borderWidth: 1, borderStyle: 'solid', borderColor: '#ddd', borderRadius: 6 }}>
                <Text>{sortRange[sortPickerIndex] || '默认'}</Text>
              </View>
            </Picker>
          </View>
        </View>
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
              onClick={() => {
                const storeQuery = selectedStoreId ? `&store_id=${selectedStoreId}` : '';
                Taro.navigateTo({ url: `/pages/product-detail/index?id=${product.id}${storeQuery}` });
              }}
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
