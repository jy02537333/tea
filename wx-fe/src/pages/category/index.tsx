import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, ScrollView, Input, Picker } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { listCategories } from '../../services/categories';
import { getProducts } from '../../services/products';
import { listStores } from '../../services/stores';
import { Category, Product, Store } from '../../services/types';
import usePermission from '../../hooks/usePermission';
import { PERM_HINT_STORE_MGMT_READONLY_PAGE } from '../../constants/permission';
import './index.scss';
import ProductCard from '../../components/ProductCard';

const PAGE_SIZE = 10;

export default function CategoryPage() {
  const router = useRouter();
  const perm = usePermission();
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

      // 路由入参接管分类（在分类列表加载后设置初始选中状态）
      const paramCidRaw = (router as any)?.params?.category_id;
      const paramCid = paramCidRaw ? Number(paramCidRaw) : NaN;
      if (!Number.isNaN(paramCid) && paramCid > 0) {
        setSelectedCategoryId(paramCid);
      }

      // 路由入参接管门店（在门店列表加载后设置 picker 状态）
      const paramSidRaw = router?.params?.store_id;
      const paramSid = paramSidRaw ? Number(paramSidRaw) : NaN;
      if (!Number.isNaN(paramSid) && paramSid > 0) {
        setSelectedStoreId(paramSid);
        try { Taro.setStorageSync('current_store_id', String(paramSid)); } catch (_) {}
        const index = stores.findIndex((s) => s.id === paramSid);
        setStorePickerIndex(index >= 0 ? index + 1 : 0);
      }

      // 若路由指定 category_id，则首屏直接拉该分类；否则走默认逻辑
      if (!Number.isNaN(paramCid) && paramCid > 0) {
        await fetchProducts({ reset: true, overrides: { category_id: paramCid } });
      } else {
        await fetchProducts({ reset: true });
      }
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
    <View className="page-category">
      {!perm.allowedStoreMgmt && (
        <Text style={{ color: '#999', marginBottom: 8 }}>{PERM_HINT_STORE_MGMT_READONLY_PAGE}</Text>
      )}
      {/* 管理快捷入口（仅有权限且已选择具体门店时显示） */}
      {selectedStoreId && (
        <View style={{ marginBottom: 8, display: 'flex', flexDirection: 'row', flexWrap: 'wrap' }}>
          {perm.allowedStoreAccounts && (
            <View
              style={{ padding: '6px 10px', borderRadius: 12, backgroundColor: '#f0f0f0', marginRight: 8, marginBottom: 8 }}
              onClick={() => Taro.navigateTo({ url: `/pages/store-accounts/index?store_id=${selectedStoreId}` }).catch(() => {})}
            >
              <Text style={{ fontSize: 12, color: '#333' }}>管理收款账户</Text>
            </View>
          )}
          {perm.allowedStoreFinance && (
            <View
              style={{ padding: '6px 10px', borderRadius: 12, backgroundColor: '#f0f0f0', marginRight: 8, marginBottom: 8 }}
              onClick={() => Taro.navigateTo({ url: `/pages/store-finance/index?store_id=${selectedStoreId}` }).catch(() => {})}
            >
              <Text style={{ fontSize: 12, color: '#333' }}>查看财务流水</Text>
            </View>
          )}
        </View>
      )}
      <View className="search-bar">
        <Input
          type="text"
          placeholder="搜索商品关键字"
          value={keyword}
          onInput={(e) => setKeyword((e.detail as any).value)}
          onConfirm={handleKeywordConfirm}
        />
      </View>

      <View className="filters">
        <Text className="label">选择门店</Text>
        <Picker mode="selector" range={storePickerRange} onChange={handleStoreChange} value={storePickerIndex}>
          <View className="picker">
            <Text>{storePickerRange[storePickerIndex] || '全部门店'}</Text>
          </View>
        </Picker>
      </View>

      {/* 筛选与排序区域 */}
      <View className="filters">
        <Text className="label">筛选与排序</Text>
        <View className="filters-body">
          {/* 产地 */}
          <View>
            <Text className="sublabel">产地</Text>
            <Picker mode="selector" range={originRange} onChange={handleOriginChange} value={originPickerIndex}>
              <View className="picker sm">
                <Text>{originRange[originPickerIndex] || '不限'}</Text>
              </View>
            </Picker>
          </View>

          {/* 包装 */}
          <View>
            <Text className="sublabel">包装</Text>
            <Picker mode="selector" range={packagingRange} onChange={handlePackagingChange} value={packagingPickerIndex}>
              <View className="picker sm">
                <Text>{packagingRange[packagingPickerIndex] || '不限'}</Text>
              </View>
            </Picker>
          </View>

          {/* 价格区间 */}
          <View className="price-range">
            <View style={{ flex: 1 }}>
              <Text className="sublabel">最低价</Text>
              <Input type="number" value={priceMin} placeholder="例如 10" onConfirm={handlePriceMinConfirm} onInput={(e) => setPriceMin(String((e.detail as any).value))} />
            </View>
            <Text style={{ marginTop: 18 }}>-</Text>
            <View style={{ flex: 1 }}>
              <Text className="sublabel">最高价</Text>
              <Input type="number" value={priceMax} placeholder="例如 100" onConfirm={handlePriceMaxConfirm} onInput={(e) => setPriceMax(String((e.detail as any).value))} />
            </View>
          </View>

          {/* 排序 */}
          <View>
            <Text className="sublabel">排序</Text>
            <Picker mode="selector" range={sortRange} onChange={handleSortChange} value={sortPickerIndex}>
              <View className="picker sm">
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
        <View className="grid">
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              showCover
              onClick={() => {
                const storeQuery = selectedStoreId ? `&store_id=${selectedStoreId}` : '';
                Taro.navigateTo({ url: `/pages/product-detail/index?id=${product.id}${storeQuery}` });
              }}
            />
          ))}
        </View>
      </View>

      <View className="load-more">
        {hasMore ? (
          <View
            style={{
              display: 'inline-block',
              padding: '6px 12px',
              borderRadius: 12,
              backgroundColor: '#f0f0f0',
              opacity: loadingProducts ? 0.6 : 1,
            }}
            onClick={() => {
              if (loadingProducts) return;
              handleLoadMore();
            }}
          >
            <Text style={{ fontSize: 12, color: '#333' }}>{loadingProducts ? '加载中...' : '加载更多'}</Text>
          </View>
        ) : (
          <Text style={{ color: '#999' }}>已加载全部商品</Text>
        )}
      </View>
    </View>
  );
}
