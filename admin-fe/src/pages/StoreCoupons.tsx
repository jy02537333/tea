import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, DatePicker, Form, Input, InputNumber, Modal, Select, Space, Spin, Table, Tag, Typography, message } from 'antd';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ColumnsType } from 'antd/es/table';
import type { Store } from '../services/stores';
import { StoreCouponPayload, createStoreCoupon, getStores, listStoreCoupons, updateStoreCoupon } from '../services/stores';
import type { Coupon } from '../services/types';
import dayjs from 'dayjs';

const COUPON_TYPE_LABELS: Record<number, string> = {
  1: '满减券',
  2: '折扣券',
  3: '免单券',
};

const COUPON_STATUS_COLORS: Record<number, string> = {
  1: 'green',
  2: 'red',
};

export default function StoreCouponsPage() {
  const queryClient = useQueryClient();
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<number | undefined>(undefined);
  const [form] = Form.useForm<StoreCouponFormValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingCoupon, setEditingCoupon] = useState<any | null>(null);

  const storesQuery = useQuery({
    queryKey: ['stores-for-coupons'],
    queryFn: () => getStores({ page: 1, limit: 200 }),
    placeholderData: keepPreviousData,
  });

  const storeOptions = useMemo(
    () =>
      (storesQuery.data?.list ?? []).map((s: Store) => ({
        label: s.name || `门店 #${s.id}`,
        value: s.id,
      })),
    [storesQuery.data?.list]
  );

  useEffect(() => {
    if (!selectedStoreId && storeOptions.length === 1) {
      setSelectedStoreId(storeOptions[0].value);
    }
  }, [selectedStoreId, storeOptions]);

  const couponsQuery = useQuery<Coupon[]>({
    queryKey: ['store-coupons', selectedStoreId, statusFilter ?? 'all'],
    queryFn: () => listStoreCoupons(selectedStoreId!, { status: statusFilter }),
    enabled: !!selectedStoreId,
  });

  const createMutation = useMutation({
    mutationFn: async (values: StoreCouponFormValues) => {
      if (!selectedStoreId) throw new Error('未选择门店');
      const payload = normalizeFormToPayload(values);
      return createStoreCoupon(selectedStoreId, payload);
    },
    onSuccess: () => {
      message.success('优惠券已创建');
      setModalOpen(false);
      setEditingCoupon(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-coupons'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '创建优惠券失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (values: StoreCouponFormValues) => {
      if (!selectedStoreId || !editingCoupon?.id) throw new Error('未选择门店或优惠券');
      const payload = normalizeFormToPayload(values);
      return updateStoreCoupon(selectedStoreId, editingCoupon.id, payload);
    },
    onSuccess: () => {
      message.success('优惠券已更新');
      setModalOpen(false);
      setEditingCoupon(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['store-coupons'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新优惠券失败');
    },
  });

  const disableMutation = useMutation({
    mutationFn: async (coupon: Coupon) => {
      if (!selectedStoreId || !coupon.id) throw new Error('未选择门店或优惠券');
      const values = mapCouponToForm(coupon as any);
      values.status = 2;
      const payload = normalizeFormToPayload(values);
      return updateStoreCoupon(selectedStoreId, coupon.id, payload);
    },
    onSuccess: () => {
      message.success('已禁用该优惠券');
      queryClient.invalidateQueries({ queryKey: ['store-coupons'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '禁用优惠券失败');
    },
  });

  const columns: ColumnsType<Coupon> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '名称', dataIndex: 'name', width: 200 },
    {
      title: '类型',
      dataIndex: 'type',
      width: 120,
      render: (val?: number) => (val ? COUPON_TYPE_LABELS[val] || `类型${val}` : '-'),
    },
    { title: '面值/折扣', dataIndex: 'amount', width: 140 },
    { title: '最低消费', dataIndex: 'min_amount', width: 140 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (val?: number) => {
        if (!val) return '-';
        const color = COUPON_STATUS_COLORS[val] || 'default';
        const label = val === 1 ? '启用' : val === 2 ? '禁用' : `状态${val}`;
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: '总发放', dataIndex: 'total_count', width: 100 },
    { title: '已使用', dataIndex: 'used_count', width: 100 },
    { title: '开始时间', dataIndex: 'start_time', width: 180 },
    { title: '结束时间', dataIndex: 'end_time', width: 180 },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setEditingCoupon(record as any);
              form.setFieldsValue(mapCouponToForm(record as any));
              setModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            onClick={() => {
              setEditingCoupon(null);
              const base = mapCouponToForm(record as any);
              form.setFieldsValue({
                ...base,
                name: `${(record as any).name || ''} 副本`,
              });
              setModalOpen(true);
            }}
          >
            复制新建
          </Button>
          <Button
            type="link"
            danger
            disabled={(record as any).status === 2 || disableMutation.isPending}
            onClick={() => disableMutation.mutate(record as Coupon)}
          >
            一键禁用
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>门店优惠券</Typography.Title>

      <Card>
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Space wrap>
            <Typography.Text>选择门店：</Typography.Text>
            <Select
              style={{ minWidth: 240 }}
              placeholder="请选择门店"
              loading={storesQuery.isLoading}
              options={storeOptions}
              value={selectedStoreId}
              onChange={(val) => setSelectedStoreId(val)}
              allowClear
            />
            <Select
              allowClear
              placeholder="全部状态"
              style={{ minWidth: 160 }}
              value={statusFilter}
              onChange={(val) => setStatusFilter(val)}
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 2 },
              ]}
            />
          </Space>
          {storesQuery.isError && <Alert type="error" message="无法获取门店列表" showIcon />}
          {!selectedStoreId && !storesQuery.isLoading && (
            <Typography.Text type="secondary">请选择要查看的门店，以查看该门店配置的优惠券。</Typography.Text>
          )}
        </Space>
      </Card>

      {selectedStoreId && (
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              onClick={() => {
                setEditingCoupon(null);
                form.resetFields();
                setModalOpen(true);
              }}
            >
              新建优惠券
            </Button>
          </Space>
          {couponsQuery.isLoading && <Spin />}
          {couponsQuery.isError && <Alert type="error" message="无法获取优惠券列表" showIcon />}
          {couponsQuery.data && (
            <Table<Coupon>
              size="small"
              rowKey="id"
              dataSource={couponsQuery.data}
              pagination={false}
              columns={columns}
            />
          )}
        </Card>
      )}

      <Modal
        title={editingCoupon ? '编辑优惠券' : '新建优惠券'}
        open={modalOpen}
        onCancel={() => {
          if (createMutation.isPending || updateMutation.isPending) return;
          setModalOpen(false);
          setEditingCoupon(null);
        }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnClose
      >
        <Form<StoreCouponFormValues>
          layout="vertical"
          form={form}
          initialValues={{ type: 1, status: 1, total_count: 100 }}
          onFinish={(values) => {
            if (editingCoupon) updateMutation.mutate(values);
            else createMutation.mutate(values);
          }}
        >
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入优惠券名称' }]}
          >
            <Input maxLength={100} placeholder="例如：满30减5元" />
          </Form.Item>
          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: '请选择优惠券类型' }]}
          >
            <Select
              options={[
                { label: '满减券', value: 1 },
                { label: '折扣券', value: 2 },
                { label: '免单券', value: 3 },
              ]}
            />
          </Form.Item>
          <Form.Item shouldUpdate={(prev, curr) => prev.type !== curr.type} noStyle>
            {({ getFieldValue }) => {
              const type = getFieldValue('type');
              if (type === 2) {
                return (
                  <Form.Item
                    label="折扣"
                    name="discount"
                    rules={[{ required: true, message: '请输入折扣' }]}
                  >
                    <InputNumber min={0.1} max={1} step={0.01} style={{ width: '100%' }} placeholder="例如：0.8 表示 8 折" />
                  </Form.Item>
                );
              }
              if (type === 1) {
                return (
                  <Form.Item
                    label="减免金额(元)"
                    name="amount"
                    rules={[{ required: true, message: '请输入减免金额' }]}
                  >
                    <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} />
                  </Form.Item>
                );
              }
              return null;
            }}
          </Form.Item>
          <Form.Item label="最低消费(元)" name="min_amount">
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="总发放数量"
            name="total_count"
            rules={[{ required: true, message: '请输入总发放数量' }]}
          >
            <InputNumber min={1} step={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="状态"
            name="status"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 2 },
              ]}
            />
          </Form.Item>
          <Form.Item
            label="开始时间"
            name="start_time"
            rules={[{ required: true, message: '请选择开始时间' }]}
          >
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label="结束时间"
            name="end_time"
            rules={[{ required: true, message: '请选择结束时间' }]}
          >
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea rows={3} maxLength={200} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}

interface StoreCouponFormValues {
  name: string;
  type: number;
  amount?: number;
  discount?: number;
  min_amount?: number;
  total_count: number;
  status: number;
  start_time: dayjs.Dayjs;
  end_time: dayjs.Dayjs;
  description?: string;
}

function normalizeFormToPayload(values: StoreCouponFormValues): StoreCouponPayload {
  return {
    name: values.name,
    type: values.type,
    amount: values.type === 1 ? values.amount : undefined,
    discount: values.type === 2 ? values.discount : undefined,
    min_amount: values.min_amount,
    total_count: values.total_count,
    status: values.status,
    start_time: values.start_time.toISOString(),
    end_time: values.end_time.toISOString(),
    description: values.description,
  };
}

function mapCouponToForm(coupon: any): StoreCouponFormValues {
  return {
    name: coupon.name,
    type: coupon.type,
    amount: coupon.amount ? Number(coupon.amount) : undefined,
    discount: coupon.discount ? Number(coupon.discount) : undefined,
    min_amount: coupon.min_amount ? Number(coupon.min_amount) : undefined,
    total_count: coupon.total_count,
    status: coupon.status,
    start_time: coupon.start_time ? dayjs(coupon.start_time) : dayjs(),
    end_time: coupon.end_time ? dayjs(coupon.end_time) : dayjs(),
    description: coupon.description,
  };
}
