import { useMemo, useState } from 'react';
import { Button, Drawer, Form, Input, InputNumber, Popconfirm, Select, Space, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Category,
  CategoryPayload,
  createCategory,
  deleteCategory,
  getCategories,
  updateCategory,
} from '../services/categories';

const STATUS_OPTIONS: { label: string; value?: number }[] = [
  { label: '全部', value: undefined },
  { label: '启用', value: 1 },
  { label: '停用', value: 2 },
];

const STATUS_VALUE_OPTIONS = STATUS_OPTIONS.filter(
  (opt): opt is { label: string; value: number } => typeof opt.value === 'number'
);

const statusTag = (status?: number) => {
  if (status === 1) return <Tag color="green">启用</Tag>;
  if (status === 2) return <Tag color="red">停用</Tag>;
  return <Tag>未知</Tag>;
};

interface FilterState {
  status?: number;
  parent_id?: number;
  keyword?: string;
}

export default function CategoriesPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<FilterState>({});
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [filterForm] = Form.useForm<FilterState>();
  const [form] = Form.useForm<CategoryPayload>();

  const queryFilters = useMemo(
    () => ({ status: filters.status, parent_id: filters.parent_id }),
    [filters.parent_id, filters.status]
  );

  const categoriesQuery = useQuery({
    queryKey: ['categories', queryFilters.status ?? 'all', queryFilters.parent_id ?? 'all'],
    queryFn: () => getCategories(queryFilters),
  });

  const categoryOptionsQuery = useQuery({
    queryKey: ['categoryOptions'],
    queryFn: () => getCategories(),
    staleTime: 60_000,
  });

  const displayData = useMemo(() => {
    const list = categoriesQuery.data ?? [];
    if (!filters.keyword) return list;
    const keyword = filters.keyword.trim().toLowerCase();
    if (!keyword) return list;
    return list.filter((item) => {
      const source = `${item.name ?? ''} ${item.description ?? ''}`.toLowerCase();
      return source.includes(keyword);
    });
  }, [categoriesQuery.data, filters.keyword]);

  const parentNameMap = useMemo(() => {
    const map = new Map<number, string>();
    categoryOptionsQuery.data?.forEach((cat) => map.set(cat.id, cat.name));
    return map;
  }, [categoryOptionsQuery.data]);

  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditing(null);
    form.resetFields();
  };

  const openDrawer = (category?: Category) => {
    if (category) {
      setEditing(category);
      form.setFieldsValue({
        name: category.name,
        description: category.description,
        sort: category.sort,
        parent_id: category.parent_id,
        image: category.image,
        status: category.status ?? 1,
      });
    } else {
      setEditing(null);
      form.resetFields();
      form.setFieldsValue({ status: 1, sort: 0 });
    }
    setDrawerOpen(true);
  };

  const saveMutation = useMutation({
    mutationFn: async (values: CategoryPayload) => {
      if (editing) {
        return updateCategory(editing.id, values);
      }
      return createCategory(values);
    },
    onSuccess: () => {
      message.success(editing ? '分类已更新' : '分类已创建');
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['categoryOptions'] });
      closeDrawer();
    },
    onError: (error: any) => {
      message.error(error?.message || '保存失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (category: Category) => deleteCategory(category.id),
    onSuccess: () => {
      message.success('已删除分类');
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['categoryOptions'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '删除失败');
    },
  });

  const columns: ColumnsType<Category> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name', width: 180 },
    {
      title: '父级分类',
      dataIndex: 'parent_id',
      render: (val?: number) => (val ? parentNameMap.get(val) || `#${val}` : '顶级'),
    },
    { title: '排序', dataIndex: 'sort', width: 90 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (val?: number) => statusTag(val),
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
          <Button type="link" onClick={() => openDrawer(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该分类吗？" onConfirm={() => deleteMutation.mutate(record)}>
            <Button type="link" danger loading={deleteMutation.isPending}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const handleFilter = (values: FilterState) => {
    setFilters(values);
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
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
            <Input allowClear placeholder="名称/描述" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} options={STATUS_VALUE_OPTIONS} placeholder="全部状态" />
          </Form.Item>
          <Form.Item name="parent_id" label="父级分类">
            <Select
              allowClear
              placeholder="全部"
              style={{ width: 160 }}
              loading={categoryOptionsQuery.isLoading}
              options={categoryOptionsQuery.data?.map((cat) => ({ label: cat.name, value: cat.id }))}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={categoriesQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Button type="primary" onClick={() => openDrawer()}>
          新增分类
        </Button>
      </Space>

      <Table
        bordered
        loading={categoriesQuery.isLoading}
        rowKey="id"
        dataSource={displayData}
        columns={columns}
        pagination={false}
      />

      <Drawer
        title={editing ? '编辑分类' : '新增分类'}
        width={480}
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
        <Form layout="vertical" form={form} initialValues={{ status: 1, sort: 0 }}>
          <Form.Item label="分类名称" name="name" rules={[{ required: true, message: '请输入分类名称' }]}>
            <Input placeholder="例如 热门茶饮" />
          </Form.Item>
          <Form.Item label="父级分类" name="parent_id">
            <Select
              allowClear
              placeholder="选择父级（可选）"
              loading={categoryOptionsQuery.isLoading}
              options={categoryOptionsQuery.data?.map((cat) => ({ label: cat.name, value: cat.id }))}
            />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={STATUS_VALUE_OPTIONS} />
          </Form.Item>
          <Form.Item label="图片 URL" name="image">
            <Input placeholder="https://example.com/banner.jpg" />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea rows={4} placeholder="分类描述" />
          </Form.Item>
        </Form>
      </Drawer>
    </Space>
  );
}
