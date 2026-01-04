import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Input, Picker } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listProducts } from '../../services/products';
import { Product, Store } from '../../services/types';
import { listStores } from '../../services/stores';
import usePermission from '../../hooks/usePermission';
import { PERM_HINT_STORE_MGMT_READONLY_PAGE } from '../../constants/permission';

export default function ProductList() {
  const perm = usePermission();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');

  // 门店选择（与分类页保持一致）
  const [stores, setStores] = useState<Store[]>([]);
  const [storePickerIndex, setStorePickerIndex] = useState(0);
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);

  // 与分类页保持一致的筛选与排序控件
  const originRange = ['不限', '华北', '华东', '华南', '西南', '西北', '其他'];
  const packagingRange = ['不限', '散装', '袋装', '罐装', '礼盒'];
  const sortRange = ['默认', '价格升序', '价格降序', '销量优先'];
  const [originPickerIndex, setOriginPickerIndex] = useState(0);
  const [packagingPickerIndex, setPackagingPickerIndex] = useState(0);
  const [sortPickerIndex, setSortPickerIndex] = useState(0);
  const [priceMin, setPriceMin] = useState<string>('');
  const [priceMax, setPriceMax] = useState<string>('');

  const storePickerRange = useMemo(() => ['全部门店', ...stores.map((store) => store.name)], [stores]);

  useEffect(() => {
    void bootstrap();
  }, []);

  async function bootstrap() {
    await loadStores();
    await fetch();
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
    keyword?: string;
    store_id?: number;
    origin?: string;
    packaging?: string;
    min_price?: number;
    max_price?: number;
    sort?: string;
  };

  async function fetch(overrides: QueryOverrides = {}) {
    setLoading(true);
    try {
      const kw = overrides.keyword !== undefined ? overrides.keyword : keyword;
      const storeId = overrides.store_id !== undefined ? overrides.store_id : selectedStoreId;
      const origin = overrides.origin !== undefined ? overrides.origin : originPickerIndex > 0 ? originRange[originPickerIndex] : undefined;
      const packaging = overrides.packaging !== undefined ? overrides.packaging : packagingPickerIndex > 0 ? packagingRange[packagingPickerIndex] : undefined;
      const minPrice = overrides.min_price !== undefined ? overrides.min_price : priceMin ? Number(priceMin) : undefined;
      const maxPrice = overrides.max_price !== undefined ? overrides.max_price : priceMax ? Number(priceMax) : undefined;
      const sort = overrides.sort !== undefined ? overrides.sort : (sortPickerIndex === 1 ? 'price_asc' : sortPickerIndex === 2 ? 'price_desc' : sortPickerIndex === 3 ? 'sales_desc' : undefined);

      const res = await listProducts({
        page: 1,
        limit: 20,
        keyword: kw ? kw.trim() : undefined,
        store_id: storeId,
        origin,
        packaging,
        min_price: minPrice,
        max_price: maxPrice,
        sort,
      });
      const maybe = res as any;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setProducts(items);
    } finally {
      setLoading(false);
    }
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
    void fetch({ store_id: storeId });
  }

  function handleKeywordConfirm(e: any) {
    const value = String(e?.detail?.value ?? keyword).trim();
    setKeyword(value);
    void fetch({ keyword: value });
  }

  function handleOriginChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setOriginPickerIndex(index);
    const origin = index > 0 ? originRange[index] : undefined;
    void fetch({ origin });
  }

  function handlePackagingChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setPackagingPickerIndex(index);
    const packaging = index > 0 ? packagingRange[index] : undefined;
    void fetch({ packaging });
  }

  function handleSortChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    setSortPickerIndex(index);
    const sort = index === 1 ? 'price_asc' : index === 2 ? 'price_desc' : index === 3 ? 'sales_desc' : undefined;
    void fetch({ sort });
  }

  function handlePriceMinConfirm(e: any) {
    const value = String(e?.detail?.value ?? priceMin).trim();
    setPriceMin(value);
    const min_price = value ? Number(value) : undefined;
    void fetch({ min_price });
  }

  function handlePriceMaxConfirm(e: any) {
    const value = String(e?.detail?.value ?? priceMax).trim();
    setPriceMax(value);
    const max_price = value ? Number(value) : undefined;
    void fetch({ max_price });
  }

  return (
    <View style={{ padding: 12 }}>
      {!perm.allowedStoreMgmt && (
        <Text style={{ color: '#999', marginBottom: 8 }}>{PERM_HINT_STORE_MGMT_READONLY_PAGE}</Text>
      )}
      {/* 管理快捷入口（仅有权限且已选择具体门店时显示） */}
      {selectedStoreId && (
        <View style={{ marginBottom: 8, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {perm.allowedStoreAccounts && (
            <Button size="mini" onClick={() => Taro.navigateTo({ url: `/pages/store-accounts/index?store_id=${selectedStoreId}` })}>管理收款账户</Button>
          )}
          {perm.allowedStoreFinance && (
            <Button size="mini" onClick={() => Taro.navigateTo({ url: `/pages/store-finance/index?store_id=${selectedStoreId}` })}>查看财务流水</Button>
          )}
        </View>
      )}
      {/* 门店选择 */}
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
      {/* 搜索与筛选条 */}
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

      {products.map((p) => (
        <View key={p.id} style={{ marginBottom: 12, borderBottomWidth: 1, borderColor: '#eee', paddingBottom: 8 }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{p.name}</Text>
          <Text>价格: {p.price}</Text>
          <Button onClick={() => {
            const storeQuery = selectedStoreId ? `&store_id=${selectedStoreId}` : '';
            Taro.navigateTo({ url: `/pages/product-detail/index?id=${p.id}${storeQuery}` });
          }}>
            查看详情
          </Button>
        </View>
      ))}

      {!loading && products.length === 0 && (
        <Text style={{ color: '#999' }}>暂无商品</Text>
      )}
      {loading && (
        <Text>加载中...</Text>
      )}
    </View>
  );
}
