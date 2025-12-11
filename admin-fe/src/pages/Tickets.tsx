import { useMemo, useState } from 'react';
import { Button, Card, Drawer, Form, Input, InputNumber, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { Ticket, TicketListParams, CreateTicketPayload, UpdateTicketPayload } from '../services/tickets';
import { createTicket, listTickets, updateTicket } from '../services/tickets';

const STATUS_OPTIONS = [
  { label: '新建', value: 'new' },
  { label: '处理中', value: 'pending' },
  { label: '待用户补充', value: 'waiting_customer' },
  { label: '已解决', value: 'resolved' },
  { label: '已关闭', value: 'closed' },
  { label: '已驳回', value: 'rejected' },
];

const STATUS_TAGS: Record<string, { label: string; color: string }> = {
  new: { label: '新建', color: 'blue' },
  pending: { label: '处理中', color: 'gold' },
  waiting_customer: { label: '待用户', color: 'purple' },
  resolved: { label: '已解决', color: 'green' },
  closed: { label: '已关闭', color: 'default' },
  rejected: { label: '已驳回', color: 'red' },
};

const PRIORITY_OPTIONS = [
  { label: '低', value: 'low' },
  { label: '正常', value: 'normal' },
  { label: '高', value: 'high' },
];

const TYPE_OPTIONS = [
  { label: '咨询', value: 'consult' },
  { label: '订单问题', value: 'order' },
  { label: '退款问题', value: 'refund' },
  { label: '充值问题', value: 'recharge' },
  { label: '投诉建议', value: 'complaint' },
  { label: '其他', value: 'other' },
];

const SOURCE_OPTIONS = [
  { label: '小程序意见反馈', value: 'miniapp_feedback' },
  { label: '订单页投诉', value: 'miniapp_order' },
  { label: '门店员工反馈', value: 'store_staff' },
  { label: '客服电话录入', value: 'phone' },
  { label: '后台人工录入', value: 'manual' },
];

interface FilterValues {
  status?: string;
  type?: string;
  source?: string;
  priority?: string;
  store_id?: number;
  user_id?: number;
  keyword?: string;
}

export default function TicketsPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [activeTicket, setActiveTicket] = useState<Ticket | null>(null);
  const [filterForm] = Form.useForm<FilterValues>();
  const [editForm] = Form.useForm<UpdateTicketPayload & { assignee_id?: number }>();
  const [createForm] = Form.useForm<CreateTicketPayload>();

  const listParams: TicketListParams = useMemo(
    () => ({
      page: pagination.page,
      limit: pagination.limit,
      status: filters.status,
      type: filters.type,
      source: filters.source,
      priority: filters.priority,
      store_id: filters.store_id,
      user_id: filters.user_id,
      keyword: filters.keyword,
    }),
    [filters, pagination.limit, pagination.page]
  );

  const listQuery = useQuery({
    queryKey: ['tickets', listParams],
    queryFn: () => listTickets(listParams),
    placeholderData: keepPreviousData,
  });

  const createMutation = useMutation({
    mutationFn: async (values: CreateTicketPayload) => {
      const payload: CreateTicketPayload = {
        ...values,
        title: values.title.trim(),
        content: values.content?.trim() || undefined,
        attachments: values.attachments?.trim() || undefined,
      };
      return createTicket(payload);
    },
    onSuccess: async () => {
      message.success('工单已创建');
      setCreating(false);
      createForm.resetFields();
      await queryClient.invalidateQueries({ queryKey: ['tickets'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '创建工单失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: async (values: UpdateTicketPayload & { assignee_id?: number }) => {
      if (!activeTicket) throw new Error('未选择工单');
      const payload: UpdateTicketPayload = {
        ...values,
        remark: values.remark?.trim() || undefined,
        reject_reason: values.reject_reason?.trim() || undefined,
      };
      return updateTicket(activeTicket.id, payload);
    },
    onSuccess: async () => {
      message.success('工单已更新');
      await queryClient.invalidateQueries({ queryKey: ['tickets'] });
      setDrawerOpen(false);
      setActiveTicket(null);
      editForm.resetFields();
    },
    onError: (err: any) => {
      message.error(err?.message || '更新工单失败');
    },
  });

  const columns: ColumnsType<Ticket> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '类型', dataIndex: 'type', width: 120, render: (val: string) => TYPE_OPTIONS.find((o) => o.value === val)?.label || val },
    { title: '来源', dataIndex: 'source', width: 140, render: (val: string) => SOURCE_OPTIONS.find((o) => o.value === val)?.label || val },
    { title: '用户ID', dataIndex: 'user_id', width: 100 },
    { title: '订单ID', dataIndex: 'order_id', width: 100 },
    { title: '门店ID', dataIndex: 'store_id', width: 100 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (val: string) => {
        const meta = STATUS_TAGS[val];
        return meta ? <Tag color={meta.color}>{meta.label}</Tag> : val;
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      render: (val: string) => {
        const label = PRIORITY_OPTIONS.find((o) => o.value === val)?.label || val;
        const color = val === 'high' ? 'red' : val === 'low' ? 'default' : 'blue';
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: '负责人ID', dataIndex: 'assignee_id', width: 100 },
    { title: '标题', dataIndex: 'title' },
    { title: '最近更新时间', dataIndex: 'updated_at', width: 180 },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Button
          type="link"
          onClick={() => {
            setActiveTicket(record);
            setDrawerOpen(true);
            editForm.setFieldsValue({
              status: record.status,
              priority: record.priority,
              assignee_id: record.assignee_id ?? undefined,
              remark: record.remark,
              reject_reason: record.reject_reason,
            });
          }}
        >
          详情/处理
        </Button>
      ),
    },
  ];

  const handleTableChange = (paginationInfo: any) => {
    setPagination({ page: paginationInfo.current, limit: paginationInfo.pageSize });
  };

  const handleFilterSubmit = (values: FilterValues) => {
    setFilters(values);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const resetFilters = () => {
    filterForm.resetFields();
    setFilters({});
    setPagination({ page: 1, limit: 20 });
  };

  const list = listQuery.data?.list ?? [];
  const total = listQuery.data?.total ?? 0;

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>客服工单</Typography.Title>

      <Card>
        <Form form={filterForm} layout="inline" onFinish={handleFilterSubmit}>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} options={STATUS_OPTIONS} placeholder="全部" />
          </Form.Item>
          <Form.Item name="type" label="类型">
            <Select allowClear style={{ width: 140 }} options={TYPE_OPTIONS} placeholder="全部" />
          </Form.Item>
          <Form.Item name="source" label="来源">
            <Select allowClear style={{ width: 180 }} options={SOURCE_OPTIONS} placeholder="全部" />
          </Form.Item>
          <Form.Item name="priority" label="优先级">
            <Select allowClear style={{ width: 120 }} options={PRIORITY_OPTIONS} placeholder="全部" />
          </Form.Item>
          <Form.Item name="store_id" label="门店ID">
            <InputNumber min={1} style={{ width: 120 }} />
          </Form.Item>
          <Form.Item name="user_id" label="用户ID">
            <InputNumber min={1} style={{ width: 120 }} />
          </Form.Item>
          <Form.Item name="keyword" label="关键词">
            <Input placeholder="标题/内容" allowClear style={{ width: 200 }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                查询
              </Button>
              <Button onClick={resetFilters}>重置</Button>
              <Button type="dashed" onClick={() => setCreating(true)}>
                新建工单
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card>
        <Table<Ticket>
          rowKey="id"
          loading={listQuery.isLoading}
          dataSource={list}
          columns={columns}
          pagination={{
            current: pagination.page,
            pageSize: pagination.limit,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      <Drawer
        title={activeTicket ? `工单 #${activeTicket.id}` : '工单详情'}
        width={520}
        open={drawerOpen}
        onClose={() => {
          setDrawerOpen(false);
          setActiveTicket(null);
          editForm.resetFields();
        }}
        destroyOnClose
      >
        {activeTicket && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Typography.Paragraph>
              <strong>标题：</strong>
              {activeTicket.title}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>内容：</strong>
              {activeTicket.content || '-'}
            </Typography.Paragraph>
            <Typography.Paragraph>
              <strong>关联信息：</strong>
              用户ID {activeTicket.user_id || '-'} / 订单ID {activeTicket.order_id || '-'} / 门店ID {activeTicket.store_id || '-'}
            </Typography.Paragraph>
            <Form layout="vertical" form={editForm} onFinish={(vals) => updateMutation.mutate(vals)}>
              <Form.Item name="status" label="状态">
                <Select options={STATUS_OPTIONS} />
              </Form.Item>
              <Form.Item name="priority" label="优先级">
                <Select options={PRIORITY_OPTIONS} />
              </Form.Item>
              <Form.Item name="assignee_id" label="负责人ID">
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="remark" label="内部备注">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item name="reject_reason" label="驳回/关闭原因">
                <Input.TextArea rows={3} />
              </Form.Item>
              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>
                    保存
                  </Button>
                  <Button
                    onClick={() => {
                      editForm.resetFields();
                      if (activeTicket) {
                        editForm.setFieldsValue({
                          status: activeTicket.status,
                          priority: activeTicket.priority,
                          assignee_id: activeTicket.assignee_id ?? undefined,
                          remark: activeTicket.remark,
                          reject_reason: activeTicket.reject_reason,
                        });
                      }
                    }}
                  >
                    重置表单
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Space>
        )}
      </Drawer>

      <Drawer
        title="新建工单"
        width={520}
        open={creating}
        onClose={() => {
          setCreating(false);
          createForm.resetFields();
        }}
        destroyOnClose
      >
        <Form layout="vertical" form={createForm} onFinish={(vals) => createMutation.mutate(vals)}>
          <Form.Item name="type" label="类型" rules={[{ required: true, message: '请选择工单类型' }]}>
            <Select options={TYPE_OPTIONS} placeholder="请选择类型" />
          </Form.Item>
          <Form.Item name="source" label="来源" rules={[{ required: true, message: '请选择工单来源' }]}>
            <Select options={SOURCE_OPTIONS} placeholder="请选择来源" />
          </Form.Item>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input maxLength={200} showCount placeholder="请简要概括问题" />
          </Form.Item>
          <Form.Item name="content" label="详细描述">
            <Input.TextArea rows={4} placeholder="请输入详细描述" />
          </Form.Item>
          <Form.Item name="priority" label="优先级" initialValue="normal">
            <Select options={PRIORITY_OPTIONS} />
          </Form.Item>
          <Form.Item name="user_id" label="用户ID">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="order_id" label="订单ID">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="store_id" label="门店ID">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="attachments" label="附件JSON">
            <Input.TextArea rows={3} placeholder="可选：存储截图等附件的 JSON 字符串" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
                提交
              </Button>
              <Button
                onClick={() => {
                  createForm.resetFields();
                }}
              >
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Drawer>
    </Space>
  );
}
