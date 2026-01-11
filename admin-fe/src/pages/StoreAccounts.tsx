import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Form, Input, Modal, Select, Space, Spin, Switch, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { Store } from '../services/stores';
import {
  StoreBankAccount,
  createStoreAccount,
  deleteStoreAccount,
  getStores,
  listStoreAccounts,
  updateStoreAccount,
} from '../services/stores';
import { useAuthContext } from '../hooks/useAuth';

interface StoreAccountFormValues {
  account_type?: string;
  account_name: string;
  account_no: string;
  bank_name?: string;
  is_default?: boolean;
}

const ACCOUNT_TYPE_LABELS: Record<string, string> = {
  bank: '银行卡',
  alipay: '支付宝',
  wechat: '微信收款',
};

export default function StoreAccountsPage() {
  const { user } = useAuthContext();
  const isStoreAdmin = user?.role === 'store';
  const lockedStoreId = user?.store_id;

  const queryClient = useQueryClient();
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);
  const [form] = Form.useForm<StoreAccountFormValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingAccount, setEditingAccount] = useState<StoreBankAccount | null>(null);
  const [onlyDefault, setOnlyDefault] = useState(false);

  const storesQuery = useQuery({
    queryKey: ['stores-for-accounts'],
    queryFn: () => getStores({ page: 1, limit: 200 }),
    placeholderData: keepPreviousData,
    enabled: !isStoreAdmin,
  });

  const storeOptions = useMemo(
    () =>
      (storesQuery.data?.list ?? []).map((s: Store) => ({
        label: s.name || `门店 #${s.id}`,
        value: s.id,
      })),
    [storesQuery.data?.list],
  );

  useEffect(() => {
    if (isStoreAdmin) {
      if (lockedStoreId && lockedStoreId !== selectedStoreId) {
        setSelectedStoreId(lockedStoreId);
      }
      return;
    }
    if (!selectedStoreId && storeOptions.length === 1) {
      setSelectedStoreId(storeOptions[0].value);
    }
  }, [isStoreAdmin, lockedStoreId, selectedStoreId, storeOptions]);

  const accountsQuery = useQuery<StoreBankAccount[]>({
    queryKey: ['store-accounts', selectedStoreId],
    queryFn: () => listStoreAccounts(selectedStoreId!),
    enabled: !!selectedStoreId,
  });

  const filteredAccounts = useMemo(
    () => (onlyDefault ? (accountsQuery.data || []).filter((a) => a.is_default) : accountsQuery.data || []),
    [onlyDefault, accountsQuery.data],
  );

  const createMutation = useMutation({
    mutationFn: async (values: StoreAccountFormValues) => {
      if (!selectedStoreId) throw new Error('未选择门店');
      return createStoreAccount(selectedStoreId, values);
    },
    onSuccess: () => {
      message.success('收款账户已创建');
      setModalOpen(false);
      setEditingAccount(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-accounts'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '创建收款账户失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (values: StoreAccountFormValues) => {
      if (!selectedStoreId || !editingAccount?.id) throw new Error('未选择门店或账户');
      return updateStoreAccount(selectedStoreId, editingAccount.id, values);
    },
    onSuccess: () => {
      message.success('收款账户已更新');
      setModalOpen(false);
      setEditingAccount(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-accounts'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新收款账户失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (account: StoreBankAccount) => {
      if (!selectedStoreId) throw new Error('未选择门店');
      return deleteStoreAccount(selectedStoreId, account.id);
    },
    onSuccess: () => {
      message.success('已删除收款账户');
      queryClient.invalidateQueries({ queryKey: ['store-accounts'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '删除收款账户失败');
    },
  });

  const columns: ColumnsType<StoreBankAccount> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '账户名', dataIndex: 'account_name', width: 200 },
    {
      title: '账户类型',
      dataIndex: 'account_type',
      width: 140,
      render: (val?: string) => (val ? ACCOUNT_TYPE_LABELS[val] || val : '-'),
    },
    { title: '账号/收款号', dataIndex: 'account_no', width: 260 },
    { title: '银行/渠道', dataIndex: 'bank_name', width: 200 },
    {
      title: '是否默认',
      dataIndex: 'is_default',
      width: 120,
      render: (val?: boolean) => (val ? <Tag color="green">默认</Tag> : <Tag>否</Tag>),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setEditingAccount(record);
              form.setFieldsValue({
                account_type: record.account_type,
                account_name: record.account_name,
                account_no: record.account_no,
                bank_name: record.bank_name,
                is_default: record.is_default,
              });
              setModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            danger
            loading={deleteMutation.isPending}
            onClick={() => {
              Modal.confirm({
                title: '确认删除该收款账户？',
                content: '删除后将无法恢复，请确认已经不再使用该账户。',
                onOk: () => deleteMutation.mutate(record),
              });
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>门店收款账户管理</Typography.Title>

      <Card>
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Space wrap>
            <Typography.Text>门店：</Typography.Text>
            {isStoreAdmin ? (
              <Typography.Text type={lockedStoreId ? undefined : 'danger'}>
                {lockedStoreId ? `已锁定门店 #${lockedStoreId}` : '未绑定门店（store_admins）'}
              </Typography.Text>
            ) : (
              <Select
                style={{ minWidth: 240 }}
                placeholder="请选择门店"
                loading={storesQuery.isLoading}
                options={storeOptions}
                value={selectedStoreId}
                onChange={(val) => setSelectedStoreId(val)}
                allowClear
              />
            )}
          </Space>
          {!isStoreAdmin && storesQuery.isError && <Alert type="error" message="无法获取门店列表" showIcon />}
          {isStoreAdmin && !lockedStoreId && <Alert type="error" message="门店管理员未绑定门店，无法管理收款账户" showIcon />}
          {!selectedStoreId && !(isStoreAdmin && !lockedStoreId) && !storesQuery.isLoading && (
            <Typography.Text type="secondary">请选择要管理的门店，以查看和维护其收款账户。</Typography.Text>
          )}
        </Space>
      </Card>

      {selectedStoreId && (
        <Card>
          <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
            <Space>
              <Button
                type="primary"
                onClick={() => {
                  setEditingAccount(null);
                  form.resetFields();
                  form.setFieldsValue({ account_type: 'bank', is_default: false });
                  setModalOpen(true);
                }}
              >
                新建收款账户
              </Button>
            </Space>
            <Space>
              <Typography.Text>只看默认账户</Typography.Text>
              <Switch checked={onlyDefault} onChange={(checked) => setOnlyDefault(checked)} />
            </Space>
          </Space>

          {accountsQuery.isLoading && <Spin />}
          {accountsQuery.isError && <Alert type="error" message="无法获取收款账户列表" showIcon />}
          {accountsQuery.data && (
            <Table<StoreBankAccount>
              size="small"
              rowKey="id"
              dataSource={filteredAccounts}
              pagination={false}
              columns={columns}
            />
          )}
        </Card>
      )}

      <Modal
        title={editingAccount ? '编辑收款账户' : '新建收款账户'}
        open={modalOpen}
        onCancel={() => {
          if (createMutation.isPending || updateMutation.isPending) return;
          setModalOpen(false);
          setEditingAccount(null);
        }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnClose
      >
        <Form<StoreAccountFormValues>
          layout="vertical"
          form={form}
          initialValues={{ account_type: 'bank', is_default: false }}
          onFinish={(values) => {
            if (editingAccount) updateMutation.mutate(values);
            else createMutation.mutate(values);
          }}
        >
          <Form.Item label="账户类型" name="account_type">
            <Select
              allowClear
              options={[
                { label: '银行卡', value: 'bank' },
                { label: '支付宝', value: 'alipay' },
                { label: '微信收款', value: 'wechat' },
              ]}
              placeholder="选择账户类型"
            />
          </Form.Item>
          <Form.Item
            label="账户名"
            name="account_name"
            rules={[{ required: true, message: '请输入账户名' }]}
          >
            <Input maxLength={128} placeholder="开户名，例如：杭州茶心阁门店" />
          </Form.Item>
          <Form.Item
            label="账号/收款号"
            name="account_no"
            rules={[{ required: true, message: '请输入账号或收款号' }]}
          >
            <Input maxLength={128} placeholder="银行卡号 / 支付宝账号 / 微信收款码标识等" />
          </Form.Item>
          <Form.Item label="银行名称 / 渠道名称" name="bank_name">
            <Input maxLength={128} placeholder="如：中国银行杭州分行 / 支付宝 / 微信支付" />
          </Form.Item>
          <Form.Item label="设为默认账户" name="is_default" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
