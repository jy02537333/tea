import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { RechargeConfigItem, WalletSummary, WalletTx } from '../services/recharge';
import {
  creditWallet,
  debitWallet,
  freezeWallet,
  getRechargeUserWallet,
  listRechargeConfigs,
  listRechargeRecords,
  unfreezeWallet,
  upsertRechargeConfigs,
} from '../services/recharge';

const { Title, Text } = Typography;

function centsToYuan(cents?: number) {
  if (typeof cents !== 'number') return '-';
  return (cents / 100).toFixed(2);
}

function yuanToCents(yuan?: number | null) {
  if (typeof yuan !== 'number' || !Number.isFinite(yuan)) return 0;
  return Math.round(yuan * 100);
}

type RechargePackage = {
  amount_cents: number;
  bonus_tea_coin_cents?: number;
  bonus_coupon_template_ids?: number[];
  status?: number;
};

type WalletActionType = 'freeze' | 'unfreeze' | 'credit' | 'debit';

const actionLabel: Record<WalletActionType, string> = {
  freeze: '冻结金额',
  unfreeze: '解冻金额',
  credit: '充钱',
  debit: '扣钱',
};

const txTypeTag = (t: string) => {
  if (t === 'recharge' || t === 'admin_credit') return <Tag color="green">{t}</Tag>;
  if (t === 'admin_debit') return <Tag color="red">{t}</Tag>;
  if (t === 'freeze' || t === 'unfreeze') return <Tag color="gold">{t}</Tag>;
  return <Tag>{t}</Tag>;
};

export default function RechargePage() {
  const queryClient = useQueryClient();

  const [userId, setUserId] = useState<number | undefined>(undefined);
  const [filterForm] = Form.useForm<{ user_id?: number }>();

  const [actionOpen, setActionOpen] = useState(false);
  const [actionType, setActionType] = useState<WalletActionType>('credit');
  const [actionForm] = Form.useForm<{ amount_yuan?: number; remark?: string }>();

  const [pkgModalOpen, setPkgModalOpen] = useState(false);
  const [pkgEditingIndex, setPkgEditingIndex] = useState<number | null>(null);
  const [pkgForm] = Form.useForm<{
    amount_yuan?: number;
    bonus_tea_coin_yuan?: number;
    bonus_coupon_template_ids?: string;
    status?: boolean;
  }>();

  const walletQuery = useQuery({
    queryKey: ['rechargeWallet', userId ?? 0],
    queryFn: async () => {
      if (!userId) throw new Error('missing user');
      const res = await getRechargeUserWallet(userId);
      return res.wallet;
    },
    enabled: !!userId,
  });

  const recordsQuery = useQuery({
    queryKey: ['rechargeRecords', userId ?? 0],
    queryFn: () => listRechargeRecords({ user_id: userId, page: 1, limit: 50 }),
    enabled: !!userId,
  });

  const configsQuery = useQuery({
    queryKey: ['rechargeConfigs'],
    queryFn: async () => {
      const res = await listRechargeConfigs();
      return res.list;
    },
  });

  const [enabled, setEnabled] = useState(false);
  const [packages, setPackages] = useState<RechargePackage[]>([]);

  useEffect(() => {
    const list = configsQuery.data ?? [];
    const enabledCfg = list.find((x) => x.config_key === 'recharge.enabled');
    const pkgCfg = list.find((x) => x.config_key === 'recharge.packages');

    setEnabled((enabledCfg?.config_value ?? '') === '1');

    try {
      const raw = pkgCfg?.config_value ? JSON.parse(pkgCfg.config_value) : [];
      if (Array.isArray(raw)) setPackages(raw);
      else setPackages([]);
    } catch {
      setPackages([]);
    }
  }, [configsQuery.data]);

  useEffect(() => {
    if (!actionOpen) {
      actionForm.resetFields();
    }
  }, [actionOpen, actionForm]);

  useEffect(() => {
    if (!pkgModalOpen) {
      pkgForm.resetFields();
      setPkgEditingIndex(null);
    }
  }, [pkgModalOpen, pkgForm]);

  const wallet: WalletSummary | undefined = walletQuery.data;

  const walletActionMutation = useMutation({
    mutationFn: async (payload: { type: WalletActionType; amount_cents: number; remark?: string }) => {
      if (!userId) throw new Error('请先选择用户');
      if (payload.type === 'freeze') return freezeWallet(userId, payload);
      if (payload.type === 'unfreeze') return unfreezeWallet(userId, payload);
      if (payload.type === 'credit') return creditWallet(userId, payload);
      return debitWallet(userId, payload);
    },
    onSuccess: () => {
      message.success('操作成功');
      setActionOpen(false);
      queryClient.invalidateQueries({ queryKey: ['rechargeWallet'] });
      queryClient.invalidateQueries({ queryKey: ['rechargeRecords'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '操作失败');
    },
  });

  const saveConfigsMutation = useMutation({
    mutationFn: async (items: Array<{ config_key: string; config_value: string; config_type?: string; description?: string; status?: number }>) => {
      return upsertRechargeConfigs(items);
    },
    onSuccess: () => {
      message.success('配置已保存');
      queryClient.invalidateQueries({ queryKey: ['rechargeConfigs'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '保存失败');
    },
  });

  const openAction = (t: WalletActionType) => {
    setActionType(t);
    setActionOpen(true);
    actionForm.setFieldsValue({ amount_yuan: undefined, remark: '' });
  };

  const submitAction = async (values: { amount_yuan?: number; remark?: string }) => {
    const cents = yuanToCents(values.amount_yuan ?? 0);
    if (cents <= 0) {
      actionForm.setFields([{ name: 'amount_yuan', errors: ['请输入大于 0 的金额'] }]);
      return;
    }
    await walletActionMutation.mutateAsync({ type: actionType, amount_cents: cents, remark: values.remark?.trim() || undefined } as any);
  };

  const configsColumns: ColumnsType<RechargePackage> = [
    {
      title: '充值金额(元)',
      dataIndex: 'amount_cents',
      width: 140,
      render: (v: number) => centsToYuan(v),
    },
    {
      title: '赠送茶币(元)',
      dataIndex: 'bonus_tea_coin_cents',
      width: 140,
      render: (v?: number) => (typeof v === 'number' ? centsToYuan(v) : '-'),
    },
    {
      title: '赠送券模板ID',
      dataIndex: 'bonus_coupon_template_ids',
      width: 200,
      render: (v?: number[]) => (Array.isArray(v) && v.length > 0 ? v.join(',') : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v?: number) => (v === 2 ? <Tag color="red">停用</Tag> : <Tag color="green">启用</Tag>),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_, __, index) => (
        <Space size={8}>
          <Button
            type="link"
            onClick={() => {
              setPkgEditingIndex(index);
              const p = packages[index];
              pkgForm.setFieldsValue({
                amount_yuan: (p.amount_cents ?? 0) / 100,
                bonus_tea_coin_yuan: (p.bonus_tea_coin_cents ?? 0) / 100,
                bonus_coupon_template_ids: Array.isArray(p.bonus_coupon_template_ids) ? p.bonus_coupon_template_ids.join(',') : '',
                status: (p.status ?? 1) === 1,
              });
              setPkgModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            danger
            onClick={() => {
              Modal.confirm({
                title: '确认删除档位？',
                okText: '删除',
                cancelText: '取消',
                okButtonProps: { danger: true },
                onOk: () => {
                  setPackages((prev) => prev.filter((_, i) => i !== index));
                },
              });
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const recordColumns: ColumnsType<WalletTx> = [
    { title: 'ID', dataIndex: 'id', width: 90 },
    { title: '类型', dataIndex: 'type', width: 140, render: (v: string) => txTypeTag(v) },
    { title: '变动(元)', dataIndex: 'amount_cents', width: 120, render: (v: number) => centsToYuan(v) },
    {
      title: '余额(元)',
      dataIndex: 'balance_after_cents',
      width: 120,
      render: (v: number) => centsToYuan(v),
    },
    { title: '备注', dataIndex: 'remark', width: 220, render: (v?: string) => v || '-' },
    { title: '时间', dataIndex: 'created_at', width: 180, render: (v?: string) => v || '-' },
  ];

  const packagesData = useMemo(() => packages.slice().sort((a, b) => (a.amount_cents ?? 0) - (b.amount_cents ?? 0)), [packages]);

  const submitPackage = async (values: {
    amount_yuan?: number;
    bonus_tea_coin_yuan?: number;
    bonus_coupon_template_ids?: string;
    status?: boolean;
  }) => {
    const amountCents = yuanToCents(values.amount_yuan ?? 0);
    if (amountCents <= 0) {
      pkgForm.setFields([{ name: 'amount_yuan', errors: ['请输入大于 0 的金额'] }]);
      return;
    }
    const bonusTeaCoinCents = yuanToCents(values.bonus_tea_coin_yuan ?? 0);
    const ids = (values.bonus_coupon_template_ids || '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .map((s) => Number(s))
      .filter((n) => Number.isFinite(n) && n > 0);

    const next: RechargePackage = {
      amount_cents: amountCents,
      bonus_tea_coin_cents: bonusTeaCoinCents > 0 ? bonusTeaCoinCents : 0,
      bonus_coupon_template_ids: ids.length > 0 ? ids : undefined,
      status: values.status === false ? 2 : 1,
    };

    setPackages((prev) => {
      const arr = prev.slice();
      if (pkgEditingIndex === null) arr.push(next);
      else arr[pkgEditingIndex] = next;
      return arr;
    });

    setPkgModalOpen(false);
  };

  const saveConfigs = async () => {
    const items = [
      {
        config_key: 'recharge.enabled',
        config_value: enabled ? '1' : '0',
        config_type: 'string',
        description: '充值功能开关（1=启用,0=停用）',
        status: 1,
      },
      {
        config_key: 'recharge.packages',
        config_value: JSON.stringify(packagesData),
        config_type: 'json',
        description: '充值档位配置（JSON数组）',
        status: 1,
      },
    ];
    await saveConfigsMutation.mutateAsync(items);
  };

  const handleSearchUser = (values: { user_id?: number }) => {
    if (!values.user_id || values.user_id <= 0) {
      message.warning('请输入用户ID');
      return;
    }
    setUserId(values.user_id);
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Title level={4} style={{ margin: 0 }}>
        充值管理 / 充值配置
      </Title>

      <Tabs
        items={[
          {
            key: 'manage',
            label: '充值管理',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Card>
                  <Space direction="vertical" size={12} style={{ width: '100%' }}>
                    <Form layout="inline" form={filterForm} onFinish={handleSearchUser}>
                      <Form.Item name="user_id" label="用户ID">
                        <InputNumber min={1} placeholder="输入用户ID" style={{ width: 200 }} />
                      </Form.Item>
                      <Form.Item>
                        <Space>
                          <Button type="primary" htmlType="submit">
                            查询
                          </Button>
                          <Button
                            onClick={() => {
                              filterForm.resetFields();
                              setUserId(undefined);
                            }}
                          >
                            清空
                          </Button>
                        </Space>
                      </Form.Item>
                    </Form>

                    <Space wrap>
                      <Text>余额(元)：</Text>
                      <Text strong>{wallet ? centsToYuan(wallet.balance_cents) : '-'}</Text>
                      <Text>冻结(元)：</Text>
                      <Text strong>{wallet ? centsToYuan(wallet.frozen_cents) : '-'}</Text>
                    </Space>

                    <Space wrap>
                      <Button disabled={!userId} onClick={() => openAction('freeze')}>
                        冻结
                      </Button>
                      <Button disabled={!userId} onClick={() => openAction('unfreeze')}>
                        解冻
                      </Button>
                      <Button type="primary" disabled={!userId} onClick={() => openAction('credit')}>
                        充钱
                      </Button>
                      <Button danger disabled={!userId} onClick={() => openAction('debit')}>
                        扣钱
                      </Button>
                      <Button
                        disabled={!userId}
                        loading={walletQuery.isFetching || recordsQuery.isFetching}
                        onClick={() => {
                          queryClient.invalidateQueries({ queryKey: ['rechargeWallet'] });
                          queryClient.invalidateQueries({ queryKey: ['rechargeRecords'] });
                        }}
                      >
                        刷新
                      </Button>
                    </Space>
                  </Space>
                </Card>

                <Card title="充值/钱包流水（最近 50 条）" extra={userId ? <Text type="secondary">user_id={userId}</Text> : null}>
                  <Table
                    rowKey="id"
                    columns={recordColumns}
                    dataSource={recordsQuery.data?.list ?? []}
                    loading={recordsQuery.isFetching}
                    pagination={false}
                    scroll={{ x: 900 }}
                  />
                </Card>

                <Modal
                  open={actionOpen}
                  title={actionLabel[actionType]}
                  okText="确认"
                  cancelText="取消"
                  confirmLoading={walletActionMutation.isPending}
                  onCancel={() => setActionOpen(false)}
                  onOk={() => actionForm.submit()}
                >
                  <Form form={actionForm} layout="vertical" onFinish={submitAction}>
                    <Form.Item label="金额（元）" name="amount_yuan" rules={[{ required: true, message: '请输入金额' }]}>
                      <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item label="备注" name="remark">
                      <Input allowClear placeholder="可选，用于审计" />
                    </Form.Item>
                  </Form>
                </Modal>
              </Space>
            ),
          },
          {
            key: 'config',
            label: '充值配置',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Card>
                  <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
                    <Space>
                      <Text>充值功能：</Text>
                      <Switch checked={enabled} onChange={setEnabled} />
                    </Space>
                    <Space>
                      <Button
                        onClick={() => {
                          pkgForm.setFieldsValue({
                            amount_yuan: undefined,
                            bonus_tea_coin_yuan: 0,
                            bonus_coupon_template_ids: '',
                            status: true,
                          });
                          setPkgEditingIndex(null);
                          setPkgModalOpen(true);
                        }}
                      >
                        新增档位
                      </Button>
                      <Button type="primary" loading={saveConfigsMutation.isPending} onClick={saveConfigs}>
                        保存配置
                      </Button>
                    </Space>
                  </Space>
                </Card>

                <Card title="充值档位（配置‘充值多少送哪些权益’）">
                  <Table rowKey={(r) => String(r.amount_cents)} columns={configsColumns} dataSource={packagesData} pagination={false} />
                </Card>

                <Modal
                  open={pkgModalOpen}
                  title={pkgEditingIndex === null ? '新增档位' : '编辑档位'}
                  okText="确认"
                  cancelText="取消"
                  onCancel={() => setPkgModalOpen(false)}
                  onOk={() => pkgForm.submit()}
                >
                  <Form form={pkgForm} layout="vertical" onFinish={submitPackage}>
                    <Form.Item label="充值金额（元）" name="amount_yuan" rules={[{ required: true, message: '请输入充值金额' }]}>
                      <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item label="赠送茶币（元）" name="bonus_tea_coin_yuan">
                      <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item label="赠送优惠券模板ID（逗号分隔）" name="bonus_coupon_template_ids">
                      <Input allowClear placeholder="如：12,15" />
                    </Form.Item>
                    <Form.Item label="启用" name="status" valuePropName="checked">
                      <Switch />
                    </Form.Item>
                  </Form>
                </Modal>
              </Space>
            ),
          },
        ]}
      />
    </Space>
  );
}
