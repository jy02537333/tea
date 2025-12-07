import { useCallback, useMemo, useRef, useState } from 'react';
import { Button, Drawer, Form, Input, InputNumber, Popconfirm, Select, Space, Switch, Table, Tag, Upload, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { UploadProps } from 'antd';
import type { UploadRequestOption } from 'rc-upload/lib/interface';
import { UploadOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';
import { Product, ProductPayload, ProductListParams, createProduct, deleteProduct, getProducts, updateProduct } from '../services/products';
import { Category, getCategories } from '../services/categories';
import { uploadMedia } from '../services/upload';

const STATUS_OPTIONS: { label: string; value?: number }[] = [
  { label: '全部', value: undefined },
  { label: '上架', value: 1 },
  { label: '下架', value: 2 },
];

const STATUS_VALUE_OPTIONS = STATUS_OPTIONS.filter(
  (opt): opt is { label: string; value: number } => typeof opt.value === 'number'
);

const statusTag = (status: number) => {
  if (status === 1) return <Tag color="green">上架</Tag>;
  if (status === 2) return <Tag color="red">下架</Tag>;
  return <Tag>未知</Tag>;
};

interface FilterState {
  keyword?: string;
  status?: number;
  category_id?: number;
}

type ProductFormValues = Omit<ProductPayload, 'images'> & {
  image_urls?: string[];
};

const parseImageList = (images?: string): string[] => {
  if (!images) return [];
  try {
    const parsed = JSON.parse(images);
    if (Array.isArray(parsed)) {
      return parsed.filter((item): item is string => typeof item === 'string' && item.trim().length > 0);
    }
  } catch (error) {
    // fall back to comma separated values
  }
  return images
    .split(',')
    .map((url) => url.trim())
    .filter((url) => url.length > 0);
};

export default function ProductsPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<FilterState>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [filterForm] = Form.useForm();
  const [form] = Form.useForm<ProductFormValues>();
  const [imageUploading, setImageUploading] = useState(false);
  const descriptionValue = Form.useWatch('description', form);
  const quillRef = useRef<ReactQuill | null>(null);
  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditing(null);
    form.resetFields();
  };

  const appendImageUrl = (url: string) => {
    const normalized = url.trim();
    if (!normalized) return;
    const current: string[] = form.getFieldValue('image_urls') || [];
    if (current.includes(normalized)) {
      return;
    }
    form.setFieldsValue({ image_urls: [...current, normalized] });
  };

  const uploadProps: UploadProps = {
    showUploadList: false,
    customRequest: async (options: UploadRequestOption) => {
      const { file, onError, onSuccess } = options;
      try {
        setImageUploading(true);
        const response = await uploadMedia(file as File);
        appendImageUrl(response.url);
        message.success('图片上传成功');
        onSuccess?.(response as any);
      } catch (error: any) {
        message.error(error?.message || '图片上传失败');
        onError?.(error as Error);
      } finally {
        setImageUploading(false);
      }
    },
  };

  const handleInsertImage = useCallback(() => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';
    input.onchange = async () => {
      const files = input.files;
      if (!files || files.length === 0) return;
      const file = files[0];
      try {
        setImageUploading(true);
        const response = await uploadMedia(file);
        const quill = quillRef.current?.getEditor();
        if (quill) {
          const range = quill.getSelection(true);
          const insertIndex = range ? range.index : quill.getLength();
          quill.insertEmbed(insertIndex, 'image', response.url, 'user');
          quill.setSelection(insertIndex + 1, 0, 'user');
        }
        message.success('图片已插入描述');
      } catch (error: any) {
        message.error(error?.message || '插入图片失败');
      } finally {
        setImageUploading(false);
        input.value = '';
      }
    };
    input.click();
  }, []);

  const quillModules = useMemo(
    () => ({
      toolbar: {
        container: [
          [{ header: [1, 2, 3, false] }],
          ['bold', 'italic', 'underline', 'strike'],
          [{ color: [] }, { background: [] }],
          [{ list: 'ordered' }, { list: 'bullet' }],
          [{ align: [] }],
          ['link', 'image'],
          ['clean'],
        ],
        handlers: {
          image: handleInsertImage,
        },
      },
    }),
    [handleInsertImage]
  );

  const quillFormats = useMemo(
    () => ['header', 'bold', 'italic', 'underline', 'strike', 'color', 'background', 'list', 'align', 'link', 'image'],
    []
  );

  const listParams: ProductListParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      ...filters,
    }),
    [filters, pagination.limit, pagination.page]
  );

  const productQueryKey: (string | number | undefined)[] = [
    'products',
    listParams.page,
    listParams.limit,
    listParams.category_id ?? 'all',
    listParams.status ?? 'all',
    listParams.keyword ?? '',
  ];

  const productsQuery = useQuery({
    queryKey: productQueryKey,
    queryFn: () => getProducts(listParams),
    placeholderData: keepPreviousData,
  });

  const categoriesQuery = useQuery({
    queryKey: ['categories'],
    queryFn: () => getCategories(),
  });

  const saveMutation = useMutation({
    mutationFn: async (values: ProductPayload) => {
      if (editing) {
        return updateProduct(editing.id, values);
      }
      return createProduct(values);
    },
    onSuccess: () => {
      message.success(editing ? '商品已更新' : '商品已创建');
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setDrawerOpen(false);
      setEditing(null);
    },
    onError: (error: any) => {
      message.error(error?.message || '保存失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (product: Product) => deleteProduct(product.id),
    onSuccess: () => {
      message.success('已删除商品');
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '删除失败');
    },
  });

  const columns: ColumnsType<Product> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name', width: 180 },
    {
      title: '分类',
      dataIndex: ['category', 'name'],
      render: (_, record) => record.category?.name || '-'
    },
    {
      title: '售价',
      dataIndex: 'price',
      render: (val: string) => `￥${val}`,
    },
    { title: '库存', dataIndex: 'stock', width: 90 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (val: number) => statusTag(val),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
      render: (val: string) => (val ? new Date(val).toLocaleString() : '-'),
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
          <Popconfirm title="确定删除该商品吗？" onConfirm={() => deleteMutation.mutate(record)}>
            <Button type="link" danger loading={deleteMutation.isPending}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const openDrawer = (product?: Product) => {
    if (product) {
      setEditing(product);
      form.setFieldsValue({
        name: product.name,
        description: product.description || '',
        category_id: product.category_id,
        price: Number(product.price),
        original_price: product.original_price ? Number(product.original_price) : undefined,
        image_urls: parseImageList(product.images),
        status: product.status,
        stock: product.stock,
        sort: product.sort,
        is_hot: product.is_hot,
        is_new: product.is_new,
        is_recommend: product.is_recommend,
      });
    } else {
      setEditing(null);
      form.resetFields();
      form.setFieldsValue({ status: 1, stock: 0, image_urls: [], description: '' });
    }
    setDrawerOpen(true);
  };

  const handleFilter = (values: FilterState) => {
    setFilters(values);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleResetFilters = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, limit: 20 });
  };

  const handleSubmit = async () => {
    try {
      const values = (await form.validateFields()) as ProductFormValues;
      const imageUrls = (values.image_urls || []).map((url) => url.trim()).filter((url) => url.length > 0);
      const payload: ProductPayload = {
        name: values.name,
        description: values.description ?? '',
        category_id: Number(values.category_id),
        price: Number(values.price),
        original_price: values.original_price !== undefined ? Number(values.original_price) : undefined,
        images: JSON.stringify(imageUrls),
        status: Number(values.status),
        stock: values.stock !== undefined ? Number(values.stock) : 0,
        sort: values.sort !== undefined ? Number(values.sort) : undefined,
        is_hot: values.is_hot,
        is_new: values.is_new,
        is_recommend: values.is_recommend,
      };

      await saveMutation.mutateAsync(payload);
    } catch (error) {
      // 表单错误已由 antd 处理
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="keyword" label="关键字">
            <Input allowClear placeholder="名称/描述" />
          </Form.Item>
          <Form.Item name="category_id" label="分类">
            <Select
              allowClear
              style={{ width: 160 }}
              loading={categoriesQuery.isLoading}
              placeholder="全部分类"
              options={categoriesQuery.data?.map((cat: Category) => ({ label: cat.name, value: cat.id }))}
            />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 120 }} options={STATUS_VALUE_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={productsQuery.isFetching}>
                搜索
              </Button>
              <Button onClick={handleResetFilters}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Button type="primary" onClick={() => openDrawer()}>
          新增商品
        </Button>
      </Space>

      <Table
        bordered
        loading={productsQuery.isLoading}
        rowKey="id"
        dataSource={productsQuery.data?.list || []}
        columns={columns}
        pagination={{
          current: pagination.page,
          pageSize: pagination.limit,
          total: productsQuery.data?.total,
          showSizeChanger: true,
          onChange: (page, pageSize) => setPagination({ page, limit: pageSize || pagination.limit }),
        }}
      />

      <Drawer
        title={editing ? '编辑商品' : '新增商品'}
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
        <Form layout="vertical" form={form} initialValues={{ status: 1, stock: 0, image_urls: [], description: '' }}>
          <Form.Item label="商品名称" name="name" rules={[{ required: true, message: '请输入商品名称' }]}>
            <Input placeholder="请输入商品名称" />
          </Form.Item>
          <Form.Item label="分类" name="category_id" rules={[{ required: true, message: '请选择分类' }]}>
            <Select
              placeholder="请选择分类"
              loading={categoriesQuery.isLoading}
              options={categoriesQuery.data?.map((cat: Category) => ({ label: cat.name, value: cat.id }))}
            />
          </Form.Item>
          <Form.Item label="售价" name="price" rules={[{ required: true, message: '请输入售价' }]}>
            <InputNumber min={0} precision={2} style={{ width: '100%' }} placeholder="例如 19.90" />
          </Form.Item>
          <Form.Item label="原价" name="original_price">
            <InputNumber min={0} precision={2} style={{ width: '100%' }} placeholder="可选" />
          </Form.Item>
          <Form.Item label="库存" name="stock" rules={[{ required: true, message: '请输入库存' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={STATUS_VALUE_OPTIONS} />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="商品图片"
            name="image_urls"
            rules={[{ required: true, message: '请至少添加一张商品图片' }]}
          >
            <Select
              mode="tags"
              style={{ width: '100%' }}
              placeholder="可粘贴图片 URL，或使用下方按钮上传"
            />
          </Form.Item>
          <Form.Item label=" " colon={false}>
            <Upload {...uploadProps} accept="image/*" showUploadList={false}>
              <Button icon={<UploadOutlined />} loading={imageUploading}>
                上传图片到 OSS
              </Button>
            </Upload>
          </Form.Item>
          <Form.Item label="商品描述" name="description" rules={[{ required: true, message: '请输入商品描述' }]}>
            <ReactQuill
              ref={quillRef}
              theme="snow"
              value={descriptionValue}
              onChange={(content) => form.setFieldsValue({ description: content })}
              modules={quillModules}
              formats={quillFormats}
              placeholder="支持图文混排，可直接上传图片插入内容"
              style={{ height: 260, marginBottom: 48 }}
            />
          </Form.Item>
          <Form.Item label="标签">
            <Space>
              <Form.Item name="is_hot" valuePropName="checked" noStyle>
                <Switch checkedChildren="热销" unCheckedChildren="热销" />
              </Form.Item>
              <Form.Item name="is_new" valuePropName="checked" noStyle>
                <Switch checkedChildren="新品" unCheckedChildren="新品" />
              </Form.Item>
              <Form.Item name="is_recommend" valuePropName="checked" noStyle>
                <Switch checkedChildren="推荐" unCheckedChildren="推荐" />
              </Form.Item>
            </Space>
          </Form.Item>
        </Form>
      </Drawer>
    </Space>
  );
}
