import React, { useEffect, useMemo, useState } from 'react';
import { Button, Input, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { getUserInfo, updateUserInfo } from '../../services/auth';
import type { User } from '../../services/types';
import { listMembershipPackages, createMembershipOrder } from '../../services/membership';
import type { MembershipPackage } from '../../services/membership';
import { createUnifiedOrder } from '../../services/payments';

const GENDER_TEXT = ['未知', '男', '女'];

export default function MembershipPage() {
  const [user, setUser] = useState<User | null>(null);
  const [form, setForm] = useState<{ nickname: string; gender: number }>({ nickname: '', gender: 0 });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [packagesLoading, setPackagesLoading] = useState(false);
  const [packages, setPackages] = useState<MembershipPackage[]>([]);
  const [openingPackageId, setOpeningPackageId] = useState<number | null>(null);

  useEffect(() => {
    void loadProfile();
    void loadPackages();
  }, []);

  usePullDownRefresh(() => {
    void loadProfile();
    void loadPackages();
  });

  const levelText = useMemo(() => {
    if (!user) return '访客';
    const points = user.points || 0;
    if (points >= 2000) return '黑金会员';
    if (points >= 1000) return '铂金会员';
    if (points >= 500) return '黄金会员';
    if (points >= 200) return '白银会员';
    return '青铜会员';
  }, [user]);

  async function loadPackages() {
    setPackagesLoading(true);
    try {
      const res = await listMembershipPackages({ page: 1, limit: 20, type: 'membership' });
      setPackages(res.data || []);
    } catch (err) {
      console.error('load membership packages failed', err);
      Taro.showToast({ title: '加载套餐失败', icon: 'none' });
    } finally {
      setPackagesLoading(false);
    }
  }

  async function loadProfile() {
    setLoading(true);
    try {
      const info = await getUserInfo();
      setUser(info);
      setForm({ nickname: info.nickname || '', gender: info.gender ?? 0 });
    } catch (err) {
      console.error('load membership profile failed', err);
      Taro.showToast({ title: '加载失败', icon: 'none' });
    } finally {
      setLoading(false);
      Taro.stopPullDownRefresh();
    }
  }

  async function handleOpenPackage(pkg: MembershipPackage) {
    if (!pkg.id) return;
    if (openingPackageId && openingPackageId === pkg.id) {
      return;
    }
    setOpeningPackageId(pkg.id);
    try {
      const order = await createMembershipOrder({ package_id: pkg.id });
      const p = await createUnifiedOrder(order.order_id);
      await Taro.requestPayment({
			timeStamp: String((p as any).timestamp || (p as any).timeStamp),
			nonceStr: (p as any).nonce_str || (p as any).nonceStr,
			package: p.package,
			signType: 'MD5',
			paySign: p.sign,
      });

      Taro.showToast({ title: '开通成功', icon: 'success' });
      await loadProfile();
      await loadPackages();
    } catch (err) {
      console.error('open membership package failed', err);
      Taro.showToast({ title: '支付未完成，可稍后重试', icon: 'none' });
    } finally {
      setOpeningPackageId(null);
    }
  }

  function handleFieldChange(key: 'nickname' | 'gender', value: string | number) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSave() {
    if (!form.nickname.trim()) {
      Taro.showToast({ title: '请输入昵称', icon: 'none' });
      return;
    }
    setSaving(true);
    try {
      await updateUserInfo({ nickname: form.nickname.trim(), gender: form.gender });
      Taro.showToast({ title: '资料已更新', icon: 'success' });
      await loadProfile();
    } catch (err) {
      console.error('update membership info failed', err);
      Taro.showToast({ title: '更新失败', icon: 'none' });
    } finally {
      setSaving(false);
    }
  }

  return (
    <View style={{ padding: 16, backgroundColor: '#f5f6f8', minHeight: '100vh' }}>
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16 }}>
        <Text style={{ fontSize: 24, fontWeight: 'bold' }}>{levelText}</Text>
        <Text style={{ color: '#666', marginTop: 4 }}>UID：{user?.uid || '-'}</Text>
        <Text style={{ color: '#666', marginTop: 4 }}>角色：{user?.role || 'user'}</Text>
        <View style={{ display: 'flex', justifyContent: 'space-around', marginTop: 16 }}>
          <View style={{ textAlign: 'center' }}>
            <Text style={{ fontSize: 20, fontWeight: 'bold' }}>{user?.points ?? 0}</Text>
            <Text style={{ color: '#999', fontSize: 12 }}>成长积分</Text>
          </View>
          <View style={{ textAlign: 'center' }}>
            <Text style={{ fontSize: 20, fontWeight: 'bold' }}>{user?.balance ?? 0}</Text>
            <Text style={{ color: '#999', fontSize: 12 }}>账户余额</Text>
          </View>
        </View>
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>完善资料</Text>
        <View style={{ marginTop: 12 }}>
          <Text>昵称</Text>
          <Input value={form.nickname} onInput={(e) => handleFieldChange('nickname', (e.detail as any).value)} placeholder="请输入昵称" />
        </View>
        <View style={{ marginTop: 12 }}>
          <Text>性别</Text>
          <View style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            {GENDER_TEXT.map((label, idx) => (
              <Button key={label} size="mini" type={form.gender === idx ? 'primary' : undefined} onClick={() => handleFieldChange('gender', idx)}>
                {label}
              </Button>
            ))}
          </View>
        </View>
        <Button style={{ marginTop: 16 }} type="primary" loading={saving} onClick={handleSave}>
          {saving ? '保存中...' : '保存资料'}
        </Button>
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>会员权益</Text>
        <Text style={{ color: '#666', marginTop: 8 }}>· 订单积分：每成功支付 1 元累积 1 积分</Text>
        <Text style={{ color: '#666', marginTop: 4 }}>· 冻结金额：售后退款中的订单会暂时影响积分</Text>
        <Text style={{ color: '#666', marginTop: 4 }}>· 会员日：积分超过 1000 可在会员日享受折扣</Text>
      </View>

      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>可开通的会员套餐</Text>
        {packagesLoading && (
          <Text style={{ color: '#999', marginTop: 8 }}>加载中...</Text>
        )}
        {!packagesLoading && packages.length === 0 && (
          <Text style={{ color: '#999', marginTop: 8 }}>暂未配置可用的会员套餐</Text>
        )}
        {!packagesLoading && packages.length > 0 && (
          <View style={{ marginTop: 12 }}>
            {packages.map((pkg) => (
              <View
                key={pkg.id}
                style={{
                  padding: 12,
                  borderRadius: 8,
                  borderWidth: 1,
                  borderColor: '#eee',
                  marginBottom: 8,
                  display: 'flex',
                  flexDirection: 'row',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <View>
                  <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{pkg.name}</Text>
                  <View style={{ marginTop: 4 }}>
                    <Text style={{ color: '#f56c6c', fontSize: 16 }}>
                      ￥{Number(pkg.price || 0).toFixed(2)}
                    </Text>
                  </View>
                  {pkg.discount_rate && (
                    <Text style={{ color: '#666', fontSize: 12, marginTop: 4 }}>
                      会员折扣约：{(Number(pkg.discount_rate) * 100).toFixed(0)} 折
                    </Text>
                  )}
                  {pkg.tea_coin_award && (
                    <Text style={{ color: '#666', fontSize: 12, marginTop: 2 }}>
                      开通赠送茶币：{Number(pkg.tea_coin_award)}
                    </Text>
                  )}
                </View>
                <Button
                  size="mini"
                  type="primary"
                  loading={openingPackageId === pkg.id}
                  onClick={() => {
                    void handleOpenPackage(pkg);
                  }}
                >
                  {openingPackageId === pkg.id ? '开通中...' : '去开通'}
                </Button>
              </View>
            ))}
          </View>
        )}
      </View>

      <Button
        style={{ marginTop: 24 }}
        loading={loading || packagesLoading}
        onClick={() => {
          void loadProfile();
          void loadPackages();
        }}
      >
        重新加载
      </Button>

      <Button
        style={{ marginTop: 12 }}
        onClick={() => {
          Taro.navigateTo({ url: '/pages/membership-orders/index' });
        }}
      >
        查看我的会员订单
      </Button>
    </View>
  );
}
