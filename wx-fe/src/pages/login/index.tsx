import React, { useEffect, useState } from 'react';
import { Button, Input, Text, View, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getUserInfo, login } from '../../services/auth';
import { setToken } from '../../services/api';
import type { User } from '../../services/types';

export default function LoginPage() {
  const [devOpenId, setDevOpenId] = useState('');
  const [loading, setLoading] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [fetchingProfile, setFetchingProfile] = useState(false);

  useEffect(() => {
    void fetchProfile();
  }, []);

  async function fetchProfile() {
    setFetchingProfile(true);
    try {
      const profile = await getUserInfo();
      setUser(profile);
    } catch (err) {
      setUser(null);
    } finally {
      setFetchingProfile(false);
    }
  }

  async function handleWxLogin() {
    setLoading(true);
    try {
      const res = await Taro.login();
      if (!res.code) throw new Error('未获取到微信 code');
      await login({ code: res.code });
      await fetchProfile();
      Taro.showToast({ title: '登录成功', icon: 'success', duration: 1200 });
      setTimeout(() => {
        Taro.switchTab({ url: '/pages/index/index' }).catch(() => {
          Taro.redirectTo({ url: '/pages/index/index' });
        });
      }, 500);
    } catch (err: any) {
      Taro.showToast({ title: err?.message || '登录失败', icon: 'none', duration: 1500 });
    } finally {
      setLoading(false);
    }
  }

  async function handleDevLogin() {
    if (!devOpenId) {
      Taro.showToast({ title: '请输入 openid', icon: 'none' });
      return;
    }
    setLoading(true);
    try {
      await login({ openid: devOpenId });
      await fetchProfile();
      Taro.showToast({ title: 'Dev 登录成功', icon: 'success', duration: 1200 });
    } catch (err: any) {
      Taro.showToast({ title: err?.message || 'Dev 登录失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  function handleLogout() {
    setToken(null);
    setUser(null);
    Taro.showToast({ title: '已退出', icon: 'none' });
  }

  return (
    <View style={{ minHeight: '100vh', padding: 24, backgroundColor: '#f5f6f8' }}>
      <View style={{ marginBottom: 24, textAlign: 'center' }}>
        <Image
          src={user?.avatar || 'https://dummyimage.com/120x120/EEEEEE/333333&text=Tea'}
          style={{ width: 120, height: 120, borderRadius: 60, margin: '0 auto 12px' }}
        />
        <Text style={{ fontSize: 20, fontWeight: 'bold' }}>{user ? user.nickname || '微信用户' : '请登录'}</Text>
        {user?.phone && <Text style={{ display: 'block', marginTop: 6, color: '#666' }}>{user.phone}</Text>}
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 20, marginBottom: 20 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>微信一键登录</Text>
        <Text style={{ display: 'block', marginTop: 6, color: '#666' }}>授权后即可同步订单、优惠券等数据</Text>
        <Button
          style={{ marginTop: 16, backgroundColor: '#07c160', color: '#fff' }}
          loading={loading}
          onClick={handleWxLogin}
        >
          使用微信授权登录
        </Button>
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 20, marginBottom: 20 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>开发者 OpenID 登录</Text>
        <Input
          type="text"
          placeholder="admin_openid"
          value={devOpenId}
          onInput={(e) => setDevOpenId((e.detail as any).value)}
          style={{ marginTop: 12, backgroundColor: '#f7f7f7', borderRadius: 6, padding: 8 }}
        />
        <Button style={{ marginTop: 12 }} loading={loading} onClick={handleDevLogin}>
          使用 OpenID 登录
        </Button>
      </View>

      {user && (
        <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 20 }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>已登录</Text>
          <Text style={{ display: 'block', marginTop: 8 }}>用户ID：{user.id}</Text>
          <Button style={{ marginTop: 16 }} onClick={handleLogout}>
            退出登录
          </Button>
        </View>
      )}

      {!user && !fetchingProfile && (
        <Text style={{ display: 'block', textAlign: 'center', marginTop: 12, color: '#888' }}>登录后可继续访问商城</Text>
      )}
    </View>
  );
}
