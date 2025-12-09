import { useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Descriptions,
  Divider,
  Drawer,
  Form,
  Input,
  InputNumber,
  Popconfirm,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Store,
  StoreListParams,
  StoreOrderStats,
  StorePayload,
  createStore,
  deleteStore,
  getStoreOrderStats,
  getStores,
  updateStore,
} from '../services/stores';

const STATUS_OPTIONS: { label: string; value?: number }[] = [
  { label: '全部', value: undefined },
  { label: '营业中', value: 1 },
  { label: '停业', value: 2 },
];

const STATUS_VALUE_OPTIONS = STATUS_OPTIONS.filter(
  (opt): opt is { label: string; value: number } => typeof opt.value === 'number'
);

const ORDER_STATUS_LABELS: Record<number, string> = {
  1: '待付款',
  2: '已付款',
  3: '配送中',
  4: '已完成',
  5: '已取消',
};

const ORDER_STATUS_ORDER = [1, 2, 3, 4, 5];

const statusTag = (status?: number) => {
  if (status === 1) return <Tag color="green">营业中</Tag>;
  if (status === 2) return <Tag color="red">停业</Tag>;
  return <Tag>未知</Tag>;
};

interface FilterState {
  keyword?: string;
  status?: number;
}

export default function StoresPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<FilterState>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<Store | null>(null);
  const [statsStore, setStatsStore] = useState<Store | null>(null);
  const [filterForm] = Form.useForm<FilterState>();
  const [form] = Form.useForm<StorePayload>();

  const listParams: StoreListParams = useMemo(
    () => ({ page: pagination.page, limit: pagination.limit, status: filters.status }),
    [filters.status, pagination.limit, pagination.page]
  );

  const storesQuery = useQuery({
    queryKey: ['stores', listParams.page, listParams.limit, listParams.status ?? 'all'],
    queryFn: () => getStores(listParams),
    placeholderData: keepPreviousData,
  });

  const tableData = useMemo(() => {
    const list = storesQuery.data?.list ?? [];
    if (!filters.keyword) return list;
    const keyword = filters.keyword.trim().toLowerCase();
    if (!keyword) return list;
    return list.filter((store: Store) => {
      const source = `${store.name ?? ''} ${store.address ?? ''} ${store.phone ?? ''}`.toLowerCase();
      return source.includes(keyword);
    });
  }, [filters.keyword, storesQuery.data?.list]);

  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditing(null);
    form.resetFields();
  };

  const openStatsDrawer = (store: Store) => {
    setStatsStore(store);
  };

  const closeStatsDrawer = () => {
    setStatsStore(null);
  };

  const openDrawer = (store?: Store) => {
    if (store) {
      setEditing(store);
      form.setFieldsValue({
        name: store.name,
        address: store.address,
        phone: store.phone,
        latitude: store.latitude,
        longitude: store.longitude,
        business_hours: store.business_hours,
        images: store.images,
        status: store.status ?? 1,
      });
    } else {
      setEditing(null);
      form.resetFields();
      form.setFieldsValue({ status: 1 });
    }
    setDrawerOpen(true);
  };

  const saveMutation = useMutation({
    mutationFn: async (values: StorePayload) => {
      if (editing) {
        return updateStore(editing.id, values);
      }
      return createStore(values);
    },
    onSuccess: () => {
      message.success(editing ? '门店已更新' : '门店已创建');
      queryClient.invalidateQueries({ queryKey: ['stores'] });
      closeDrawer();
    },
    onError: (error: any) => {
      message.error(error?.message || '保存失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (store: Store) => deleteStore(store.id),
    onSuccess: () => {
      message.success('已删除门店');
      queryClient.invalidateQueries({ queryKey: ['stores'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '删除失败');
    },
  });

  const columns: ColumnsType<Store> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '门店名称', dataIndex: 'name', width: 200 },
    { title: '地址', dataIndex: 'address' },
    { title: '电话', dataIndex: 'phone', width: 140 },
    {
      title: '营业状态',
      dataIndex: 'status',
      width: 120,
      render: (val?: number) => statusTag(val),
    },
    {
      title: '营业时间',
      dataIndex: 'business_hours',
      width: 160,
      ellipsis: true,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
      render: (val?: string) => (val ? new Date(val).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button type="link" onClick={() => openStatsDrawer(record)}>
            统计
          </Button>
          <Button type="link" onClick={() => openDrawer(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该门店吗？" onConfirm={() => deleteMutation.mutate(record)}>
            <Button type="link" danger loading={deleteMutation.isPending}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const orderStatsQuery = useQuery<StoreOrderStats, Error>({
    queryKey: ['store-order-stats', statsStore?.id],
    queryFn: () => getStoreOrderStats(statsStore!.id),
    enabled: !!statsStore,
  });

  const statsTableData = useMemo(() => {
    const counts = new Map<number, number>();
    orderStatsQuery.data?.status_counts?.forEach((item: { status: number; count: number }) =>
      counts.set(item.status, item.count)
    );
    return ORDER_STATUS_ORDER.map((status) => ({
      key: status,
      status,
      label: ORDER_STATUS_LABELS[status] ?? `状态${status}`,
      count: counts.get(status) ?? 0,
    }));
  }, [orderStatsQuery.data?.status_counts]);

  const statsColumns: ColumnsType<{ key: number; status: number; label: string; count: number }> = [
    { title: '状态', dataIndex: 'label' },
    { title: '订单数量', dataIndex: 'count', width: 120 },
  ];

  const handleFilter = (values: FilterState) => {
    setFilters(values);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, limit: 20 });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await saveMutation.mutateAsync({ ...values, status: values.status ?? 1 });
    } catch (error) {
      // antd 已提示
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="keyword" label="关键字">
            <Input allowClear placeholder="名称/地址/电话" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} placeholder="全部状态" options={STATUS_VALUE_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={storesQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Button type="primary" onClick={() => openDrawer()}>
          新增门店
        </Button>
      </Space>

      <Table
        bordered
        loading={storesQuery.isLoading}
        rowKey="id"
        dataSource={tableData}
        columns={columns}
        pagination={{
          current: pagination.page,
          pageSize: pagination.limit,
          total: storesQuery.data?.total,
          showSizeChanger: true,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize || pagination.limit }),
        }}
      />

      <Drawer
        title={editing ? '编辑门店' : '新增门店'}
        width={520}
        open={drawerOpen}
        onClose={closeDrawer}
        destroyOnClose
        extra={
          <Space>
            <Button onClick={closeDrawer}>取消</Button>
            <Button type="primary" onClick={handleSubmit} loading={saveMutation.isPending}>
              保存
            </Button>
          </Space>
        }
      >
        <Form layout="vertical" form={form} initialValues={{ status: 1 }}>
          <Form.Item label="门店名称" name="name" rules={[{ required: true, message: '请输入门店名称' }]}>
            <Input placeholder="例如：茶心阁 · 西湖店" />
          </Form.Item>
          <Form.Item label="地址" name="address">
            <Input placeholder="杭州市西湖区..." />
          </Form.Item>
          <Form.Item label="联系电话" name="phone">
            <Input placeholder="138xxxx0000" />
          </Form.Item>
          <Space size={16} style={{ width: '100%' }}>
            <Form.Item label="纬度" name="latitude" style={{ flex: 1 }}>
              <InputNumber style={{ width: '100%' }} placeholder="30.2741" step={0.000001} />
            </Form.Item>
            <Form.Item label="经度" name="longitude" style={{ flex: 1 }}>
              <InputNumber style={{ width: '100%' }} placeholder="120.1551" step={0.000001} />
            </Form.Item>
          </Space>
          <Form.Item label="营业时间" name="business_hours">
            <Input placeholder="每天 10:00 - 22:00" />
          </Form.Item>
          <Form.Item label="图片 URL" name="images">
            <Input placeholder="多个地址可用逗号分隔" />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={STATUS_VALUE_OPTIONS} />
          </Form.Item>
        </Form>
      </Drawer>

      <Drawer
        title={statsStore ? `门店订单统计 · ${statsStore.name}` : '门店订单统计'}
        width={420}
        open={!!statsStore}
        onClose={closeStatsDrawer}
        destroyOnClose
      >
        {orderStatsQuery.isLoading && (
          <Space style={{ width: '100%', justifyContent: 'center', padding: '24px 0' }}>
            <Spin />
          </Space>
        )}
        {orderStatsQuery.isError && (
          <Alert type="error" message="无法获取统计数据" description={orderStatsQuery.error?.message} showIcon />
        )}
        {orderStatsQuery.data && (
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="门店">{statsStore?.name}</Descriptions.Item>
              <Descriptions.Item label="总订单数">{orderStatsQuery.data.total_orders}</Descriptions.Item>
              <Descriptions.Item label="成交额（已完成）">
                ￥{Number(orderStatsQuery.data.completed_amount || 0).toFixed(2)}
              </Descriptions.Item>
            </Descriptions>
            <Divider style={{ margin: '12px 0' }} />
            <Table
              size="small"
              rowKey="key"
              columns={statsColumns}
              dataSource={statsTableData}
              pagination={false}
            />
          </Space>
        )}
      </Drawer>
    </Space>
  );
}
