import { useMemo, useState } from 'react';
import { Alert, Button, Divider, Form, Input, InputNumber, Select, Space, Table, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { Navigate, useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthContext } from '../hooks/useAuth';
import { getCategories } from '../services/categories';
import { addCartItem, clearCart, listCart, removeCartItem, updateCartItem } from '../services/cart';
import { getStoreExclusiveProducts, StoreExclusiveProduct } from '../services/storeExclusiveProducts';

const { Title, Text } = Typography;

type CartItemWithRefs = {
  id: number;
  product_id: number;
  sku_id?: number;
  quantity: number;
  store_price_override?: string | number;
  effective_price?: string | number;
  product?: {
    id: number;
    name?: string;
    price?: string | number;
  };
  sku?: {
    id: number;
    name?: string;
  };
};

interface FilterValues {
  category_id?: number;
  keyword?: string;
}

const formatCurrency = (val?: number | string) => `￥${Number(val ?? 0).toFixed(2)}`;

export default function StoreMallPage() {
  const navigate = useNavigate();
  const params = useParams();
  const storeId = Number(params.id || 0);
  const queryClient = useQueryClient();
  const { user } = useAuthContext();

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [filterForm] = Form.useForm<FilterValues>();

  const isStoreAdmin = user?.role === 'store';
  const lockedStoreId = user?.store_id;

  const listParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      keyword: filters.keyword,
      category_id: filters.category_id,
    }),
    [filters.category_id, filters.keyword, pagination.limit, pagination.page]
  );

  const categoriesQuery = useQuery({
    queryKey: ['categories', 'active'],
    queryFn: () => getCategories({ status: 1 }),
  });

  const productsQuery = useQuery({
    queryKey: [
      'storeExclusiveProducts',
      storeId,
      listParams.page,
      listParams.limit,
      listParams.keyword ?? '',
      listParams.category_id ?? 'all',
    ],
    queryFn: () => getStoreExclusiveProducts(storeId, listParams),
    enabled: storeId > 0 && isStoreAdmin,
    placeholderData: keepPreviousData,
  });

  const cartQuery = useQuery({
    queryKey: ['cart'],
    queryFn: async () => (await listCart({ store_id: storeId })) as unknown as CartItemWithRefs[],
    enabled: isStoreAdmin,
  });

  const addToCartMutation = useMutation({
    mutationFn: async (productId: number) => {
      await addCartItem(productId, null, 1);
    },
    onSuccess: async () => {
      message.success('已加入购物车');
      await queryClient.invalidateQueries({ queryKey: ['cart'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '加入购物车失败');
    },
  });

  const updateQtyMutation = useMutation({
    mutationFn: async ({ id, quantity }: { id: number; quantity: number }) => {
      await updateCartItem(id, quantity);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['cart'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '更新数量失败');
    },
  });

  const removeItemMutation = useMutation({
    mutationFn: async (id: number) => {
      await removeCartItem(id);
    },
    onSuccess: async () => {
      message.success('已移除');
      await queryClient.invalidateQueries({ queryKey: ['cart'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '移除失败');
    },
  });

  const clearCartMutation = useMutation({
    mutationFn: async () => {
      await clearCart();
    },
    onSuccess: async () => {
      message.success('购物车已清空');
      await queryClient.invalidateQueries({ queryKey: ['cart'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '清空失败');
    },
  });

  const handleFilter = (values: FilterValues) => {
    const keyword = values.keyword?.trim();
    setFilters({
      category_id: values.category_id,
      keyword: keyword || undefined,
    });
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, limit: 20 });
  };

  if (!isStoreAdmin) {
    return <Alert type="warning" showIcon message="商家商城仅门店管理员可用" />;
  }

  if (lockedStoreId && storeId > 0 && storeId !== lockedStoreId) {
    return <Navigate to={`/stores/${lockedStoreId}/mall`} replace />;
  }

  const categoryOptions = (categoriesQuery.data || []).map((c) => ({ label: `${c.id} - ${c.name}`, value: c.id }));

  const productColumns: ColumnsType<StoreExclusiveProduct> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '商品名称', dataIndex: 'name' },
    {
      title: '价格',
      key: 'price',
      width: 160,
      render: (_, record) => {
        const override = record.store_price_override;
        if (override && Number(override) > 0) return `${formatCurrency(override)}（门店价）`;
        return formatCurrency(record.price);
      },
    },
    {
      title: '库存',
      key: 'stock',
      width: 120,
      render: (_, record) => (typeof record.store_stock === 'number' ? record.store_stock : '-'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 140,
      render: (_, record) => (
        <Button type="primary" onClick={() => addToCartMutation.mutate(record.id)} loading={addToCartMutation.isPending}>
          加入购物车
        </Button>
      ),
    },
  ];

  const cartItems = cartQuery.data || [];
  const cartTotalQty = cartItems.reduce((sum, it) => sum + (it.quantity || 0), 0);
  const cartTotalAmount = cartItems.reduce((sum, it) => {
    const unit = Number(it.effective_price ?? it.product?.price ?? 0);
    return sum + unit * (it.quantity || 0);
  }, 0);

  const cartColumns: ColumnsType<CartItemWithRefs> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    {
      title: '商品',
      key: 'product',
      render: (_, record) => record.product?.name || `商品#${record.product_id}`,
    },
    {
      title: '单价',
      key: 'unit_price',
      width: 120,
      render: (_, record) => formatCurrency(record.effective_price ?? record.product?.price ?? 0),
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      width: 160,
      render: (val: number, record) => (
        <InputNumber
          min={0}
          value={val}
          onChange={(v) => updateQtyMutation.mutate({ id: record.id, quantity: Number(v || 0) })}
          style={{ width: 120 }}
        />
      ),
    },
    {
      title: '小计',
      key: 'subtotal',
      width: 140,
      render: (_, record) => {
        const unit = Number(record.effective_price ?? record.product?.price ?? 0);
        return formatCurrency(unit * (record.quantity || 0));
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Button type="link" danger loading={removeItemMutation.isPending} onClick={() => removeItemMutation.mutate(record.id)}>
          删除
        </Button>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Button type="link" onClick={() => navigate('/store-home')}>
        ← 返回门店首页
      </Button>

      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          商家商城 {storeId ? <Text type="secondary">(门店ID: {storeId})</Text> : null}
        </Title>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="category_id" label="分类">
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              style={{ width: 220 }}
              placeholder="全部分类"
              loading={categoriesQuery.isFetching}
              options={categoryOptions}
            />
          </Form.Item>
          <Form.Item name="keyword" label="关键词">
            <Input allowClear placeholder="名称/描述" style={{ width: 220 }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={productsQuery.isFetching}>
                查询
              </Button>
              <Button onClick={handleReset} disabled={productsQuery.isFetching}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Space>

      {storeId === 0 && <Alert type="error" message="缺少有效的门店ID" showIcon />}

      <Table<StoreExclusiveProduct>
        rowKey={(r) => String(r.id)}
        loading={productsQuery.isFetching}
        columns={productColumns}
        dataSource={productsQuery.data?.list || []}
        pagination={{
          current: productsQuery.data?.page ?? pagination.page,
          pageSize: productsQuery.data?.limit ?? pagination.limit,
          total: productsQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize }),
        }}
      />

      <Divider />

      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={5} style={{ margin: 0 }}>
          购物车 {cartTotalQty ? <Text type="secondary">(共 {cartTotalQty} 件)</Text> : null}
        </Title>
        <Space>
          <Text>合计：{formatCurrency(cartTotalAmount)}</Text>
          <Button danger onClick={() => clearCartMutation.mutate()} loading={clearCartMutation.isPending}>
            清空购物车
          </Button>
        </Space>
      </Space>

      <Table<CartItemWithRefs>
        rowKey={(r) => String(r.id)}
        loading={cartQuery.isFetching}
        columns={cartColumns}
        dataSource={cartItems}
        pagination={false}
      />

      <Text type="secondary">提示：当前购物车为通用购物车接口，价格以商品基础价为准。</Text>
    </Space>
  );
}
