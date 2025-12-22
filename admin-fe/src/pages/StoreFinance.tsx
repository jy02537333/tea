import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Descriptions,
  Divider,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Spin,
  Table,
  Typography,
  message,
} from 'antd';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { DownloadOutlined } from '@ant-design/icons';
import {
  Store,
  StoreWalletSummary,
  StoreWithdrawRecord,
  getStores,
  getStoreWallet,
  getStoreWithdraws,
  applyStoreWithdraw,
} from '../services/stores';
import { WITHDRAW_STATUS_LABELS } from '../constants/withdraw';

export default function StoreFinancePage() {
  const queryClient = useQueryClient();
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);
  const [walletPage, setWalletPage] = useState({ page: 1, limit: 20 });
  const [walletStatus, setWalletStatus] = useState<number | undefined>(undefined);
  const [withdrawForm] = Form.useForm<{ amount: number; remark?: string }>();

  const storesQuery = useQuery({
    queryKey: ['stores-for-finance'],
    queryFn: () => getStores({ page: 1, limit: 200 }),
    placeholderData: keepPreviousData,
  });

  const storeOptions = useMemo(
    () =>
      (storesQuery.data?.list ?? []).map((s: Store) => ({
        label: s.name || `门店 #${s.id}`,
        value: s.id,
      })),
    [storesQuery.data?.list]
  );

  useEffect(() => {
    if (!selectedStoreId && storeOptions.length === 1) {
      setSelectedStoreId(storeOptions[0].value);
    }
  }, [selectedStoreId, storeOptions]);

  const walletQuery = useQuery<StoreWalletSummary | null>({
    queryKey: ['store-wallet', selectedStoreId],
    queryFn: () => getStoreWallet(selectedStoreId!),
    enabled: !!selectedStoreId,
  });

  const withdrawListQuery = useQuery({
    queryKey: ['store-withdraws', selectedStoreId, walletPage.page, walletPage.limit, walletStatus ?? 'all'],
    queryFn: () =>
      getStoreWithdraws(selectedStoreId!, {
        page: walletPage.page,
        limit: walletPage.limit,
        status: walletStatus,
      }),
    enabled: !!selectedStoreId,
  });

  const withdrawMutation = useMutation({
    mutationFn: async (payload: { amount: number; remark?: string }) => {
      if (!selectedStoreId) throw new Error('未选择门店');
      return applyStoreWithdraw(selectedStoreId, { amount: payload.amount, remark: payload.remark });
    },
    onSuccess: () => {
      message.success('提现申请已提交');
      queryClient.invalidateQueries({ queryKey: ['store-wallet', selectedStoreId] });
      queryClient.invalidateQueries({ queryKey: ['store-withdraws', selectedStoreId] });
      withdrawForm.resetFields();
    },
    onError: (error: any) => {
      message.error(error?.message || '提现申请失败');
    },
  });

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>门店财务提现</Typography.Title>

      <Card>
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Space wrap>
            <Typography.Text>选择门店：</Typography.Text>
            <Select
              style={{ minWidth: 240 }}
              placeholder="请选择门店"
              loading={storesQuery.isLoading}
              options={storeOptions}
              value={selectedStoreId}
              onChange={(val) => {
                setSelectedStoreId(val);
                setWalletPage({ page: 1, limit: 20 });
                setWalletStatus(undefined);
              }}
              allowClear
            />
          </Space>
          {storesQuery.isError && <Alert type="error" message="无法获取门店列表" showIcon />}
          {!selectedStoreId && !storesQuery.isLoading && (
            <Typography.Text type="secondary">请选择要查看的门店，以查看钱包余额和提现记录。</Typography.Text>
          )}
        </Space>
      </Card>

      {selectedStoreId && (
        <Card>
          {walletQuery.isLoading && <Spin />}
          {walletQuery.isError && <Alert type="error" message="无法获取钱包信息" showIcon />}
              {withdrawListQuery.data && (
                <Table<StoreWithdrawRecord>
              <Descriptions column={3} size="small" bordered>
                <Descriptions.Item label="门店ID">{walletQuery.data.store_id}</Descriptions.Item>
                <Descriptions.Item label="总收入">￥{Number(walletQuery.data.total_paid).toFixed(2)}</Descriptions.Item>
                <Descriptions.Item label="总退款">￥{Number(walletQuery.data.total_refunded).toFixed(2)}</Descriptions.Item>
                <Descriptions.Item label="总提现">￥{Number(walletQuery.data.total_withdrawn).toFixed(2)}</Descriptions.Item>
                <Descriptions.Item label="可用余额">￥{Number(walletQuery.data.available).toFixed(2)}</Descriptions.Item>
              </Descriptions>
              <Divider />
            </>
          )}
                  columns={[
                    // 解析 remark JSON 的辅助渲染
                    // 仅在 remark 为合法 JSON 时展示解析列
                    // 否则该列返回 '-'
                    {
                      title: '时间',
                      dataIndex: 'created_at',
                      width: 160,
                    },
          <Space align="start" style={{ width: '100%' }} size={24}>
            <Form
              layout="vertical"
              form={withdrawForm}
              style={{ width: 260 }}
              onFinish={(values) => withdrawMutation.mutate(values)}
            >
              <Form.Item label="提现状态筛选">
                <Select
                  allowClear
                    { title: '备注', dataIndex: 'remark' },
                    {
                      title: '阶段',
                      key: 'remark_phase',
                      width: 100,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.phase ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '币种',
                      key: 'remark_currency',
                      width: 80,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.currency ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '金额(分)',
                      key: 'remark_amount_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.amount_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '手续费(分)',
                      key: 'remark_fee_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.fee_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '实付(分)',
                      key: 'remark_net_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.net_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                  value={walletStatus}
                  onChange={(val) => {
                    setWalletStatus(val);
                    setWalletPage((prev) => ({ ...prev, page: 1 }));
                  }}
                  options={[
                    { label: '申请中', value: 1 },
                    { label: '处理中', value: 2 },
                    { label: '已完成', value: 3 },
                    { label: '已拒绝', value: 4 },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label="提现金额"
                name="amount"
                rules={[{ required: true, message: '请输入提现金额' }]}
              >
                <InputNumber
                  min={0.01}
                  max={walletQuery.data ? Number(walletQuery.data.available) : undefined}
                  step={0.01}
                  style={{ width: '100%' }}
                  placeholder="单位：元"
                />
              </Form.Item>
              <Form.Item label="备注" name="remark">
                <Input.TextArea rows={3} maxLength={200} showCount />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={withdrawMutation.isPending}
                  disabled={
                    walletQuery.isLoading ||
                    !!walletQuery.error ||
                    !walletQuery.data ||
                    Number(walletQuery.data.available) <= 0
                  }
                >
                  提交提现申请
                </Button>
              </Form.Item>
            </Form>

            <div style={{ flex: 1 }}>
              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <Divider orientation="left" style={{ margin: 0 }}>
                  提现记录
                </Divider>
                <Button
                  size="small"
                  type="default"
                  icon={<DownloadOutlined />}
                  disabled={!withdrawListQuery.data || withdrawListQuery.data.list.length === 0}
                  onClick={() => {
                    const data = withdrawListQuery.data?.list ?? [];
                    if (!data.length) return;
                    const header = ['ID', '时间', '单号', '金额', '手续费', '实付金额', '状态', '备注', '阶段', '币种', '金额(分)', '手续费(分)', '实付(分)'];
                    const rows = data.map((it) => {
                      let phase = '-', currency = '-', amountCents: string | number = '-', feeCents: string | number = '-', netCents: string | number = '-';
                      try {
                        const obj = it.remark ? JSON.parse(it.remark) : null;
                        if (obj) {
                          phase = obj.phase ?? phase;
                          currency = obj.currency ?? currency;
                          amountCents = obj.amount_cents ?? amountCents;
                          feeCents = obj.fee_cents ?? feeCents;
                          netCents = obj.net_cents ?? netCents;
                        }
                      } catch {}
                      return [
                        it.id,
                        it.created_at ?? '',
                        it.withdraw_no,
                        it.amount,
                        it.fee,
                        it.actual_amount,
                        WITHDRAW_STATUS_LABELS[it.status] ?? `状态${it.status}`,
                        (it.remark ?? '').replace(/\n/g, ' '),
                        phase,
                        currency,
                        amountCents,
                        feeCents,
                        netCents,
                      ];
                    });
                    const csv = [header, ...rows]
                      .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
                      .join('\n');
                    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
                    const url = window.URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `store_withdraws_${selectedStoreId}_${walletPage.page}.csv`;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    window.URL.revokeObjectURL(url);
                  }}
                >
                  导出当前页
                </Button>
              </Space>

              {withdrawListQuery.isLoading && <Spin />}
              {withdrawListQuery.isError && <Alert type="error" message="无法获取提现记录" showIcon />}
              {withdrawListQuery.data && (
                <Table<StoreWithdrawRecord>
                  size="small"
                  rowKey="id"
                  dataSource={withdrawListQuery.data.list}
                  pagination={{
                    current: walletPage.page,
                    pageSize: walletPage.limit,
                    total: withdrawListQuery.data.total,
                    showSizeChanger: true,
                    onChange: (page, pageSize) => setWalletPage({ page, limit: pageSize || walletPage.limit }),
                  }}
                  columns={[
                    { title: '时间', dataIndex: 'created_at', width: 160 },
                    { title: '单号', dataIndex: 'withdraw_no', width: 160 },
                    { title: '金额', dataIndex: 'amount', width: 100 },
                    { title: '手续费', dataIndex: 'fee', width: 100 },
                    { title: '实付金额', dataIndex: 'actual_amount', width: 100 },
                    {
                      title: '状态',
                      dataIndex: 'status',
                      width: 100,
                      render: (val: number) => WITHDRAW_STATUS_LABELS[val] ?? `状态${val}`,
                    },
                    { title: '备注', dataIndex: 'remark' },
                    {
                      title: '阶段',
                      key: 'remark_phase',
                      width: 100,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.phase ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '币种',
                      key: 'remark_currency',
                      width: 80,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.currency ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '金额(分)',
                      key: 'remark_amount_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.amount_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '手续费(分)',
                      key: 'remark_fee_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.fee_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                    {
                      title: '实付(分)',
                      key: 'remark_net_cents',
                      width: 120,
                      render: (_, record) => {
                        try {
                          const obj = record.remark ? JSON.parse(record.remark) : null;
                          return obj?.net_cents ?? '-';
                        } catch { return '-'; }
                      },
                    },
                  ]}
                />
              )}
            </div>
          </Space>
        </Card>
      )}
    </Space>
  );
}
