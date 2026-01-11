import { useEffect, useMemo, useRef, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Select, Space, Table, Tabs, Tag, Typography, message, Upload, Image, Divider } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Store,
  StoreProduct,
  StoreProductListParams,
  deleteStoreProduct,
  deleteStoreProductScoped,
  getStore,
  listStoreProducts,
  listStoreProductsScoped,
  upsertStoreProduct,
  upsertStoreProductScoped,
  createStoreExclusiveProduct,
} from '../services/stores';
import { listProducts } from '../services/products';
import { getCategories } from '../services/categories';
import { useAuthContext } from '../hooks/useAuth';
import { getOssPolicy } from '../services/upload';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';

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
  const location = useLocation();
  const params = useParams();
  const storeId = Number(params.id || 0);
  const queryClient = useQueryClient();
  const { user } = useAuthContext();

  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [filterForm] = Form.useForm<FilterValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<StoreProduct | null>(null);
  const [form] = Form.useForm<StoreProductFormValues>();
  const [createOwnedOpen, setCreateOwnedOpen] = useState(false);
  const [ownedForm] = Form.useForm<{ name: string; category_id: number; price: number; stock?: number; price_override?: number; description?: string; image_url?: string }>();
  const [descHtml, setDescHtml] = useState<string>('');
  const [previewUrl, setPreviewUrl] = useState<string | undefined>(undefined);
  const quillRef = useRef<ReactQuill | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const quillModules = useMemo(() => ({
    toolbar: {
      container: [
        [{ header: [1, 2, 3, false] }],
        ['bold', 'italic', 'underline', 'strike'],
        [{ list: 'ordered' }, { list: 'bullet' }],
        [{ align: [] }],
        ['link', 'image'],
        ['clean'],
      ],
      handlers: {
        image: () => {
          // 触发隐藏的文件选择
          fileInputRef.current?.click();
        },
      },
    },
  }), []);

  const searchParams = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const initialTab = searchParams.get('tab') || 'categories';
  const [activeTab, setActiveTab] = useState<string>(initialTab);

  const [statusFilter, setStatusFilter] = useState<number | undefined>(() => {
    if (initialTab === 'on') return 1;
    if (initialTab === 'off') return 2;
    if (initialTab === 'draft') return 0;
    return undefined;
  });

  useEffect(() => {
    setActiveTab(initialTab);
  }, [initialTab]);

  useEffect(() => {
    const nextStatus = activeTab === 'on' ? 1 : activeTab === 'off' ? 2 : activeTab === 'draft' ? 0 : undefined;

    if (activeTab === 'create') {
      openModal();
      return;
    }

    setStatusFilter(nextStatus);
    setPagination((prev) => ({ ...prev, page: 1 }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab]);

  const listParams: StoreProductListParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      biz_type: filters.biz_type,
      status: statusFilter,
    }),
    [filters.biz_type, pagination.limit, pagination.page, statusFilter]
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

  const categoriesQuery = useQuery({
    queryKey: ['categories-all'],
    queryFn: () => getCategories({ status: 1 }),
  });

  const inventoryQuery = useQuery({
    queryKey: ['storeProducts', storeId, listParams.page, listParams.limit, listParams.biz_type ?? 'all', listParams.status ?? 'all'],
    queryFn: () => (user?.role === 'store' ? listStoreProductsScoped(storeId, listParams) : listStoreProducts(storeId, listParams)),
    enabled: storeId > 0,
    placeholderData: keepPreviousData,
  });

  const setTab = (tab: string) => {
    const sp = new URLSearchParams(location.search);
    sp.set('tab', tab);
    navigate({ pathname: location.pathname, search: sp.toString() }, { replace: true });
  };

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
      return user?.role === 'store' ? upsertStoreProductScoped(storeId, payload) : upsertStoreProduct(storeId, payload);
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
    mutationFn: async (record: StoreProduct) =>
      user?.role === 'store'
        ? deleteStoreProductScoped(storeId, record.product_id)
        : deleteStoreProduct(storeId, record.product_id),
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

  useEffect(() => {
    if (createOwnedOpen) {
      setDescHtml('');
      setPreviewUrl(undefined);
      ownedForm.resetFields();
      ownedForm.setFieldValue('stock', 0);
    }
  }, [createOwnedOpen]);

  const submitOwned = async () => {
    try {
      const values = await ownedForm.validateFields();
      const imageUrl = values.image_url?.trim();
      const payload = {
        name: values.name.trim(),
        category_id: values.category_id,
        price: values.price,
        stock: values.stock ?? 0,
        price_override: values.price_override,
        description: (descHtml || values.description || '').trim(),
        images: imageUrl ? JSON.stringify([imageUrl]) : '',
      };
      await createStoreExclusiveProduct(storeId, payload);
      message.success('门店自有商品已创建并上架');
      setCreateOwnedOpen(false);
      ownedForm.resetFields();
      setDescHtml('');
      setPreviewUrl(undefined);
      queryClient.invalidateQueries({ queryKey: ['storeProducts'] });
    } catch (err: any) {
      if (err?.message) message.error(err.message);
    }
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
        <Tabs
          activeKey={activeTab}
          onChange={setTab}
          items={[
            { key: 'categories', label: '分类' },
            { key: 'on', label: '上架中' },
            { key: 'off', label: '已下架' },
            { key: 'draft', label: '草稿' },
            { key: 'create', label: '新增' },
          ]}
        />
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
        <Button type="primary" onClick={() => setTab('create')}>
          绑定平台商品
        </Button>
        {user?.role === 'store' && (
          <Button type="primary" onClick={() => setCreateOwnedOpen(true)}>
            新增自有商品
          </Button>
        )}
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

      <Modal
        title="新增自有商品（仅本门店可用）"
        open={createOwnedOpen}
        onCancel={() => setCreateOwnedOpen(false)}
        onOk={submitOwned}
        confirmLoading={false}
        destroyOnClose
      >
        <Form layout="vertical" form={ownedForm} initialValues={{ stock: 0 }}>
          <Form.Item name="name" label="商品名称" rules={[{ required: true, message: '请输入商品名称' }]}> 
            <Input placeholder="例如：一次性纸杯（门店用品）" />
          </Form.Item>
          <Form.Item name="category_id" label="商品分类" rules={[{ required: true, message: '请选择分类' }]}> 
            <Select
              placeholder="请选择分类"
              options={(categoriesQuery.data || []).map(c => ({ label: c.name, value: c.id }))}
              loading={categoriesQuery.isLoading}
            />
          </Form.Item>
          <Form.Item name="price" label="平台基础价" rules={[{ required: true, message: '请输入价格' }]}> 
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="商品主图">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Upload
                listType="picture-card"
                maxCount={1}
                customRequest={async (options) => {
                  const { file, onError, onSuccess } = options as any;
                  try {
                    const policy = await getOssPolicy();
                    const f = file as File;
                    const key = `${policy.dir}${Date.now()}-${encodeURIComponent(f.name)}`;
                    const form = new FormData();
                    form.append('key', key);
                    form.append('policy', policy.policy);
                    form.append('OSSAccessKeyId', policy.accessid);
                    form.append('signature', policy.signature);
                    form.append('success_action_status', '200');
                    form.append('file', f);
                    const resp = await fetch(policy.host, { method: 'POST', body: form });
                    if (!resp.ok) throw new Error(`OSS 上传失败: ${resp.status}`);
                    const url = `${policy.host}/${key}`;
                    ownedForm.setFieldValue('image_url', url);
                    setPreviewUrl(url);
                    onSuccess?.({}, new XMLHttpRequest());
                  } catch (e: any) {
                    message.error(e?.message || '上传失败');
                    onError?.(e);
                  }
                }}
                onRemove={() => {
                  ownedForm.setFieldValue('image_url', '');
                  setPreviewUrl(undefined);
                }}
              >
                {previewUrl ? null : <div>上传</div>}
              </Upload>
              {previewUrl ? <Image src={previewUrl} width={120} /> : null}
              <Input
                placeholder="或粘贴图片 URL"
                onChange={(e) => {
                  const v = e.target.value || '';
                  ownedForm.setFieldValue('image_url', v);
                  setPreviewUrl(v || undefined);
                }}
                allowClear
              />
            </Space>
          </Form.Item>
          <Form.Item name="stock" label="门店库存"> 
            <InputNumber min={0} step={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="price_override" label="门店售价（可选）"> 
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} placeholder="不填则跟随平台价" />
          </Form.Item>
          <Divider style={{ margin: '8px 0' }} />
          <Form.Item label="商品描述（富文本，可插入图片）">
            <ReactQuill ref={quillRef as any} theme="snow" value={descHtml} onChange={setDescHtml} modules={quillModules} />
            {/* 隐藏的文件输入，用于 Toolbar 的 image 按钮 */}
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              style={{ display: 'none' }}
              onChange={async (e) => {
                const file = e.target.files?.[0];
                if (!file) return;
                try {
                  const policy = await getOssPolicy();
                  const key = `${policy.dir}${Date.now()}-${encodeURIComponent(file.name)}`;
                  const form = new FormData();
                  form.append('key', key);
                  form.append('policy', policy.policy);
                  form.append('OSSAccessKeyId', policy.accessid);
                  form.append('signature', policy.signature);
                  form.append('success_action_status', '200');
                  form.append('file', file);
                  const resp = await fetch(policy.host, { method: 'POST', body: form });
                  if (!resp.ok) throw new Error(`OSS 上传失败: ${resp.status}`);
                  const url = `${policy.host}/${key}`;
                  // 插入到富文本
                  const quill = quillRef.current?.getEditor?.();
                  if (quill) {
                    const range = quill.getSelection(true);
                    const index = range?.index ?? quill.getLength();
                    quill.insertEmbed(index, 'image', url, 'user');
                    quill.setSelection(index + 1, 0, 'user');
                  } else {
                    // 回退：直接拼接 img
                    setDescHtml((prev) => `${prev}<p><img src="${url}" /></p>`);
                  }
                  // 清空选择，避免重复触发
                  e.target.value = '';
                } catch (err: any) {
                  message.error(err?.message || '图片上传失败');
                }
              }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
