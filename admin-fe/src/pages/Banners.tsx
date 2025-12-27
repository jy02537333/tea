import { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { Banner, BannerPayload, BannerQuery } from '../services/banners';
import { createBanner, deleteBanner, listBanners, updateBanner } from '../services/banners';

const { Title } = Typography;

const STATUS_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '启用', value: 1 },
  { label: '停用', value: 2 },
];

const EDIT_STATUS_OPTIONS = [
  { label: '启用', value: 1 },
  { label: '停用', value: 2 },
];

const LINK_TYPE_OPTIONS = [
  { label: '无链接', value: 1 },
  { label: '商品详情', value: 2 },
  { label: '分类页', value: 3 },
  { label: '外部链接', value: 4 },
];

const statusTag = (status?: number) => {
  if (status === 1) return <Tag color="green">启用</Tag>;
  if (status === 2) return <Tag color="red">停用</Tag>;
  return <Tag>未知</Tag>;
};

interface FilterValues {
  keyword?: string;
  status?: number;
}

export default function BannersPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<BannerQuery>({});
  const [filterForm] = Form.useForm<FilterValues>();
  const [form] = Form.useForm<BannerPayload>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Banner | null>(null);

  const bannersQuery = useQuery({
    queryKey: ['banners', filters.keyword ?? '', filters.status ?? 'all'],
    queryFn: () => listBanners(filters),
  });

  useEffect(() => {
    if (!modalOpen) {
      form.resetFields();
      setEditing(null);
    }
  }, [modalOpen, form]);

  const tableData = useMemo(() => bannersQuery.data ?? [], [bannersQuery.data]);

  const createMutation = useMutation({
    mutationFn: (payload: BannerPayload) => createBanner(payload),
    onSuccess: () => {
      message.success('广告已创建');
      setModalOpen(false);
      queryClient.invalidateQueries({ queryKey: ['banners'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: BannerPayload) => {
      if (!editing?.id) throw new Error('未选择广告');
      return updateBanner(editing.id, payload);
    },
    onSuccess: () => {
      message.success('广告已更新');
      setModalOpen(false);
      queryClient.invalidateQueries({ queryKey: ['banners'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新失败');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (b: Banner) => {
      await deleteBanner(b.id);
    },
    onSuccess: () => {
      message.success('已删除');
      queryClient.invalidateQueries({ queryKey: ['banners'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '删除失败');
    },
  });

  const toggleMutation = useMutation({
    mutationFn: async (b: Banner) => {
      const nextStatus = b.status === 1 ? 2 : 1;
      return updateBanner(b.id, {
        title: b.title,
        image_url: b.image_url,
        link_type: b.link_type,
        link_url: b.link_url,
        sort: b.sort,
        status: nextStatus,
      });
    },
    onSuccess: () => {
      message.success('状态已更新');
      queryClient.invalidateQueries({ queryKey: ['banners'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '操作失败');
    },
  });

  const openCreate = () => {
    setEditing(null);
    form.setFieldsValue({
      title: '',
      image_url: '',
      link_type: 1,
      link_url: '',
      sort: 0,
      status: 1,
    });
    setModalOpen(true);
  };

  const openEdit = (b: Banner) => {
    setEditing(b);
    form.setFieldsValue({
      title: b.title,
      image_url: b.image_url,
      link_type: b.link_type ?? 1,
      link_url: b.link_url,
      sort: b.sort ?? 0,
      status: b.status ?? 1,
    });
    setModalOpen(true);
  };

  const handleSubmit = async (values: BannerPayload) => {
    const payload: BannerPayload = {
      title: values.title?.trim(),
      image_url: values.image_url?.trim(),
      link_type: values.link_type,
      link_url: values.link_url?.trim(),
      sort: values.sort ?? 0,
      status: values.status ?? 1,
    };

    if (!payload.image_url) {
      form.setFields([{ name: 'image_url', errors: ['请输入图片地址'] }]);
      return;
    }

    if (editing) {
      await updateMutation.mutateAsync(payload);
    } else {
      await createMutation.mutateAsync(payload);
    }
  };

  const handleFilter = (values: FilterValues) => {
    const next: BannerQuery = {
      keyword: values.keyword?.trim() || undefined,
      status: values.status,
    };
    setFilters(next);
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
  };

  const columns: ColumnsType<Banner> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '标题', dataIndex: 'title', width: 200, render: (v?: string) => v || '-' },
    {
      title: '图片',
      dataIndex: 'image_url',
      width: 260,
      render: (v: string) => (v ? <a href={v} target="_blank" rel="noreferrer">{v}</a> : '-'),
    },
    {
      title: '跳转类型',
      dataIndex: 'link_type',
      width: 120,
      render: (v?: number) => LINK_TYPE_OPTIONS.find((x) => x.value === v)?.label || (v ? `类型${v}` : '-'),
    },
    {
      title: '跳转值',
      dataIndex: 'link_url',
      width: 220,
      render: (v?: string) => (v ? <span>{v}</span> : '-'),
    },
    { title: '排序', dataIndex: 'sort', width: 90, render: (v?: number) => (typeof v === 'number' ? v : '-') },
    { title: '状态', dataIndex: 'status', width: 90, render: (v?: number) => statusTag(v) },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <Space size={8}>
          <Button type="link" onClick={() => openEdit(record)}>
            编辑
          </Button>
          <Button
            type="link"
            danger={record.status === 1}
            onClick={() => toggleMutation.mutate(record)}
            loading={toggleMutation.isPending && toggleMutation.variables?.id === record.id}
          >
            {record.status === 1 ? '下架' : '上架'}
          </Button>
          <Button
            type="link"
            danger
            onClick={() => {
              Modal.confirm({
                title: '确认删除？',
                content: record.title ? `将删除：${record.title}` : undefined,
                okText: '删除',
                cancelText: '取消',
                okButtonProps: { danger: true },
                onOk: async () => {
                  await deleteMutation.mutateAsync(record);
                },
              });
            }}
            loading={deleteMutation.isPending && deleteMutation.variables?.id === record.id}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="keyword" label="关键字">
            <Input allowClear placeholder="按标题查询" style={{ width: 220 }} />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} options={STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={bannersQuery.isFetching}>
                查找
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Space>
          <Button onClick={() => bannersQuery.refetch()} loading={bannersQuery.isFetching}>
            刷新
          </Button>
          <Button type="primary" onClick={openCreate}>
            新增广告
          </Button>
        </Space>
      </Space>

      <Title level={5} style={{ margin: 0 }}>
        广告管理
      </Title>

      <Table
        bordered
        rowKey="id"
        loading={bannersQuery.isLoading}
        dataSource={tableData}
        columns={columns}
        pagination={false}
        scroll={{ x: 1100 }}
      />

      <Modal
        title={editing ? `编辑广告 · #${editing.id}` : '新增广告'}
        open={modalOpen}
        onCancel={() => {
          if (!createMutation.isPending && !updateMutation.isPending) {
            setModalOpen(false);
          }
        }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnClose
      >
        <Form layout="vertical" form={form} onFinish={handleSubmit} initialValues={{ status: 1, link_type: 1, sort: 0 }}>
          <Form.Item name="title" label="标题">
            <Input allowClear placeholder="可选" />
          </Form.Item>
          <Form.Item name="image_url" label="图片地址" rules={[{ required: true, message: '请输入图片地址' }]}>
            <Input allowClear placeholder="https://example.com/banner.jpg" />
          </Form.Item>
          <Form.Item name="link_type" label="跳转类型">
            <Select options={LINK_TYPE_OPTIONS} />
          </Form.Item>
          <Form.Item name="link_url" label="跳转值">
            <Input allowClear placeholder="可选：商品ID/分类ID/外链URL" />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <Input type="number" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={EDIT_STATUS_OPTIONS} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
