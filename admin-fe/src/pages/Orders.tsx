import { useCallback, useMemo, useState } from 'react';
import { Button, Descriptions, Divider, Drawer, Form, Input, InputNumber, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AdminOrder, AdminOrderItem, AdminOrderListParams, exportAdminOrders, getAdminOrderDetail, getAdminOrders, postOrderAction } from '../services/orders';
import { useAuthContext } from '../hooks/useAuth';

const { Title } = Typography;

const ORDER_STATUS_OPTIONS = [
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

type OrderActionMeta = {
  key: string;
  label: string;
  confirm: string;
  shouldShow: (order: AdminOrder) => boolean;
  danger?: boolean;
  type?: 'primary';
  permission?: string;
};

const ORDER_ACTIONS: OrderActionMeta[] = [
  {
    key: 'deliver',
    label: '发货',
    confirm: '确认发货该订单？',
    shouldShow: (order) => order.status === 2,
    permission: 'order:deliver',
  },
  {
    key: 'complete',
    label: '完成',
    confirm: '确认标记订单完成？',
    shouldShow: (order) => order.status === 3,
    type: 'primary',
    permission: 'order:complete',
  },
  {
    key: 'admin-cancel',
    label: '管理员取消',
    confirm: '确认强制取消该订单？',
    shouldShow: (order) => order.status > 0 && order.status < 4,
    danger: true,
    permission: 'order:cancel',
  },
  {
    key: 'refund/start',
    label: '标记退款中',
    confirm: '确认标记该订单为退款中？',
    shouldShow: (order) => order.pay_status === 2 && order.status >= 2 && order.status < 5,
    permission: 'order:refund',
  },
  {
    key: 'refund/confirm',
    label: '确认退款完成',
    confirm: '确认退款已完成？',
    shouldShow: (order) => order.pay_status === 3,
    permission: 'order:refund',
  },
  {
    key: 'refund',
    label: '立即退款',
    confirm: '确认执行立即退款？',
    shouldShow: (order) => order.pay_status === 2,
    danger: true,
    permission: 'order:refund',
  },
];

const ACTION_REASON_MAP: Record<string, string> = {
  'admin-cancel': '后台强制取消',
  refund: '后台立即退款',
  'refund/start': '后台标记退款中',
  'refund/confirm': '后台确认退款完成',
};

const itemColumns: ColumnsType<AdminOrderItem> = [
  { title: '商品', dataIndex: 'product_name' },
  { title: 'SKU', dataIndex: 'sku_name', width: 120, render: (val?: string) => val || '-' },
  { title: '数量', dataIndex: 'quantity', width: 80 },
  {
    title: '单价',
    dataIndex: 'price',
    width: 120,
    render: (val: number | string) => formatCurrency(val),
  },
  {
    title: '小计',
    dataIndex: 'amount',
    width: 120,
    render: (val: number | string) => formatCurrency(val),
  },
];

interface FilterValues {
  status?: number;
  store_id?: number;
  order_no?: string;
}

export default function OrdersPage() {
  const queryClient = useQueryClient();
  const { hasPermission } = useAuthContext();
  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [detailId, setDetailId] = useState<number | null>(null);
  const [filterForm] = Form.useForm<FilterValues>();

  const listParams: AdminOrderListParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      store_id: filters.store_id,
      status: filters.status,
    }),
    [filters.status, filters.store_id, pagination.limit, pagination.page]
  );

  const ordersQuery = useQuery({
    queryKey: ['adminOrders', listParams.page, listParams.limit, listParams.store_id ?? 'all', listParams.status ?? 'all'],
    queryFn: () => getAdminOrders(listParams),
    placeholderData: keepPreviousData,
  });

  const orderDetailQuery = useQuery({
    queryKey: ['adminOrderDetail', detailId],
    queryFn: () => getAdminOrderDetail(detailId!),
    enabled: !!detailId,
  });

  const detailOrder = orderDetailQuery.data?.order;
  const detailItems = orderDetailQuery.data?.items ?? [];
  const allowAction = useCallback((meta: OrderActionMeta) => !meta.permission || hasPermission(meta.permission), [hasPermission]);
  const detailActions = detailOrder ? ORDER_ACTIONS.filter((meta) => meta.shouldShow(detailOrder) && allowAction(meta)) : [];

  const exportMutation = useMutation({
    mutationFn: async () => {
      const blob = await exportAdminOrders({ store_id: filters.store_id, status: filters.status });
      const url = window.URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = 'orders_export.csv';
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      window.URL.revokeObjectURL(url);
    },
    onError: (error: any) => {
      message.error(error?.message || '导出失败');
    },
    onSuccess: () => {
      message.success('导出任务已开始');
    },
  });

  const actionMutation = useMutation({
    mutationFn: async ({ id, action }: { id: number; action: string }) => {
      const reason = ACTION_REASON_MAP[action];
      await postOrderAction(id, action, reason ? { reason } : undefined);
    },
    onSuccess: async () => {
      message.success('操作成功');
      await queryClient.invalidateQueries({ queryKey: ['adminOrders'] });
      if (detailId) queryClient.invalidateQueries({ queryKey: ['adminOrderDetail', detailId] });
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
    { title: '门店ID', dataIndex: 'store_id', width: 100 },
    { title: '用户ID', dataIndex: 'user_id', width: 100 },
    {
      title: '金额',
      dataIndex: 'pay_amount',
      render: (val: number) => formatCurrency(val),
      width: 120,
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
      width: 220,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" onClick={() => handleOpenDrawer(record.id)}>
            查看详情
          </Button>
          {ORDER_ACTIONS.filter((meta) => meta.shouldShow(record) && allowAction(meta)).map((meta) => (
            <Popconfirm key={meta.key} title={meta.confirm} onConfirm={() => actionMutation.mutate({ id: record.id, action: meta.key })}>
              <Button
                type="link"
                danger={meta.danger}
                disabled={actionMutation.isPending}
                style={meta.type === 'primary' ? { fontWeight: 600 } : undefined}
              >
                {meta.label}
              </Button>
            </Popconfirm>
          ))}
        </Space>
      ),
    },
  ];

  const handleOpenDrawer = (id: number) => {
    setDetailId(id);
    setDrawerOpen(true);
  };

  const handleFilter = (values: FilterValues) => {
    setFilters({
      status: values.status,
      store_id: values.store_id,
      order_no: values.order_no?.trim(),
    });
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, limit: 20 });
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="order_no" label="订单号">
            <Input allowClear placeholder="模糊搜索订单号" />
          </Form.Item>
          <Form.Item name="store_id" label="门店ID">
            <InputNumber min={1} style={{ width: 160 }} placeholder="门店 ID" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 160 }} options={ORDER_STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={ordersQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => ordersQuery.refetch()} loading={ordersQuery.isFetching}>
            刷新
          </Button>
          <Button icon={<DownloadOutlined />} onClick={() => exportMutation.mutate()} loading={exportMutation.isPending}>
            导出
          </Button>
        </Space>
      </Space>

      <Table
        bordered
        rowKey="id"
        loading={ordersQuery.isLoading}
        dataSource={tableData}
        columns={columns}
        pagination={{
          current: pagination.page,
          pageSize: pagination.limit,
          total: ordersQuery.data?.total,
          showSizeChanger: true,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize || pagination.limit }),
        }}
      />

      <Drawer
        title={detailId ? `订单详情 · #${detailId}` : '订单详情'}
        width={640}
        open={drawerOpen}
        onClose={() => {
          setDrawerOpen(false);
          setDetailId(null);
        }}
        destroyOnClose
      >
        {orderDetailQuery.isLoading && <p>加载中...</p>}
        {orderDetailQuery.isError && <p style={{ color: 'red' }}>无法加载订单详情</p>}
        {detailOrder && (
          <>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="订单号">{detailOrder.order_no}</Descriptions.Item>
              <Descriptions.Item label="门店 ID">{detailOrder.store_id}</Descriptions.Item>
              <Descriptions.Item label="用户 ID">{detailOrder.user_id}</Descriptions.Item>
              <Descriptions.Item label="金额">{formatCurrency(detailOrder.pay_amount)}</Descriptions.Item>
              <Descriptions.Item label="状态">
                {ORDER_STATUS_MAP[detailOrder.status]?.label || detailOrder.status}
              </Descriptions.Item>
              <Descriptions.Item label="支付状态">
                {PAY_STATUS_MAP[detailOrder.pay_status]?.label || detailOrder.pay_status}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {detailOrder.created_at ? new Date(detailOrder.created_at).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {detailOrder.updated_at ? new Date(detailOrder.updated_at).toLocaleString() : '-'}
              </Descriptions.Item>
            </Descriptions>
            <Divider />
            <Title level={5}>可执行操作</Title>
            <Space wrap>
              {detailActions.length === 0 && <span style={{ color: '#888' }}>暂无可执行操作</span>}
              {detailActions.map((meta) => (
                <Popconfirm key={meta.key} title={meta.confirm} onConfirm={() => actionMutation.mutate({ id: detailOrder.id, action: meta.key })}>
                  <Button
                    type={meta.type === 'primary' ? 'primary' : 'default'}
                    danger={meta.danger}
                    loading={actionMutation.isPending}
                  >
                    {meta.label}
                  </Button>
                </Popconfirm>
              ))}
            </Space>
            {detailItems.length ? (
              <>
                <Divider />
                <Title level={5}>商品明细</Title>
                <Table
                  size="small"
                  rowKey="id"
                  columns={itemColumns}
                  dataSource={detailItems}
                  pagination={false}
                />
              </>
            ) : null}
          </>
        )}
      </Drawer>
    </Space>
  );
}
