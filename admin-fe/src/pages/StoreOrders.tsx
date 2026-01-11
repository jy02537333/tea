import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AdminOrder, getAdminStoreOrders, getStoreOrders, postOrderAction } from '../services/orders';
import { useAuthContext } from '../hooks/useAuth';

const { Title, Text } = Typography;

const ORDER_STATUS_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '待付款', value: 1 },
  { label: '已付款', value: 2 },
  { label: '配送中', value: 3 },
  { label: '已完成', value: 4 },
  { label: '已取消', value: 5 },
];

const ORDER_STATUS_MAP: Record<number, { label: string; color: string }> = {
  1: { label: '待付款', color: 'orange' },
  2: { label: '已付款', color: 'blue' },
  3: { label: '配送中', color: 'purple' },
  4: { label: '已完成', color: 'green' },
  5: { label: '已取消', color: 'red' },
};

const PAY_STATUS_MAP: Record<number, { label: string; color: string }> = {
  1: { label: '未支付', color: 'default' },
  2: { label: '已支付', color: 'green' },
  3: { label: '退款中', color: 'orange' },
  4: { label: '已退款', color: 'red' },
};

const formatCurrency = (val?: number | string) => `￥${Number(val ?? 0).toFixed(2)}`;

interface FilterValues {
  status?: number;
  order_no?: string;
}

export default function StoreOrdersPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();
  const storeId = Number(params.id || 0);
  const queryClient = useQueryClient();
  const { hasPermission, user } = useAuthContext();

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, pageSize: 20 });
  const [filterForm] = Form.useForm<FilterValues>();

  // 支持从 Header 搜索跳转：/stores/:id/orders?orderNo=...
  useEffect(() => {
    const sp = new URLSearchParams(location.search);
    const orderNo = sp.get('orderNo');
    if (!orderNo) return;
    const keyword = String(orderNo).trim();
    if (!keyword) return;
    filterForm.setFieldsValue({ order_no: keyword });
    setFilters((prev) => ({ ...prev, order_no: keyword }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.search]);

  const listParams = useMemo(
    () => ({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: filters.status,
    }),
    [filters.status, pagination.page, pagination.pageSize]
  );

  const ordersQuery = useQuery({
    queryKey: ['storeOrders', storeId, listParams.page, listParams.page_size, listParams.status ?? 'all'],
    queryFn: () => (user?.role === 'store' ? getStoreOrders(storeId, listParams) : getAdminStoreOrders(storeId, listParams)),
    enabled: storeId > 0,
    placeholderData: keepPreviousData,
  });

  const ACTION_REASON_MAP: Record<string, string> = {
    refund: '后台立即退款',
    'refund/start': '后台标记退款中',
    'refund/confirm': '后台确认退款完成',
  };

  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [cancelTargetId, setCancelTargetId] = useState<number | null>(null);
  const [cancelReason, setCancelReason] = useState('');

  const [reasonModalOpen, setReasonModalOpen] = useState(false);
  const [reasonTarget, setReasonTarget] = useState<{ id: number; action: string } | null>(null);
  const [reasonText, setReasonText] = useState('');

  // 按 PRD：订单中的桌号不可编辑，移除相关弹窗与操作

  const actionMutation = useMutation({
    mutationFn: async ({ id, action, reason }: { id: number; action: string; reason?: string }) => {
      const body = reason ? { reason } : undefined;
      await postOrderAction(id, action, body);
    },
    onSuccess: async () => {
      message.success('操作成功');
      await queryClient.invalidateQueries({ queryKey: ['storeOrders', storeId] });
      await queryClient.invalidateQueries({ queryKey: ['adminOrders'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '操作失败');
    },
  });

  const tableData = useMemo(() => {
    const list = ordersQuery.data?.list ?? [];
    if (!filters.order_no) return list;
    const keyword = filters.order_no.trim().toLowerCase();
    if (!keyword) return list;
    return list.filter((order) => order.order_no.toLowerCase().includes(keyword));
  }, [filters.order_no, ordersQuery.data?.list]);

  const columns: ColumnsType<AdminOrder> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '订单号', dataIndex: 'order_no', width: 200 },
    { title: '用户ID', dataIndex: 'user_id', width: 100 },
    {
      title: '金额',
      dataIndex: 'pay_amount',
      width: 120,
      render: (val: number) => formatCurrency(val),
    },
    {
      title: '订单状态',
      dataIndex: 'status',
      width: 140,
      render: (status: number) => {
        const meta = ORDER_STATUS_MAP[status];
        return meta ? <Tag color={meta.color}>{meta.label}</Tag> : '-';
      },
    },
    {
      title: '支付状态',
      dataIndex: 'pay_status',
      width: 140,
      render: (status: number) => {
        const meta = PAY_STATUS_MAP[status];
        return meta ? <Tag color={meta.color}>{meta.label}</Tag> : '-';
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 200,
      render: (val?: string) => (val ? new Date(val).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 420,
      render: (_, record) => (
        <Space wrap>
                    {/* 订单桌号不可编辑（按 PRD） */}
          <Button type="link" onClick={() => navigate(`/orders?orderId=${record.id}&storeId=${storeId}`)}>
            在订单操作区打开
          </Button>
          {record.status === 2 && hasPermission('order:deliver') && (
            <Popconfirm title="确认发货该订单？" onConfirm={() => actionMutation.mutate({ id: record.id, action: 'deliver' })}>
              <Button type="link" disabled={actionMutation.isPending}>发货</Button>
            </Popconfirm>
          )}
          {record.status === 3 && hasPermission('order:complete') && (
            <Popconfirm title="确认标记订单完成？" onConfirm={() => actionMutation.mutate({ id: record.id, action: 'complete' })}>
              <Button type="link" style={{ fontWeight: 600 }} disabled={actionMutation.isPending}>完成</Button>
            </Popconfirm>
          )}
          {hasPermission('order:refund') && record.pay_status === 2 && record.status >= 2 && record.status < 5 && (
            <Popconfirm
              title="确认标记该订单为退款中？"
              onConfirm={() => {
                setReasonTarget({ id: record.id, action: 'refund/start' });
                setReasonText(ACTION_REASON_MAP['refund/start']);
                setReasonModalOpen(true);
              }}
            >
              <Button type="link" disabled={actionMutation.isPending}>标记退款中</Button>
            </Popconfirm>
          )}
          {hasPermission('order:refund') && record.pay_status === 3 && (
            <Popconfirm
              title="确认退款已完成？"
              onConfirm={() => {
                setReasonTarget({ id: record.id, action: 'refund/confirm' });
                setReasonText(ACTION_REASON_MAP['refund/confirm']);
                setReasonModalOpen(true);
              }}
            >
              <Button type="link" disabled={actionMutation.isPending}>确认退款完成</Button>
            </Popconfirm>
          )}
          {hasPermission('order:refund') && record.pay_status === 2 && (
            <Popconfirm
              title="确认执行立即退款？"
              onConfirm={() => {
                setReasonTarget({ id: record.id, action: 'refund' });
                setReasonText(ACTION_REASON_MAP['refund']);
                setReasonModalOpen(true);
              }}
            >
              <Button type="link" danger disabled={actionMutation.isPending}>立即退款</Button>
            </Popconfirm>
          )}
          {record.status > 0 && record.status < 4 && hasPermission('order:cancel') && (
            <Popconfirm
              title="确认强制取消该订单？"
              okButtonProps={{ 'data-testid': 'store-cancel-popconfirm-ok' } as any}
              onConfirm={() => {
                setCancelTargetId(record.id);
                setCancelReason('');
                setCancelModalOpen(true);
              }}
            >
              <Button type="link" danger disabled={actionMutation.isPending}>管理员取消</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  const handleFilter = (values: FilterValues) => {
    setFilters(values);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, pageSize: 20 });
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Button type="link" onClick={() => navigate('/stores')}>
        ← 返回门店列表
      </Button>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          门店订单列表 {storeId ? <Text type="secondary">(门店ID: {storeId})</Text> : null}
        </Title>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="order_no" label="订单号">
            <Input allowClear placeholder="支持模糊查询" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select
              allowClear
              style={{ width: 140 }}
              placeholder="全部状态"
              options={ORDER_STATUS_OPTIONS.filter((opt) => opt.value !== undefined)}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={ordersQuery.isFetching}>
                查询
              </Button>
              <Button onClick={handleReset} disabled={ordersQuery.isFetching}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Space>

      {storeId === 0 && <Alert type="error" message="缺少有效的门店ID" showIcon />}

      <Table<AdminOrder>
        rowKey="id"
        dataSource={tableData}
        loading={ordersQuery.isFetching}
        columns={columns}
        pagination={{
          current: ordersQuery.data?.page ?? pagination.page,
          pageSize: ordersQuery.data?.limit ?? pagination.pageSize,
          total: ordersQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (page, pageSize) => {
            setPagination({ page, pageSize });
          },
        }}
      />

      <Modal
        title={cancelTargetId ? `取消订单 · #${cancelTargetId}` : '取消订单'}
        open={cancelModalOpen}
        okText="确认取消"
        cancelText="再想想"
        confirmLoading={actionMutation.isPending}
        onCancel={() => {
          setCancelModalOpen(false);
          setCancelTargetId(null);
        }}
        onOk={async () => {
          if (!cancelTargetId) return;
          if (!cancelReason.trim()) {
            message.warning('请输入取消原因');
            return;
          }
          await actionMutation.mutateAsync({ id: cancelTargetId, action: 'admin-cancel', reason: cancelReason.trim() });
          setCancelModalOpen(false);
          setCancelTargetId(null);
          setCancelReason('');
        }}
        destroyOnClose
      >
        <Form layout="vertical">
          <Form.Item label="取消原因" required>
            <Input.TextArea rows={3} value={cancelReason} onChange={(e) => setCancelReason(e.target.value)} placeholder="请填写此次取消的原因（必填）" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 订单桌号不可编辑，移除相关弹窗 */}

      <Modal
        title={reasonTarget ? `填写原因 · #${reasonTarget.id}` : '填写原因'}
        open={reasonModalOpen}
        okText="确认提交"
        cancelText="取消"
        confirmLoading={actionMutation.isPending}
        onCancel={() => {
          setReasonModalOpen(false);
          setReasonTarget(null);
          setReasonText('');
        }}
        onOk={async () => {
          if (!reasonTarget) return;
          if (!reasonText.trim()) {
            message.warning('请输入原因');
            return;
          }
          await actionMutation.mutateAsync({ id: reasonTarget.id, action: reasonTarget.action, reason: reasonText.trim() });
          setReasonModalOpen(false);
          setReasonTarget(null);
          setReasonText('');
        }}
        destroyOnClose
      >
        <Form layout="vertical">
          <Form.Item label="原因" required>
            <Input.TextArea rows={3} value={reasonText} onChange={(e) => setReasonText(e.target.value)} placeholder="请填写本次操作的原因（必填）" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
