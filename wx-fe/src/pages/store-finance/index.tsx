import React, { useEffect, useState } from 'react';
import { View, Text, Button, Picker } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listStoreFinanceTransactions, type StoreFinanceTransaction, type StoreFinanceQuery } from '../../services/stores';

export default function StoreFinancePage() {
  const [storeId, setStoreId] = useState<number | undefined>(undefined);
  const [records, setRecords] = useState<StoreFinanceTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [txType, setTxType] = useState<StoreFinanceQuery['type']>(undefined);
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  useEffect(() => {
    const router = Taro.getCurrentInstance().router;
    const storeIdParam = router?.params?.store_id;
    if (storeIdParam) {
      const id = Number(storeIdParam);
      if (!Number.isNaN(id) && id > 0) {
        setStoreId(id);
        void fetchFinanceRecords(id, txType, startDate || undefined, endDate || undefined);
      }
    } else {
      Taro.showToast({ title: '缺少门店信息，请从首页进入', icon: 'none' });
    }
  }, []);

  useEffect(() => {
    if (storeId) {
      void fetchFinanceRecords(storeId, txType, startDate || undefined, endDate || undefined);
    }
  }, [storeId, txType, startDate, endDate]);

  async function fetchFinanceRecords(id: number, type?: string, start?: string, end?: string) {
    setLoading(true);
    try {
      const res = await listStoreFinanceTransactions(id, {
        page: 1,
        limit: 50,
        type,
        start,
        end,
      });
      const maybe: any = res;
      let items: StoreFinanceTransaction[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data as StoreFinanceTransaction[];
      else if (Array.isArray(maybe?.items)) items = maybe.items as StoreFinanceTransaction[];
      else if (Array.isArray(maybe)) items = maybe as StoreFinanceTransaction[];
      setRecords(items);
    } catch (e) {
      console.error('load store finance records failed', e);
      Taro.showToast({ title: '财务流水加载失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  function changeType(type?: StoreFinanceQuery['type']) {
    setTxType(type);
  }

  function handleStartChange(e: any) {
    const value = e?.detail?.value as string;
    setStartDate(value || '');
  }

  function handleEndChange(e: any) {
    const value = e?.detail?.value as string;
    setEndDate(value || '');
  }

  function renderTypeLabel(type: string) {
    if (type === 'payment') return '收款';
    if (type === 'refund') return '退款';
    if (type === 'withdraw') return '提现';
    return type;
  }

  function renderDirectionLabel(direction: string) {
    if (direction === 'in') return '收入';
    if (direction === 'out') return '支出';
    return direction;
  }

  function formatAmount(value: string | number | undefined) {
    if (value === undefined || value === null) return '-';
    const num = Number(value);
    if (Number.isNaN(num)) return String(value);
    return num.toFixed(2);
  }

  return (
    <View style={{ padding: 12 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>门店财务流水</Text>

      {/* 时间范围筛选 */}
      <View
        style={{
          marginTop: 12,
          marginBottom: 8,
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          flexWrap: 'wrap',
        }}
      >
        <Text>起始日期：</Text>
        <Picker mode="date" value={startDate} onChange={handleStartChange}>
          <View
            style={{
              padding: 4,
              borderWidth: 1,
              borderStyle: 'solid',
              borderColor: '#ddd',
              minWidth: 120,
              marginRight: 8,
            }}
          >
            <Text>{startDate || '不限'}</Text>
          </View>
        </Picker>
        <Text>结束日期：</Text>
        <Picker mode="date" value={endDate} onChange={handleEndChange}>
          <View
            style={{
              padding: 4,
              borderWidth: 1,
              borderStyle: 'solid',
              borderColor: '#ddd',
              minWidth: 120,
            }}
          >
            <Text>{endDate || '不限'}</Text>
          </View>
        </Picker>
      </View>

      {/* 类型筛选 Tab */}
      <View style={{ marginTop: 8, marginBottom: 12, display: 'flex', flexDirection: 'row' }}>
        <Button size="mini" onClick={() => changeType(undefined)}>
          全部
        </Button>
        <Button size="mini" onClick={() => changeType('payment')} style={{ marginLeft: 8 }}>
          收款
        </Button>
        <Button size="mini" onClick={() => changeType('refund')} style={{ marginLeft: 8 }}>
          退款
        </Button>
        <Button size="mini" onClick={() => changeType('withdraw')} style={{ marginLeft: 8 }}>
          提现
        </Button>
      </View>

      {loading && <Text>加载中...</Text>}
      {!loading && !records.length && <Text>暂无财务流水</Text>}
      {!loading &&
        records.map((rec) => (
          <View
            key={`${rec.type}-${rec.id}`}
            style={{
              marginBottom: 12,
              paddingBottom: 8,
              borderBottomWidth: 1,
              borderStyle: 'solid',
              borderColor: '#eee',
            }}
          >
            <View style={{ marginBottom: 4 }}>
              <Text>
                {renderTypeLabel(rec.type)} · {renderDirectionLabel(rec.direction)}
              </Text>
            </View>
            <View style={{ marginBottom: 2 }}>
              <Text>金额：¥{formatAmount(rec.amount)}</Text>
            </View>
            {rec.fee && Number(rec.fee) > 0 && (
              <View style={{ marginBottom: 2 }}>
                <Text>手续费：¥{formatAmount(rec.fee)}</Text>
              </View>
            )}
            {rec.related_no && (
              <View style={{ marginBottom: 2 }}>
                <Text>关联单号：{rec.related_no}</Text>
              </View>
            )}
            {rec.remark && (
              <View style={{ marginBottom: 2 }}>
                <Text>备注：{rec.remark}</Text>
              </View>
            )}
            {/* 解析 JSON remark 并展示标准化字段 */}
            {(() => {
              try {
                const obj = rec.remark ? JSON.parse(rec.remark) : null;
                if (!obj || typeof obj !== 'object') return null;
                const phase = obj.phase;
                const currency = obj.currency;
                const amountCents = obj.amount_cents;
                const feeCents = obj.fee_cents;
                const netCents = obj.net_cents;
                const withdrawNo = obj.withdraw_no || rec.related_no;
                const rows: Array<{ label: string; value: any }> = [];
                if (phase != null) rows.push({ label: '阶段', value: phase });
                if (withdrawNo != null) rows.push({ label: '提现单号', value: withdrawNo });
                if (currency != null) rows.push({ label: '币种', value: currency });
                if (amountCents != null) rows.push({ label: '金额(分)', value: amountCents });
                if (feeCents != null) rows.push({ label: '手续费(分)', value: feeCents });
                if (netCents != null) rows.push({ label: '实付(分)', value: netCents });
                if (!rows.length) return null;
                return (
                  <View style={{ marginTop: 2, marginBottom: 2 }}>
                    {rows.map((r) => (
                      <View key={String(r.label)} style={{ marginBottom: 2 }}>
                        <Text>
                          {r.label}：{String(r.value)}
                        </Text>
                      </View>
                    ))}
                  </View>
                );
              } catch {
                return null;
              }
            })()}
            {rec.created_at && (
              <View>
                <Text>时间：{rec.created_at}</Text>
              </View>
            )}
          </View>
        ))}
    </View>
  );
}
