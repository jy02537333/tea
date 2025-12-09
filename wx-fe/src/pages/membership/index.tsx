import React, { useEffect, useMemo, useState } from 'react';
import { Button, Input, Text, View } from '@tarojs/components';
import Taro, { usePullDownRefresh } from '@tarojs/taro';
import { getUserInfo, updateUserInfo } from '../../services/auth';
import type { User } from '../../services/types';

const GENDER_TEXT = ['未知', '男', '女'];

export default function MembershipPage() {
  const [user, setUser] = useState<User | null>(null);
  const [form, setForm] = useState<{ nickname: string; gender: number }>({ nickname: '', gender: 0 });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    void loadProfile();
  }, []);

  usePullDownRefresh(() => {
    void loadProfile();
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

      <Button style={{ marginTop: 24 }} loading={loading} onClick={() => loadProfile()}>
        重新加载
      </Button>
    </View>
  );
}
