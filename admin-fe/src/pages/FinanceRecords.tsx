import { useMemo, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useQuery } from '@tanstack/react-query';
import dayjs, { type Dayjs } from 'dayjs';
import {
  listPaymentsRecords,
  listRefundsRecords,
  exportPaymentsRecords,
  exportRefundsRecords,
  type PaymentRecord,
  type RefundRecord,
} from '../services/financeRecords';
import { PAYMENT_METHOD_LABELS, PAYMENT_STATUS_LABELS, REFUND_STATUS_LABELS } from '../constants/payment';

const { RangePicker } = DatePicker;

interface FilterValues {
  order_id?: number;
  store_id?: number;
  payment_no?: string;
  refund_no?: string;
  status?: string;
  method?: string;
  range?: [Dayjs, Dayjs];
}

function formatAmount(val: string | number | undefined): string {
  if (val == null) return '-';
  const n = Number(val);
  if (Number.isNaN(n)) return String(val);
  return n.toFixed(2);
}

export default function FinanceRecordsPage() {
  const [activeTab, setActiveTab] = useState<'payments' | 'refunds'>('payments');
  const [filters, setFilters] = useState<FilterValues>({});
  const [pagination, setPagination] = useState({ page: 1, limit: 20 });
  const [form] = Form.useForm<FilterValues>();

  const queryParams = useMemo(() => {
    const base: any = {
      page: pagination.page,
      limit: pagination.limit,
      order_id: filters.order_id,
      store_id: filters.store_id,
      status: filters.status,
      method: activeTab === 'payments' ? filters.method : undefined,
    };
    if (activeTab === 'payments' && filters.payment_no) base.payment_no = filters.payment_no.trim();
    if (activeTab === 'refunds' && filters.refund_no) base.refund_no = filters.refund_no.trim();
    if (filters.range && filters.range.length === 2) {
      base.start = filters.range[0].format('YYYY-MM-DD');
      base.end = filters.range[1].format('YYYY-MM-DD');
    }
    return base;
  }, [activeTab, filters, pagination.page, pagination.limit]);

  const paymentsQuery = useQuery({
    queryKey: ['financePayments', queryParams, activeTab],
    queryFn: async () => {
      if (activeTab !== 'payments') return null;
      return await listPaymentsRecords(queryParams);
    },
    enabled: activeTab === 'payments',
    placeholderData: (old) => old,
  });

  const refundsQuery = useQuery({
    queryKey: ['financeRefunds', queryParams, activeTab],
    queryFn: async () => {
      if (activeTab !== 'refunds') return null;
      return await listRefundsRecords(queryParams);
    },
    enabled: activeTab === 'refunds',
    placeholderData: (old) => old,
  });

  const payments = paymentsQuery.data;
  const refunds = refundsQuery.data;

  const handleFilterSubmit = (values: FilterValues) => {
    setFilters(values);
    setPagination({ page: 1, limit: pagination.limit });
  };

  const handleTableChange = (paginationInfo: any) => {
    setPagination({ page: paginationInfo.current, limit: paginationInfo.pageSize });
  };

  const handleExport = async () => {
    try {
      const paramsForExport: any = { ...queryParams, format: 'csv' };
      let blob: Blob;
      if (activeTab === 'payments') {
        blob = await exportPaymentsRecords(paramsForExport);
      } else {
        blob = await exportRefundsRecords(paramsForExport);
      }
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${activeTab === 'payments' ? 'payments' : 'refunds'}_${Date.now()}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      message.success('导出任务已完成，CSV 已下载');
    } catch (err: any) {
      message.error(err?.message || '导出失败');
    }
  };

  const paymentColumns: ColumnsType<PaymentRecord> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '支付单号', dataIndex: 'payment_no', width: 200 },
    { title: '订单ID', dataIndex: 'order_id', width: 100 },
    { title: '门店', dataIndex: 'store_name', width: 160 },
    { title: '金额', dataIndex: 'amount', width: 120, render: (val: string) => formatAmount(val) },
    {
      title: '支付方式',
      dataIndex: 'payment_method',
      width: 120,
      render: (val: number) => PAYMENT_METHOD_LABELS[val] ?? `方式${val}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (val: number) => {
        const label = PAYMENT_STATUS_LABELS[val] ?? `状态${val}`;
        const color = val === 2 ? 'green' : val === 1 ? 'default' : 'red';
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ];

  const refundColumns: ColumnsType<RefundRecord> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '退款单号', dataIndex: 'refund_no', width: 200 },
    { title: '订单ID', dataIndex: 'order_id', width: 100 },
    { title: '门店', dataIndex: 'store_name', width: 160 },
    { title: '退款金额', dataIndex: 'refund_amount', width: 120, render: (val: string) => formatAmount(val) },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (val: number) => {
        const label = REFUND_STATUS_LABELS[val] ?? `状态${val}`;
        const color = val === 2 ? 'green' : val === 1 ? 'default' : 'red';
        return <Tag color={color}>{label}</Tag>;
      },
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ];

  const total = activeTab === 'payments' ? payments?.total ?? 0 : refunds?.total ?? 0;
  const dataSource = (activeTab === 'payments' ? payments?.list : refunds?.list) ?? [];
  const loading = activeTab === 'payments' ? paymentsQuery.isLoading : refundsQuery.isLoading;

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>财务记录（收支流水）</Typography.Title>

      <Card>
        <Form<FilterValues> layout="inline" form={form} onFinish={handleFilterSubmit}>
          <Form.Item name="order_id" label="订单ID">
            <InputNumber min={1} style={{ width: 140 }} />
          </Form.Item>
          <Form.Item name="store_id" label="门店ID">
            <InputNumber min={1} style={{ width: 140 }} />
          </Form.Item>
          {activeTab === 'payments' && (
            <Form.Item name="payment_no" label="支付单号">
              <Input allowClear style={{ width: 200 }} />
            </Form.Item>
          )}
          {activeTab === 'refunds' && (
            <Form.Item name="refund_no" label="退款单号">
              <Input allowClear style={{ width: 200 }} />
            </Form.Item>
          )}
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} placeholder="全部" />
          </Form.Item>
          {activeTab === 'payments' && (
            <Form.Item name="method" label="支付方式">
              <Select allowClear style={{ width: 160 }} placeholder="全部" />
            </Form.Item>
          )}
          <Form.Item name="range" label="时间范围">
            <RangePicker
              allowClear
              style={{ width: 260 }}
              defaultValue={[dayjs().subtract(6, 'day'), dayjs()]}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                查询
              </Button>
              <Button
                onClick={() => {
                  form.resetFields();
                  setFilters({});
                  setPagination({ page: 1, limit: 20 });
                }}
              >
                重置
              </Button>
              <Button onClick={handleExport}>导出当前条件</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            const next = key === 'refunds' ? 'refunds' : 'payments';
            setActiveTab(next);
            setPagination({ page: 1, limit: 20 });
          }}
          items={[
            { key: 'payments', label: '收款记录', children: null },
            { key: 'refunds', label: '退款记录', children: null },
          ]}
        />
        <Table
          rowKey="id"
          loading={loading}
          dataSource={dataSource as any[]}
          columns={(activeTab === 'payments' ? paymentColumns : refundColumns) as any}
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
    </Space>
  );
}
