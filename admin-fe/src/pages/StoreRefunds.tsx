import { useMemo, useState } from 'react';
import { Alert, Button, Form, Input, InputNumber, Select, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import type { PaginatedResult } from '../services/api';
import { getStore, Store } from '../services/stores';
import { listStoreRefundsRecords, StoreRefundRecord } from '../services/storeRefunds';

const { Title, Text } = Typography;

interface FilterValues {
  refund_no?: string;
  order_id?: number;
  status?: number;
}

const STATUS_OPTIONS: { label: string; value?: number }[] = [
  { label: '全部', value: undefined },
  { label: '申请中', value: 1 },
  { label: '退款成功', value: 2 },
  { label: '退款失败', value: 3 },
];

export default function StoreRefundsPage() {
  const navigate = useNavigate();
  const params = useParams();
  const storeId = Number(params.id || 0);

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [filterForm] = Form.useForm<FilterValues>();

  const storeQuery = useQuery<Store>({
    queryKey: ['store', storeId],
    queryFn: () => getStore(storeId),
    enabled: storeId > 0,
  });

  const queryParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      refund_no: filters.refund_no?.trim() || undefined,
      order_id: filters.order_id,
      status: filters.status,
    }),
    [filters.order_id, filters.refund_no, filters.status, pagination.limit, pagination.page]
  );

  const refundsQuery = useQuery<PaginatedResult<StoreRefundRecord>>({
    queryKey: ['storeRefunds', storeId, queryParams.page, queryParams.limit, queryParams.refund_no ?? '', queryParams.order_id ?? 0, queryParams.status ?? 'all'],
    queryFn: () => listStoreRefundsRecords(storeId, queryParams),
    enabled: storeId > 0,
    placeholderData: keepPreviousData,
  });

  const columns: ColumnsType<StoreRefundRecord> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '退款单号', dataIndex: 'refund_no', width: 220 },
    {
      title: '订单号',
      key: 'order_no',
      width: 180,
      render: (_, record) => record.order?.order_no || '-',
    },
    { title: '订单ID', dataIndex: 'order_id', width: 110 },
    {
      title: '退款金额',
      dataIndex: 'refund_amount',
      width: 120,
      render: (val: string) => `￥${val}`,
    },
    { title: '状态', dataIndex: 'status', width: 120 },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: (val?: string) => (val ? new Date(val).toLocaleString() : '-'),
    },
  ];

  const handleFilter = (values: FilterValues) => {
    setFilters({
      refund_no: values.refund_no?.trim(),
      order_id: values.order_id ? Number(values.order_id) : undefined,
      status: typeof values.status === 'number' ? values.status : undefined,
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
      <Button type="link" onClick={() => navigate('/store-orders')}>
        ← 返回订单管理
      </Button>

      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          售后订单（仅退款）{' '}
          {storeId ? (
            <Text type="secondary">
              (门店ID: {storeId}
              {storeQuery.data ? ` · ${storeQuery.data.name}` : ''})
            </Text>
          ) : null}
        </Title>

        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="refund_no" label="退款单号">
            <Input allowClear placeholder="模糊搜索退款单号" style={{ width: 220 }} />
          </Form.Item>
          <Form.Item name="order_id" label="订单ID">
            <InputNumber min={1} precision={0} placeholder="订单ID" style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 160 }} options={STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={refundsQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset} disabled={refundsQuery.isFetching}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Space>

      {refundsQuery.isError && (
        <Alert
          type="error"
          showIcon
          message="加载退款列表失败"
          description={(refundsQuery.error as any)?.message || '请确认已登录且具备权限'}
        />
      )}

      <Table<StoreRefundRecord>
        bordered
        rowKey="id"
        dataSource={refundsQuery.data?.list || []}
        loading={refundsQuery.isFetching}
        columns={columns}
        pagination={{
          current: refundsQuery.data?.page ?? pagination.page,
          pageSize: refundsQuery.data?.limit ?? pagination.limit,
          total: refundsQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize || pagination.limit }),
        }}
      />
    </Space>
  );
}
