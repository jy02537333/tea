import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, DatePicker, Drawer, Form, Input, InputNumber, Modal, Select, Space, Spin, Table, Tag, Typography, message } from 'antd';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ColumnsType } from 'antd/es/table';
import { DownloadOutlined } from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';
import type { Store, StoreActivity, StoreActivityRegistration } from '../services/stores';
import {
  StoreActivityPayload,
  createStoreActivity,
  getStores,
  getStoreActivityRegistrations,
  listStoreActivities,
  refundStoreActivityRegistration,
  updateStoreActivity,
} from '../services/stores';
import { useAuthContext } from '../hooks/useAuth';

const ACTIVITY_TYPE_LABELS: Record<number, string> = {
  1: '限时折扣',
  2: '满减活动',
  3: '买赠活动',
};

const ACTIVITY_STATUS_COLORS: Record<number, string> = {
  1: 'green',
  2: 'red',
};

interface StoreActivityFormValues {
  name: string;
  type: number;
  time_range: [Dayjs, Dayjs];
  status: number;
  priority?: number;
  description?: string;
  rules?: string;
}

function normalizeActivityFormToPayload(values: StoreActivityFormValues): StoreActivityPayload {
  const [start, end] = values.time_range;
  return {
    name: values.name,
    type: values.type,
    start_time: start.toISOString(),
    end_time: end.toISOString(),
    status: values.status,
    priority: values.priority,
    description: values.description,
    rules: values.rules,
  };
}

function mapActivityToForm(act: StoreActivity): StoreActivityFormValues {
  return {
    name: act.name,
    type: act.type,
    time_range: [dayjs(act.start_time), dayjs(act.end_time)],
    status: act.status,
    priority: act.priority,
    description: act.description,
    rules: act.rules,
  };
}

export default function StoreActivitiesPage() {
  const { user } = useAuthContext();
  const isStoreAdmin = user?.role === 'store';
  const lockedStoreId = user?.store_id;

  const queryClient = useQueryClient();
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<number | undefined>(undefined);
  const [form] = Form.useForm<StoreActivityFormValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingActivity, setEditingActivity] = useState<StoreActivity | null>(null);
  const [registrationDrawerOpen, setRegistrationDrawerOpen] = useState(false);
  const [activeActivity, setActiveActivity] = useState<StoreActivity | null>(null);
  const [regStatusFilter, setRegStatusFilter] = useState<number | undefined>(undefined);
  const [regPage, setRegPage] = useState<{ page: number; limit: number }>({ page: 1, limit: 20 });

  const storesQuery = useQuery({
    queryKey: ['stores-for-activities'],
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
    [storesQuery.data?.list]
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

  const activitiesQuery = useQuery<StoreActivity[]>({
    queryKey: ['store-activities', selectedStoreId, statusFilter ?? 'all'],
    queryFn: () => listStoreActivities(selectedStoreId!, { status: statusFilter }),
    enabled: !!selectedStoreId,
  });

  const createMutation = useMutation({
    mutationFn: async (values: StoreActivityFormValues) => {
      if (!selectedStoreId) throw new Error('未选择门店');
      const payload = normalizeActivityFormToPayload(values);
      return createStoreActivity(selectedStoreId, payload);
    },
    onSuccess: () => {
      message.success('活动已创建');
      setModalOpen(false);
      setEditingActivity(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-activities'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '创建活动失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (values: StoreActivityFormValues) => {
      if (!selectedStoreId || !editingActivity?.id) throw new Error('未选择门店或活动');
      const payload = normalizeActivityFormToPayload(values);
      return updateStoreActivity(selectedStoreId, editingActivity.id, payload);
    },
    onSuccess: () => {
      message.success('活动已更新');
      setModalOpen(false);
      setEditingActivity(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-activities'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新活动失败');
    },
  });

  const registrationsQuery = useQuery({
    queryKey: [
      'store-activity-registrations',
      selectedStoreId,
      activeActivity?.id,
      regStatusFilter ?? 'all',
      regPage.page,
      regPage.limit,
    ],
    queryFn: () =>
      getStoreActivityRegistrations(selectedStoreId!, activeActivity!.id, {
        page: regPage.page,
        limit: regPage.limit,
        status: regStatusFilter,
      }),
    enabled: !!selectedStoreId && !!activeActivity && registrationDrawerOpen,
    placeholderData: keepPreviousData,
  });

  const refundMutation = useMutation({
    mutationFn: async (reg: StoreActivityRegistration) => {
      if (!selectedStoreId || !activeActivity) throw new Error('未选择门店或活动');
      return refundStoreActivityRegistration(selectedStoreId, activeActivity.id, reg.id, { reason: '' });
    },
    onSuccess: () => {
      message.success('已标记为已退款');
      queryClient.invalidateQueries({ queryKey: ['store-activity-registrations'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '标记退款失败');
    },
  });

  const columns: ColumnsType<StoreActivity> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '名称', dataIndex: 'name', width: 200 },
    {
      title: '类型',
      dataIndex: 'type',
      width: 120,
      render: (val?: number) => (val ? ACTIVITY_TYPE_LABELS[val] || `类型${val}` : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (val?: number) => {
        if (!val) return '-';
        const color = ACTIVITY_STATUS_COLORS[val] || 'default';
        const label = val === 1 ? '启用' : val === 2 ? '禁用' : `状态${val}`;
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: '优先级', dataIndex: 'priority', width: 100 },
    { title: '开始时间', dataIndex: 'start_time', width: 180 },
    { title: '结束时间', dataIndex: 'end_time', width: 180 },
    { title: '描述', dataIndex: 'description' },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setEditingActivity(record);
              form.setFieldsValue(mapActivityToForm(record));
              setModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            onClick={() => {
              setActiveActivity(record);
              setRegStatusFilter(undefined);
              setRegPage({ page: 1, limit: 20 });
              setRegistrationDrawerOpen(true);
            }}
          >
            报名列表
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>门店活动</Typography.Title>

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
            <Select
              allowClear
              placeholder="全部状态"
              style={{ minWidth: 160 }}
              value={statusFilter}
              onChange={(val) => setStatusFilter(val)}
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 2 },
              ]}
            />
          </Space>
          {!isStoreAdmin && storesQuery.isError && <Alert type="error" message="无法获取门店列表" showIcon />}
          {isStoreAdmin && !lockedStoreId && <Alert type="error" message="门店管理员未绑定门店，无法管理活动" showIcon />}
          {!selectedStoreId && !(isStoreAdmin && !lockedStoreId) && !storesQuery.isLoading && (
            <Typography.Text type="secondary">请选择要查看的门店，以查看该门店配置的活动。</Typography.Text>
          )}
        </Space>
      </Card>

      {selectedStoreId && (
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              onClick={() => {
                setEditingActivity(null);
                form.resetFields();
                form.setFieldsValue({
                  type: 1,
                  status: 1,
                  time_range: [dayjs(), dayjs().add(7, 'day')],
                } as Partial<StoreActivityFormValues>);
                setModalOpen(true);
              }}
            >
              新建活动
            </Button>
          </Space>
          {activitiesQuery.isLoading && <Spin />}
          {activitiesQuery.isError && <Alert type="error" message="无法获取活动列表" showIcon />}
          {activitiesQuery.data && (
            <Table<StoreActivity>
              size="small"
              rowKey="id"
              dataSource={activitiesQuery.data}
              pagination={false}
              columns={columns}
            />
          )}
        </Card>
      )}

      <Modal
        title={editingActivity ? '编辑活动' : '新建活动'}
        open={modalOpen}
        onCancel={() => {
          if (createMutation.isPending || updateMutation.isPending) return;
          setModalOpen(false);
          setEditingActivity(null);
        }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnClose
      >
        <Form<StoreActivityFormValues>
          layout="vertical"
          form={form}
          onFinish={(values) => {
            if (editingActivity) updateMutation.mutate(values);
            else createMutation.mutate(values);
          }}
        >
          <Form.Item name="name" label="活动名称" rules={[{ required: true, message: '请输入活动名称' }]}>
            <Input placeholder="请输入活动名称" />
          </Form.Item>
          <Form.Item name="type" label="活动类型" rules={[{ required: true, message: '请选择活动类型' }]}>
            <Select
              options={[
                { label: '限时折扣', value: 1 },
                { label: '满减活动', value: 2 },
                { label: '买赠活动', value: 3 },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="time_range"
            label="活动时间"
            rules={[{ required: true, message: '请选择活动起止时间' }]}
          >
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 2 },
              ]}
            />
          </Form.Item>
          <Form.Item name="priority" label="优先级">
            <InputNumber style={{ width: '100%' }} placeholder="数字越大优先级越高，可选" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="可选，补充活动说明" />
          </Form.Item>
          <Form.Item name="rules" label="规则(JSON)">
            <Input.TextArea rows={4} placeholder="可选，按需填写活动规则JSON" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={activeActivity ? `活动报名列表 - ${activeActivity.name}` : '活动报名列表'}
        width={720}
        open={registrationDrawerOpen}
        onClose={() => {
          setRegistrationDrawerOpen(false);
          setActiveActivity(null);
        }}
      >
        {!activeActivity && <Typography.Text type="secondary">请选择活动查看报名情况。</Typography.Text>}
        {activeActivity && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Space wrap>
              <Typography.Text>报名状态：</Typography.Text>
              <Select
                allowClear
                style={{ minWidth: 160 }}
                placeholder="全部状态"
                value={regStatusFilter}
                onChange={(val) => {
                  setRegStatusFilter(val);
                  setRegPage({ page: 1, limit: regPage.limit });
                }}
                options={[
                  { label: '已报名（待支付/免费）', value: 1 },
                  { label: '已支付报名', value: 2 },
                  { label: '已退款', value: 3 },
                ]}
              />
              <Button
                size="small"
                type="default"
                icon={<DownloadOutlined />}
                disabled={!registrationsQuery.data || registrationsQuery.data.list.length === 0}
                onClick={() => {
                  const data = registrationsQuery.data?.list ?? [];
                  if (!data.length) return;
                  const header = [
                    'ID',
                    '报名时间',
                    '用户ID',
                    '姓名',
                    '手机号',
                    '订单ID',
                    '订单状态',
                    '状态',
                    '报名金额',
                    '退款金额',
                    '退款原因',
                  ];
                  const rows = data.map((it) => [
                    it.id,
                    it.created_at ?? '',
                    it.user_id,
                    it.user_name ?? '',
                    it.user_phone ?? '',
                    it.order_id ?? '',
                    it.order_status ?? '',
                    it.status === 1
                      ? '已报名（待支付/免费）'
                      : it.status === 2
                      ? '已支付报名'
                      : it.status === 3
                      ? '已退款'
                      : `状态${it.status}`,
                    it.fee,
                    it.refund_amount,
                    (it.refund_reason ?? '').replace(/\n/g, ' '),
                  ]);
                  const csv = [header, ...rows]
                    .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
                    .join('\n');
                  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
                  const url = window.URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = `activity_regs_${selectedStoreId}_${activeActivity.id}_${regPage.page}.csv`;
                  document.body.appendChild(a);
                  a.click();
                  document.body.removeChild(a);
                  window.URL.revokeObjectURL(url);
                }}
              >
                导出当前页
              </Button>
            </Space>
            {registrationsQuery.isLoading && <Spin />}
            {registrationsQuery.isError && <Alert type="error" message="无法获取报名列表" showIcon />}
            {registrationsQuery.data && (
              <Table<StoreActivityRegistration>
                size="small"
                rowKey="id"
                dataSource={registrationsQuery.data.list}
                pagination={{
                  current: regPage.page,
                  pageSize: regPage.limit,
                  total: registrationsQuery.data.total,
                  showSizeChanger: true,
                  onChange: (page, pageSize) =>
                    setRegPage({ page, limit: pageSize || regPage.limit }),
                }}
                columns={[
                  { title: '报名时间', dataIndex: 'created_at', width: 160 },
                  { title: '用户ID', dataIndex: 'user_id', width: 100 },
                  { title: '姓名', dataIndex: 'user_name', width: 120 },
                  { title: '手机号', dataIndex: 'user_phone', width: 140 },
                  { title: '订单ID', dataIndex: 'order_id', width: 100 },
                  {
                    title: '订单状态',
                    dataIndex: 'order_status',
                    width: 100,
                    render: (val?: number) => {
                      if (val == null) return '-';
                      if (val === 1) return '待付款';
                      if (val === 2) return '已付款';
                      if (val === 3) return '配送中';
                      if (val === 4) return '已完成';
                      if (val === 5) return '已取消';
                      return `状态${val}`;
                    },
                  },
                  { title: '报名金额', dataIndex: 'fee', width: 100 },
                  { title: '退款金额', dataIndex: 'refund_amount', width: 100 },
                  {
                    title: '状态',
                    dataIndex: 'status',
                    width: 100,
                    render: (val: number) =>
                      val === 1
                        ? '已报名（待支付/免费）'
                        : val === 2
                        ? '已支付报名'
                        : val === 3
                        ? '已退款'
                        : `状态${val}`,
                  },
                  { title: '退款原因', dataIndex: 'refund_reason' },
                  {
                    title: '操作',
                    key: 'actions',
                    width: 120,
                    render: (_, record) => (
                      <Button
                        type="link"
                        danger
                        disabled={record.status === 3 || refundMutation.isPending}
                        onClick={() => refundMutation.mutate(record)}
                      >
                        标记退款
                      </Button>
                    ),
                  },
                ]}
              />
            )}
          </Space>
        )}
      </Drawer>
    </Space>
  );
}
