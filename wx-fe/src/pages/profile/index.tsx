import React, { useEffect, useMemo, useState } from 'react';
import { Button, Image, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { getMeSummary } from '../../services/me';
import { setToken } from '../../services/api';
import type { MeSummary, User } from '../../services/types';
import usePermission from '../../hooks/usePermission';
import { PERM_TOAST_NO_STORE_MGMT } from '../../constants/permission';

export default function ProfilePage() {
  const [summary, setSummary] = useState<MeSummary | null>(null);
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
      const s = await getMeSummary();
      setSummary(s);
      setError(null);
    } catch (err: any) {
      setSummary(null);
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

  const user: User | undefined | null = summary?.user;
  const perm = usePermission();

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

  function handleViewWallet() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/wallet/index' }).catch(() => {});
  }

  function handleViewPoints() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/points/index' }).catch(() => {});
  }

  function handleShareCenter() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/share/index' }).catch(() => {});
  }

  function handleManageAddresses() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/address/index' }).catch(() => {});
  }

  function handleServiceTickets() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/after-sale/index' }).catch(() => {});
  }

  function handleFeedback() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/feedback/index' }).catch(() => {});
  }

  function handleHelpDocs() {
    Taro.navigateTo({ url: '/pages/help/index' }).catch(() => {});
  }
  function handleSettings() {
    Taro.navigateTo({ url: '/pages/settings/index' }).catch(() => {});
  }

  function handleStoreManagement() {
    if (!ensureLoggedIn()) return;
    if (!perm.allowedStoreMgmt) {
      Taro.showToast({ title: PERM_TOAST_NO_STORE_MGMT, icon: 'none' });
      return;
    }
    // 若未选定当前门店，先到门店列表；否则跳转到账户管理
    let currentId: number | null = null;
    try {
      const v = Taro.getStorageSync('current_store_id');
      const n = Number(v);
      if (Number.isFinite(n) && n > 0) currentId = n;
    } catch (_) {}
    if (currentId) {
      Taro.navigateTo({ url: `/pages/store-accounts/index?store_id=${currentId}` }).catch(() => {
        Taro.navigateTo({ url: '/pages/stores/index' }).catch(() => {});
      });
    } else {
      Taro.navigateTo({ url: '/pages/stores/index' }).catch(() => {});
    }
  }

  function handleViewMembership() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/membership/index' }).catch(() => {});
  }

  function handleLogout() {
    setToken(null);
    setSummary(null);
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
    () => {
      const base = [
      { key: 'orders', title: '订单', desc: '查看全部订单', action: handleViewOrders },
      { key: 'wallet', title: '钱包', desc: '余额/提现/茶币', action: handleViewWallet },
      { key: 'points', title: '积分', desc: '积分与成长值', action: handleViewPoints },
      { key: 'coupons', title: '优惠券', desc: '查看可用权益', action: handleViewCoupons },
      { key: 'share', title: '分享推广', desc: '推广数据与收益', action: handleShareCenter },
      { key: 'address', title: '收货地址', desc: '管理常用地址', action: handleManageAddresses },
      {
        key: 'membership',
        title: '会员权益',
        desc: user ? '查看会员与权益' : '登录后可同步权益',
        action: handleViewMembership,
      },
      { key: 'service', title: '售后服务', desc: '进度&售后操作', action: handleServiceTickets },
      { key: 'feedback', title: '意见反馈', desc: '提交工单反馈', action: handleFeedback },
      { key: 'help', title: '帮助文档', desc: '常见问题说明', action: handleHelpDocs },
      { key: 'settings', title: '设置', desc: '账号/登录/地址/关于', action: handleSettings },
      ];
      if (perm.allowedStoreMgmt) {
        base.splice(6, 0, { key: 'store_mgmt', title: '门店管理', desc: '门店与收款账户', action: handleStoreManagement });
      }
      return base;
    },
    [handleViewOrders, handleViewWallet, handleViewPoints, handleViewCoupons, handleShareCenter, handleManageAddresses, handleViewMembership, handleServiceTickets, handleFeedback, handleHelpDocs, handleSettings, handleStoreManagement, perm.allowedStoreMgmt],
  );

  function yuan(c?: number) {
    if (typeof c !== 'number') return '--';
    return (c / 100).toFixed(2);
  }

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
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>资产概览</Text>
        <View style={{ display: 'flex', flexWrap: 'wrap', marginTop: 12 }}>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>余额（¥）</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{yuan(summary?.wallet?.balance_cents)}</Text>
          </View>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>茶币</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{summary?.wallet?.tea_coins ?? '--'}</Text>
          </View>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>积分</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{summary?.points?.balance ?? '--'}</Text>
          </View>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>可用佣金（¥）</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{yuan(summary?.share?.available_commission_cents)}</Text>
          </View>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>冻结佣金（¥）</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{yuan(summary?.share?.frozen_commission_cents)}</Text>
          </View>
          <View style={{ width: '50%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>累计佣金（¥）</Text>
            <Text style={{ fontSize: 18, fontWeight: 'bold' }}>{yuan(summary?.share?.total_commission_cents)}</Text>
          </View>
          <View style={{ width: '100%', paddingTop: 8, paddingBottom: 8 }}>
            <Text style={{ color: '#999' }}>推广规模（直属/团队）</Text>
            <Text style={{ fontSize: 16 }}>{(summary?.share?.direct_count ?? 0) + ' / ' + (summary?.share?.team_count ?? 0)}</Text>
          </View>
        </View>
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
