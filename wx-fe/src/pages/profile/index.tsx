import React, { useEffect, useMemo, useState } from 'react';
import { Button, Image, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { getUserInfo } from '../../services/auth';
import { setToken } from '../../services/api';
import type { User } from '../../services/types';

export default function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void loadProfile();
  }, []);

  usePullDownRefresh(() => {
    void loadProfile();
  });

  async function loadProfile() {
    setLoading(true);
    try {
      const profile = await getUserInfo();
      setUser(profile);
      setError(null);
    } catch (err: any) {
      setUser(null);
      setError(err?.message || '未登录');
    } finally {
      setLoading(false);
      Taro.stopPullDownRefresh();
    }
  }

  function handleLoginNavigate() {
    Taro.navigateTo({ url: '/pages/login/index' }).catch(() => {
      Taro.switchTab({ url: '/pages/index/index' }).catch(() => {});
    });
  }

  function ensureLoggedIn() {
    if (!user) {
      Taro.showToast({ title: '请先登录', icon: 'none' });
      handleLoginNavigate();
      return false;
    }
    return true;
  }

  function handleViewOrders() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {});
  }

  function handleViewCoupons() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/coupons/index' }).catch(() => {});
  }

  function handleManageAddresses() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/address/index' }).catch(() => {});
  }

  function handleServiceTickets() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/after-sale/index' }).catch(() => {});
  }

  function handleViewMembership() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/membership/index' }).catch(() => {});
  }

  function handleLogout() {
    setToken(null);
    setUser(null);
    Taro.showToast({ title: '已退出', icon: 'none' });
  }

  function handleContactSupport() {
    if (process.env.TARO_ENV === 'weapp') {
      Taro.makePhoneCall({ phoneNumber: '4008888888' }).catch(() => {
        Taro.showToast({ title: '拨打失败，请稍后再试', icon: 'none' });
      });
      return;
    }
    Taro.showToast({ title: '联系客服：400-888-8888', icon: 'none' });
  }

  const serviceItems = useMemo(
    () => [
      { key: 'orders', title: '订单', desc: '查看全部订单', action: handleViewOrders },
      { key: 'coupons', title: '优惠券', desc: '查看可用权益', action: handleViewCoupons },
      { key: 'address', title: '收货地址', desc: '管理常用地址', action: handleManageAddresses },
      {
        key: 'membership',
        title: '会员权益',
        desc: user ? '查看积分与成长值' : '登录后可同步权益',
        action: handleViewMembership,
      },
      { key: 'service', title: '售后服务', desc: '进度&售后操作', action: handleServiceTickets },
    ],
    [handleViewOrders, handleViewCoupons, handleManageAddresses, handleViewMembership, handleServiceTickets, user],
  );

  return (
    <View style={{ minHeight: '100vh', backgroundColor: '#f5f6f8', padding: 20 }}>
      <View style={{ backgroundColor: '#fff', borderRadius: 16, padding: 20, display: 'flex', alignItems: 'center' }}>
        <Image
          src={user?.avatar || 'https://dummyimage.com/120x120/dcdcdc/333333&text=Tea'}
          style={{ width: 80, height: 80, borderRadius: 40, marginRight: 16 }}
        />
        <View style={{ flex: 1 }}>
          <Text style={{ fontSize: 20, fontWeight: 'bold', display: 'block' }}>{user ? user.nickname || '微信用户' : '未登录用户'}</Text>
          {user?.phone && <Text style={{ color: '#666', display: 'block', marginTop: 6 }}>{user.phone}</Text>}
          {error && <Text style={{ color: '#ff4d4f', display: 'block', marginTop: 6 }}>{loading ? '加载中...' : error}</Text>}
        </View>
        {!user ? (
          <Button size="mini" onClick={handleLoginNavigate} loading={loading}>
            去登录
          </Button>
        ) : (
          <Button size="mini" onClick={handleLogout}>
            退出
          </Button>
        )}
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 16, padding: 20, marginTop: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>我的服务</Text>
        <View style={{ display: 'flex', flexWrap: 'wrap', marginTop: 12 }}>
          {serviceItems.map((item, index) => (
            <View
              key={item.key}
              style={{
                width: '50%',
                paddingTop: 12,
                paddingBottom: 12,
                textAlign: 'center',
                borderRight: index % 2 === 0 ? '1px solid #f0f0f0' : 'none',
                borderBottom: index < serviceItems.length - 2 ? '1px solid #f0f0f0' : 'none',
              }}
              onClick={item.action}
            >
              <Text style={{ display: 'block', fontSize: 18, fontWeight: 'bold' }}>{item.title}</Text>
              <Text style={{ color: '#999', fontSize: 12 }}>{item.desc}</Text>
            </View>
          ))}
        </View>
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 16, padding: 20, marginTop: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>账号操作</Text>
        <Button style={{ marginTop: 12 }} onClick={handleLoginNavigate} loading={loading}>
          {user ? '切换账号' : '登录 / 注册'}
        </Button>
        <Button style={{ marginTop: 12 }} onClick={handleContactSupport}>
          联系客服
        </Button>
      </View>
    </View>
  );
}
