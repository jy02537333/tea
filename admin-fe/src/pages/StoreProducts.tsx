import { useMemo, useState } from 'react';
import { Button, Form, InputNumber, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Store,
  StoreProduct,
  StoreProductListParams,
  deleteStoreProduct,
  getStore,
  listStoreProducts,
  upsertStoreProduct,
} from '../services/stores';
import { listProducts } from '../services/products';

const { Title, Text } = Typography;

const BIZ_TYPE_OPTIONS = [
  { label: '全部类型', value: undefined },
  { label: '服务商品', value: 1 },
  { label: '外卖商品', value: 2 },
  { label: '其他', value: 3 },
];

const BIZ_TYPE_MAP: Record<number, { label: string; color: string }> = {
  1: { label: '服务', color: 'blue' },
  2: { label: '外卖', color: 'green' },
  3: { label: '其他', color: 'purple' },
};

interface FilterValues {
  biz_type?: number;
}

interface StoreProductFormValues {
  product_id: number;
  stock: number;
  price_override?: number;
  biz_type?: number;
}

export default function StoreProductsPage() {
  const navigate = useNavigate();
  const params = useParams();
  const storeId = Number(params.id || 0);
  const queryClient = useQueryClient();

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [filterForm] = Form.useForm<FilterValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<StoreProduct | null>(null);
  const [form] = Form.useForm<StoreProductFormValues>();

  const listParams: StoreProductListParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      biz_type: filters.biz_type,
    }),
    [filters.biz_type, pagination.limit, pagination.page]
  );

  const storeQuery = useQuery<Store>({
    queryKey: ['store', storeId],
    queryFn: () => getStore(storeId),
    enabled: storeId > 0,
  });

  const productsQuery = useQuery({
    queryKey: ['products-simple'],
    queryFn: () => listProducts({ page: 1, limit: 1000 }),
  });

  const inventoryQuery = useQuery({
    queryKey: ['storeProducts', storeId, listParams.page, listParams.limit, listParams.biz_type ?? 'all'],
    queryFn: () => listStoreProducts(storeId, listParams),
    enabled: storeId > 0,
    placeholderData: keepPreviousData,
  });

  const saveMutation = useMutation({
    mutationFn: async (values: StoreProductFormValues) => {
      const payload = {
        product_id: values.product_id,
        stock: values.stock,
        price_override:
          typeof values.price_override === 'number' && !Number.isNaN(values.price_override)
            ? values.price_override.toFixed(2)
            : undefined,
        biz_type: values.biz_type,
      };
      return upsertStoreProduct(storeId, payload);
    },
    onSuccess: () => {
      message.success(editing ? '门店商品已更新' : '门店商品已绑定');
      queryClient.invalidateQueries({ queryKey: ['storeProducts'] });
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
    },
    onError: (error: any) => {
      message.error(error?.message || '保存失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (record: StoreProduct) => deleteStoreProduct(storeId, record.product_id),
    onSuccess: () => {
      message.success('门店商品绑定已删除');
      queryClient.invalidateQueries({ queryKey: ['storeProducts'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '删除失败');
    },
  });

  const columns: ColumnsType<StoreProduct> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '商品ID', dataIndex: 'product_id', width: 100 },
    {
      title: '商品名称',
      dataIndex: ['product', 'name'],
      render: (_, record) => record.product?.name || '-',
    },
    {
      title: '门店库存',
      dataIndex: 'stock',
      width: 100,
    },
    {
      title: '门店售价',
      dataIndex: 'price_override',
      width: 120,
      render: (val: string, record) => {
        if (val && Number(val) > 0) return `￥${Number(val).toFixed(2)}`;
        if (record.product?.price) return `跟随平台价（￥${Number(record.product.price).toFixed(2)}）`;
        return '-';
      },
    },
    {
      title: '商品类型',
      dataIndex: 'biz_type',
      width: 120,
      render: (val: number) => {
        const meta = BIZ_TYPE_MAP[val];
        return meta ? <Tag color={meta.color}>{meta.label}</Tag> : '-';
      },
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
          <Button type="link" onClick={() => openModal(record)}>
            编辑
          </Button>
          <Button type="link" danger loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate(record)}>
            解绑
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
    setPagination({ page: 1, limit: 20 });
  };

  const openModal = (record?: StoreProduct) => {
    setEditing(record || null);
    if (record) {
      form.setFieldsValue({
        product_id: record.product_id,
        stock: record.stock,
        price_override: record.price_override ? Number(record.price_override) : undefined,
        biz_type: record.biz_type || 1,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({ stock: 0, biz_type: 1 });
    }
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await saveMutation.mutateAsync(values);
    } catch (error) {
      // antd 已提示
    }
  };

  const productOptions = (productsQuery.data?.data || []).map((p) => ({
    label: `${p.id} - ${p.name}`,
    value: p.id,
  }));

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Button type="link" onClick={() => navigate('/stores')}>
        ← 返回门店列表
      </Button>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          门店商品管理{' '}
          {storeId ? <Text type="secondary">(门店ID: {storeId}{storeQuery.data ? ` · ${storeQuery.data.name}` : ''})</Text> : null}
        </Title>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="biz_type" label="商品类型">
            <Select
              allowClear
              style={{ width: 160 }}
              placeholder="全部类型"
              options={BIZ_TYPE_OPTIONS.filter((opt) => opt.value !== undefined)}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={inventoryQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset} disabled={inventoryQuery.isFetching}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Space>

      <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
        <Button type="primary" onClick={() => openModal()}>
          新增/绑定门店商品
        </Button>
      </Space>

      <Table<StoreProduct>
        rowKey="id"
        dataSource={inventoryQuery.data?.list || []}
        loading={inventoryQuery.isFetching}
        columns={columns}
        pagination={{
          current: inventoryQuery.data?.page ?? pagination.page,
          pageSize: inventoryQuery.data?.limit ?? pagination.limit,
          total: inventoryQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (page, pageSize) => {
            setPagination({ page, limit: pageSize || pagination.limit });
          },
        }}
      />

      <Modal
        title={editing ? '编辑门店商品' : '新增/绑定门店商品'}
        open={modalOpen}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        onOk={handleSubmit}
        confirmLoading={saveMutation.isPending}
        destroyOnClose
      >
        <Form layout="vertical" form={form} initialValues={{ stock: 0, biz_type: 1 }}>
          <Form.Item
            label="平台商品"
            name="product_id"
            rules={[{ required: true, message: '请选择平台商品' }]}
          >
            <Select
              placeholder="请选择要绑定的商品"
              showSearch
              optionFilterProp="label"
              options={productOptions}
              loading={productsQuery.isLoading}
            />
          </Form.Item>
          <Form.Item label="门店库存" name="stock" rules={[{ required: true, message: '请输入门店库存' }]}> 
            <InputNumber min={0} step={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="门店售价（可选）" name="price_override">
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} placeholder="不填则跟随平台价" />
          </Form.Item>
          <Form.Item label="商品类型" name="biz_type" rules={[{ required: true, message: '请选择商品类型' }]}> 
            <Select
              options={BIZ_TYPE_OPTIONS.filter((opt) => opt.value !== undefined) as { label: string; value: number }[]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
