import React, { useEffect, useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listMyCoupons } from '../../services/coupons';
import { UserCoupon } from '../../services/types';

type TabKey = 'platform' | 'store';

export default function MyCouponsPage() {
  const [coupons, setCoupons] = useState<UserCoupon[]>([]);
  const [loading, setLoading] = useState(false);
  const [tab, setTab] = useState<TabKey>('platform');

  useEffect(() => {
    void fetchCoupons();
  }, []);

  async function fetchCoupons() {
    setLoading(true);
    try {
      const data = await listMyCoupons();
      setCoupons(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error('load my coupons failed', e);
      Taro.showToast({ title: '加载优惠券失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  function switchTab(next: TabKey) {
    setTab(next);
  }

  const platformCoupons = coupons.filter((uc) => !uc.coupon.store_id);
  const storeCoupons = coupons.filter((uc) => !!uc.coupon.store_id);
  const currentList = tab === 'platform' ? platformCoupons : storeCoupons;

  function formatCouponTitle(uc: UserCoupon) {
    const c = uc.coupon;
    if (c.type === 1 && c.amount) return `${c.name}（满减￥${c.amount}）`;
    if (c.type === 2 && c.discount) return `${c.name}（折扣${c.discount}）`;
    return c.name;
  }

  function formatExpire(uc: UserCoupon) {
    return `有效期至：${uc.coupon.end_time?.replace('T', ' ').slice(0, 16)}`;
  }

  return (
    <View style={{ padding: 12 }}>
      {/* 顶部来源 Tab */}
      <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row' }}>
        <Button
          size="mini"
          onClick={() => switchTab('platform')}
          style={{ marginRight: 8, backgroundColor: tab === 'platform' ? '#07c160' : '#f5f5f5', color: tab === 'platform' ? '#fff' : '#333' }}
        >
          平台券
        </Button>
        <Button
          size="mini"
          onClick={() => switchTab('store')}
          style={{ backgroundColor: tab === 'store' ? '#07c160' : '#f5f5f5', color: tab === 'store' ? '#fff' : '#333' }}
        >
          门店券
        </Button>
      </View>

      {loading && <Text>加载中...</Text>}
      {!loading && !currentList.length && <Text>暂无可用优惠券</Text>}

      {!loading &&
        currentList.map((uc) => (
          <View
            key={uc.id}
            style={{
              marginBottom: 12,
              padding: 12,
              borderRadius: 8,
              backgroundColor: '#ffffff',
              boxShadow: '0 2px 6px rgba(0,0,0,0.04)',
            }}
          >
            <Text style={{ display: 'block', fontWeight: 'bold', marginBottom: 4 }}>{formatCouponTitle(uc)}</Text>
            <Text style={{ display: 'block', fontSize: 12, color: '#666', marginBottom: 4 }}>{formatExpire(uc)}</Text>
            <Text style={{ display: 'block', fontSize: 12, color: '#999' }}>
              {uc.coupon.store_id ? `门店券（store_id=${uc.coupon.store_id}）` : '平台券（全平台适用或按券规则）'}
            </Text>
          </View>
        ))}
    </View>
  );
}
