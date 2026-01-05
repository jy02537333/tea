import React, { useEffect, useState } from 'react';
import { View, Text, Button, Picker } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listStoreFinanceTransactions, exportStoreFinanceTransactions, type StoreFinanceTransaction, type StoreFinanceQuery } from '../../services/stores';
import usePermission from '../../hooks/usePermission';
import { PERM_HINT_NO_STORE_FINANCE, PERM_TOAST_NO_STORE_FINANCE } from '../../constants/permission';

export default function StoreFinancePage() {
  const [storeId, setStoreId] = useState<number | undefined>(undefined);
  const [records, setRecords] = useState<StoreFinanceTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [txType, setTxType] = useState<StoreFinanceQuery['type']>(undefined);
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  const [page, setPage] = useState<number>(1);
  const [limit, setLimit] = useState<number>(20);
  const [total, setTotal] = useState<number>(0);
  const [backendMissing, setBackendMissing] = useState<boolean>(false);
  const perm = usePermission();

  useEffect(() => {
    const router = Taro.getCurrentInstance().router;
    const storeIdParam = router?.params?.store_id;
    if (storeIdParam) {
      const id = Number(storeIdParam);
      if (!Number.isNaN(id) && id > 0) {
        setStoreId(id);
      }
    } else {
      // 兜底：读取当前门店 ID
      try {
        const stored = Taro.getStorageSync('current_store_id');
        const sid = stored ? Number(stored) : NaN;
        if (!Number.isNaN(sid) && sid > 0) {
          setStoreId(sid);
          Taro.showToast({ title: '已使用当前门店', icon: 'none' });
        } else {
          Taro.showToast({ title: '缺少门店信息，请从门店列表进入', icon: 'none' });
        }
      } catch (e) {
        Taro.showToast({ title: '缺少门店信息，请从门店列表进入', icon: 'none' });
      }
    }
  }, []);

  const allowed = perm.allowedStoreFinance;

  useEffect(() => {
    if (storeId && !allowed) {
      Taro.showToast({ title: PERM_TOAST_NO_STORE_FINANCE, icon: 'none' });
    }
  }, [storeId, allowed]);

  useEffect(() => {
    if (storeId) {
      if (!allowed) {
        // 无权限则不加载记录，仅展示提示
        return;
      }
      void fetchFinanceRecords(storeId, txType, startDate || undefined, endDate || undefined, page, limit);
    }
  }, [storeId, txType, startDate, endDate, allowed, page, limit]);

  async function fetchFinanceRecords(id: number, type?: string, start?: string, end?: string, p: number = 1, l: number = limit) {
    setLoading(true);
    setBackendMissing(false);
    try {
      const res = await listStoreFinanceTransactions(id, {
        page: p,
        limit: l,
        type,
        start,
        end,
      });
      const items = Array.isArray(res?.data) ? res.data : [];
      setRecords(items);
      setTotal(Number(res?.total || items.length || 0));
    } catch (e) {
      console.error('load store finance records failed', e);
      const status = e?.response?.status || e?.status;
      if (status === 404) {
        setBackendMissing(true);
        Taro.showToast({ title: '后端路由未提供', icon: 'none' });
      } else {
        Taro.showToast({ title: '财务流水加载失败', icon: 'none' });
      }
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

  function goBackToDetail() {
    try {
      // 优先返回上一页（通常来自门店详情）
      Taro.navigateBack({ delta: 1 });
    } catch (e) {
      // H5 或无法返回时兜底跳转到门店详情
      if (storeId) {
        Taro.navigateTo({ url: `/pages/store-detail/index?store_id=${storeId}` }).catch(() => {});
      } else {
        // 再兜底：回到首页
        Taro.switchTab({ url: '/pages/index/index' }).catch(() => {});
      }
    }
  }

  function goToStoresList() {
    try {
      Taro.navigateTo({ url: '/pages/stores/index' });
    } catch (e) {
      Taro.switchTab({ url: '/pages/index/index' }).catch(() => {});
    }
  }

  function formatAmount(value: string | number | undefined) {
    if (value === undefined || value === null) return '-';
    const num = Number(value);
    if (Number.isNaN(num)) return String(value);
    return num.toFixed(2);
  }

  return (
    <View style={{ padding: 12 }}>
      <View style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Text style={{ fontSize: 18, fontWeight: 'bold' }}>门店财务流水</Text>
        <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
          <Button size="mini" onClick={goBackToDetail}>返回门店详情</Button>
          <Button size="mini" style={{ marginLeft: 8 }} onClick={goToStoresList}>回到门店列表</Button>
        </View>
      </View>
      <Text style={{ color: '#999', marginTop: 6, display: 'block' }}>提示：可通过右上角返回门店详情或回到门店列表</Text>
      {!allowed && (
        <Text style={{ color: '#999', marginTop: 8 }}>{PERM_HINT_NO_STORE_FINANCE}</Text>
      )}
      {allowed && backendMissing && (
        <Text style={{ color: '#999', marginTop: 8 }}>后端路由未提供（请升级后端版本以提供 /api/v1/stores/:id/finance/transactions）</Text>
      )}

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
        <Picker mode="date" value={startDate} onChange={handleStartChange} disabled={!allowed}>
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
        <Picker mode="date" value={endDate} onChange={handleEndChange} disabled={!allowed}>
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
        <Button size="mini" disabled={!allowed} onClick={() => changeType(undefined)}>
          全部
        </Button>
        <Button size="mini" disabled={!allowed} onClick={() => changeType('payment')} style={{ marginLeft: 8 }}>
          收款
        </Button>
        <Button size="mini" disabled={!allowed} onClick={() => changeType('refund')} style={{ marginLeft: 8 }}>
          退款
        </Button>
        <Button size="mini" disabled={!allowed} onClick={() => changeType('withdraw')} style={{ marginLeft: 8 }}>
          提现
        </Button>
        <Button
          size="mini"
          disabled={!allowed || backendMissing}
          style={{ marginLeft: 8 }}
          onClick={async () => {
            if (!storeId) return;
            try {
              await exportStoreFinanceTransactions(storeId, {
                type: txType,
                start: startDate || undefined,
                end: endDate || undefined,
              });
              Taro.showToast({ title: '已发起导出', icon: 'none' });
            } catch (e) {
              console.error('export finance failed', e);
              Taro.showToast({ title: '导出失败', icon: 'none' });
            }
          }}
        >
          导出
        </Button>
      </View>

      {loading && <Text>加载中...</Text>}
      {!loading && allowed && backendMissing && (
        <Text style={{ color: '#999', marginTop: 8 }}>后端路由未提供，无法展示财务流水</Text>
      )}
      {!loading && allowed && !backendMissing && !records.length && <Text>暂无财务流水</Text>}
      {!loading && allowed && !backendMissing &&
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

      {/* 分页控制 */}
      {!loading && allowed && !backendMissing && (
        <View style={{ marginTop: 8, display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
          <Text>
            共 {total} 条 · 每页 {limit} 条 · 第 {page} 页
          </Text>
          <Button
            size="mini"
            disabled={!allowed || page <= 1}
            style={{ marginLeft: 8 }}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            上一页
          </Button>
          <Button
            size="mini"
            disabled={!allowed || records.length < limit || page * limit >= total}
            style={{ marginLeft: 8 }}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </Button>
        </View>
      )}

      {/* 底部返回入口，适配长列表滚动 */}
      {!loading && (
        <View style={{ marginTop: 16, display: 'flex', justifyContent: 'center' }}>
          <Button size="mini" onClick={goToStoresList}>回到门店列表</Button>
        </View>
      )}
    </View>
  );
}
