import React, { useEffect, useState } from 'react';
import { View, Text, Input, Button, Switch } from '@tarojs/components';
import Taro from '@tarojs/taro';
import {
  listStoreAccounts,
  createStoreAccount,
  updateStoreAccount,
  deleteStoreAccount,
  StoreBankAccount,
} from '../../services/stores';

export default function StoreAccountsPage() {
  const instance = Taro.getCurrentInstance();
  const storeIdParam = instance?.router?.params?.store_id;
  const initialStoreId = storeIdParam ? Number(storeIdParam) : NaN;

  const [storeId, setStoreId] = useState<number | null>(
    Number.isFinite(initialStoreId) && initialStoreId > 0 ? initialStoreId : null,
  );
  const [accounts, setAccounts] = useState<StoreBankAccount[]>([]);
  const [loading, setLoading] = useState(false);
  const [onlyDefault, setOnlyDefault] = useState(false);
  const [editing, setEditing] = useState<StoreBankAccount | null>(null);
  const [accountType, setAccountType] = useState('bank');
  const [accountName, setAccountName] = useState('');
  const [accountNo, setAccountNo] = useState('');
  const [bankName, setBankName] = useState('');
  const [isDefault, setIsDefault] = useState(false);

  useEffect(() => {
    if (!storeId) {
      Taro.showToast({ title: '缺少门店信息，请从首页进入', icon: 'none' });
      return;
    }
    void fetchAccounts(storeId);
  }, [storeId]);

  async function fetchAccounts(id: number) {
    setLoading(true);
    try {
      const list = await listStoreAccounts(id);
      setAccounts(list || []);
    } catch (e) {
      console.error('load store accounts failed', e);
      Taro.showToast({ title: '加载收款账户失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  function resetForm() {
    setEditing(null);
    setAccountType('bank');
    setAccountName('');
    setAccountNo('');
    setBankName('');
    setIsDefault(false);
  }

  function startCreate() {
    resetForm();
  }

  function startEdit(acc: StoreBankAccount) {
    setEditing(acc);
    setAccountType(acc.account_type || 'bank');
    setAccountName(acc.account_name || '');
    setAccountNo(acc.account_no || '');
    setBankName(acc.bank_name || '');
    setIsDefault(!!acc.is_default);
  }

  async function handleSubmit() {
    if (!storeId) {
      Taro.showToast({ title: '缺少门店信息', icon: 'none' });
      return;
    }
    if (!accountName || !accountNo) {
      Taro.showToast({ title: '请填写账户名和账号', icon: 'none' });
      return;
    }

    const payload = {
      account_type: accountType,
      account_name: accountName,
      account_no: accountNo,
      bank_name: bankName,
      is_default: isDefault,
    };

    try {
      if (editing) {
        await updateStoreAccount(storeId, editing.id, payload);
        Taro.showToast({ title: '已更新', icon: 'success' });
      } else {
        await createStoreAccount(storeId, payload);
        Taro.showToast({ title: '已新增', icon: 'success' });
      }
      resetForm();
      void fetchAccounts(storeId);
    } catch (e) {
      console.error('submit account failed', e);
      Taro.showToast({ title: '保存失败', icon: 'none' });
    }
  }

  async function handleDelete(acc: StoreBankAccount) {
    if (!storeId) return;
    const res = await Taro.showModal({
      title: '确认删除',
      content: '删除后将无法恢复，确认删除该收款账户吗？',
    });
    if (!res.confirm) return;
    try {
      await deleteStoreAccount(storeId, acc.id);
      Taro.showToast({ title: '已删除', icon: 'success' });
      if (editing && editing.id === acc.id) {
        resetForm();
      }
      void fetchAccounts(storeId);
    } catch (e) {
      console.error('delete account failed', e);
      Taro.showToast({ title: '删除失败', icon: 'none' });
    }
  }

  function renderAccountItem(acc: StoreBankAccount) {
    return (
      <View
        key={acc.id}
        style={{
          padding: 12,
          marginBottom: 8,
          borderWidth: 1,
          borderStyle: 'solid',
          borderColor: '#eee',
          borderRadius: 4,
        }}
      >
        <View style={{ marginBottom: 4, display: 'flex', flexDirection: 'row', justifyContent: 'space-between' }}>
          <Text>
            {acc.account_name} ({acc.account_type || 'bank'})
          </Text>
          {acc.is_default && <Text style={{ color: '#07c160' }}>默认</Text>}
        </View>
        <View style={{ marginBottom: 4 }}>
          <Text>账号：{acc.account_no}</Text>
        </View>
        {acc.bank_name && (
          <View style={{ marginBottom: 4 }}>
            <Text>银行/渠道：{acc.bank_name}</Text>
          </View>
        )}
        <View style={{ display: 'flex', flexDirection: 'row' }}>
          <Button size="mini" onClick={() => startEdit(acc)}>
            编辑
          </Button>
          <View style={{ width: 8 }} />
          <Button size="mini" type="warn" onClick={() => handleDelete(acc)}>
            删除
          </Button>
        </View>
      </View>
    );
  }

  return (
    <View style={{ padding: 12 }}>
      <View style={{ marginBottom: 12 }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>门店收款账户设置</Text>
      </View>

      <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <Text style={{ marginRight: 8 }}>只看默认账户</Text>
        <Switch checked={onlyDefault} onChange={(e) => setOnlyDefault(!!(e.detail as any).value)} />
      </View>

      {loading && <Text>加载中...</Text>}

      {!loading && (onlyDefault ? accounts.filter((a) => a.is_default).length === 0 : accounts.length === 0) && (
        <Text>暂未配置收款账户</Text>
      )}

      {!loading && (onlyDefault ? accounts.filter((a) => a.is_default) : accounts).length > 0 && (
        <View style={{ marginBottom: 16 }}>
          {(onlyDefault ? accounts.filter((a) => a.is_default) : accounts).map(renderAccountItem)}
        </View>
      )}

      <View style={{ marginTop: 12, paddingTop: 12, borderTopWidth: 1, borderStyle: 'solid', borderColor: '#eee' }}>
        <Text style={{ fontSize: 16, marginBottom: 8 }}>{editing ? '编辑收款账户' : '新增收款账户'}</Text>

        <View style={{ marginBottom: 8 }}>
          <Text>账户类型（bank / alipay / wechat）</Text>
          <Input
            type="text"
            value={accountType}
            onInput={(e) => setAccountType((e.detail as any).value)}
            placeholder="bank / alipay / wechat"
          />
        </View>

        <View style={{ marginBottom: 8 }}>
          <Text>账户名</Text>
          <Input
            type="text"
            value={accountName}
            onInput={(e) => setAccountName((e.detail as any).value)}
            placeholder="请输入账户开户名"
          />
        </View>

        <View style={{ marginBottom: 8 }}>
          <Text>账号 / 收款号</Text>
          <Input
            type="text"
            value={accountNo}
            onInput={(e) => setAccountNo((e.detail as any).value)}
            placeholder="银行卡号 / 支付宝账号等"
          />
        </View>

        <View style={{ marginBottom: 8 }}>
          <Text>银行名称 / 渠道名称（可选）</Text>
          <Input
            type="text"
            value={bankName}
            onInput={(e) => setBankName((e.detail as any).value)}
            placeholder="如 中国银行杭州分行 / 支付宝"
          />
        </View>

        <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
          <Text style={{ marginRight: 8 }}>设为默认账户</Text>
          <Switch checked={isDefault} onChange={(e) => setIsDefault(!!(e.detail as any).value)} />
        </View>

        <Button type="primary" onClick={handleSubmit}>
          {editing ? '保存修改' : '新增账户'}
        </Button>
        {editing && (
          <View style={{ marginTop: 8 }}>
            <Button size="mini" onClick={resetForm}>
              取消编辑
            </Button>
          </View>
        )}
      </View>
    </View>
  );
}
