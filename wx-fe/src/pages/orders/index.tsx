import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listOrders } from '../../services/orders';
import { Order, Store } from '../../services/types';
import { getStore } from '../../services/stores';

const STATUS_TEXT: Record<number, string> = {
  1: '待支付',
  2: '已付款',
  3: '配送中',
  4: '已完成',
  5: '已取消',
  6: '已堂食',
  7: '外卖出餐',
};

function toNumber(value?: number | string): number | undefined {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = parseInt(value, 10);
    return Number.isNaN(parsed) ? undefined : parsed;
  }
  return undefined;
}

function getStatusText(status?: number | string): string {
  const n = toNumber(status);
  if (!n) return '--';
  return STATUS_TEXT[n] || '--';
}

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<number | undefined>(undefined);
  const [currentStore, setCurrentStore] = useState<Store | null>(null);

  const hasAnyDineIn = useMemo(() => orders.some((o) => (o.table_no || '').trim()), [orders]);

  useEffect(() => {
    void loadCurrentStore();
  }, []);

  useEffect(() => {
    void fetchOrders();
  }, [status, currentStore?.id]);

  async function loadCurrentStore() {
    try {
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const storeId = storeIdRaw ? Number(storeIdRaw) : NaN;
      if (!Number.isNaN(storeId) && storeId > 0) {
        const s = await getStore(storeId);
        const st = (s as any)?.status;
        if (typeof st === 'number' && st === 0) {
          setCurrentStore(null);
        } else {
          setCurrentStore(s as Store);
        }
      }
    } catch (_) {
      // ignore
    }
  }

  async function fetchOrders() {
    setLoading(true);
    try {
      const sid = currentStore?.id;
      const params: { page?: number; limit?: number; status?: number; store_id?: number } = { page: 1, limit: 20 };
      if (typeof status === 'number') params.status = status;
      if (sid && Number.isFinite(sid) && sid > 0) params.store_id = sid;
      const res = await listOrders(params);
      const maybe: any = res;
      const data: Order[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      setOrders(data);
    } catch (e) {
      console.error('load orders failed', e);
    } finally {
      setLoading(false);
    }
  }

  function changeStatus(s?: number) {
    setStatus(s);
  }

  function goDetail(id: number) {
    const sid = currentStore?.id;
    const url = sid && sid > 0
      ? `/pages/order-detail/index?id=${id}&store_id=${sid}`
      : `/pages/order-detail/index?id=${id}`;
    Taro.navigateTo({ url });
  }

  return (
    <View data-testid="page-orders" style={{ padding: 12 }}>
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
      {/* 状态切换 Tab（简化版） */}
      <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row' }}>
        <Button size="mini" onClick={() => changeStatus(undefined)}>
          全部
        </Button>
        <Button size="mini" onClick={() => changeStatus(1)} style={{ marginLeft: 8 }}>
          待支付
        </Button>
        <Button size="mini" onClick={() => changeStatus(2)} style={{ marginLeft: 8 }}>
          已支付
        </Button>
        <Button size="mini" onClick={() => changeStatus(4)} style={{ marginLeft: 8 }}>
          已完成
        </Button>
      </View>

      {loading && <Text>加载中...</Text>}
      {!loading && !orders.length && <Text>暂无订单</Text>}
      {orders.map((o) => (
        (() => {
          const items = Array.isArray(o.items) ? o.items : [];
          const first = items[0];
          const productName = (first?.product_name || first?.sku_name || '').trim() || '商品';
          const extraCount = items.length > 1 ? items.length - 1 : 0;
          const title = extraCount > 0 ? `${productName} 等${items.length}件` : productName;
          const imageUrl = (first?.image || '').trim();
          const tableNo = (o.table_no || '').trim();
          const statusText = getStatusText(o.status);
          return (
        <View
          key={o.id}
          style={{
            marginBottom: 12,
            borderBottomWidth: 1,
            borderStyle: 'solid',
            borderColor: '#eee',
            paddingBottom: 8,
          }}
        >
          <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
            {imageUrl ? (
              <Image
                src={imageUrl}
                mode="aspectFill"
                style={{ width: 56, height: 56, marginRight: 10, borderRadius: 6, backgroundColor: '#f5f5f5' }}
              />
            ) : (
              <View style={{ width: 56, height: 56, marginRight: 10, borderRadius: 6, backgroundColor: '#f5f5f5' }} />
            )}

            <View style={{ flex: 1 }}>
              <View>
                <Text>{title}</Text>
              </View>
              {hasAnyDineIn && (
                <View style={{ marginTop: 2 }}>
                  <Text>桌号: {tableNo || '--'}</Text>
                </View>
              )}
              <View style={{ marginTop: 2 }}>
                <Text>状态: {statusText}</Text>
              </View>
              <View style={{ marginTop: 2 }}>
                <Text>订单号: {o.order_no}</Text>
              </View>
              <View style={{ marginTop: 2 }}>
                <Text>金额: {o.pay_amount}</Text>
              </View>
            </View>
          </View>
          <Button size="mini" style={{ marginTop: 4 }} onClick={() => goDetail(o.id)}>
            查看详情
          </Button>
        </View>
          );
        })()
      ))}
    </View>
  );
}
