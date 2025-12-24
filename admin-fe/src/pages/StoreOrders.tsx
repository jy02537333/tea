import { useMemo, useState } from 'react';
import { Alert, Button, Form, Input, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { AdminOrder, getAdminStoreOrders, acceptAdminOrder, rejectAdminOrder, createPrintTask } from '../services/orders';

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
  const params = useParams();
  const storeId = Number(params.id || 0);

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, pageSize: 20 });
  const [filterForm] = Form.useForm<FilterValues>();

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
    queryFn: () => getAdminStoreOrders(storeId, listParams),
    enabled: storeId > 0,
    placeholderData: keepPreviousData,
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
      width: 260,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => navigate(`/orders?orderId=${record.id}&storeId=${storeId}`)}
          >
            在订单操作区打开
          </Button>
          <Button
            size="small"
            onClick={async () => {
              try {
                await acceptAdminOrder(record.id, { note: 'store accept' });
                message.success('已接受订单');
                ordersQuery.refetch();
              } catch (e: any) {
                message.error(e?.message || '接受失败');
              }
            }}
          >
            接受
          </Button>
          <Button
            size="small"
            danger
            onClick={async () => {
              try {
                await rejectAdminOrder(record.id, { reason: 'store reject' });
                message.success('已拒绝订单');
                ordersQuery.refetch();
              } catch (e: any) {
                message.error(e?.message || '拒绝失败');
              }
            }}
          >
            拒绝
          </Button>
          <Button
            size="small"
            onClick={async () => {
              try {
                const resp = await createPrintTask({ order_id: record.id, target: 'receipt', payload: JSON.stringify({ order_no: record.order_no }), priority: 1 });
                const id = (resp as any)?.id ?? (resp as any)?.data?.id;
                message.success(`已创建打印任务${id ? `：${id}` : ''}`);
              } catch (e: any) {
                message.error(e?.message || '打印任务创建失败');
              }
            }}
          >
            打印任务
          </Button>
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
    </Space>
  );
}
