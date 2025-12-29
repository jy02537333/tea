import { useMemo, useState } from 'react';
import { Button, DatePicker, Form, Input, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DownloadOutlined } from '@ant-design/icons';
import { keepPreviousData, useMutation, useQuery } from '@tanstack/react-query';
import {
  AccessLog,
  AccessLogQuery,
  OperationLog,
  OperationLogQuery,
  exportAccessLogs,
  exportOperationLogs,
  listAccessLogs,
  listOperationLogs,
} from '../services/logs';

const { Title } = Typography;

const METHOD_OPTIONS = [
  { label: '全部', value: undefined },
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
];

const toISOStringSafe = (val: any) => {
  if (!val) return undefined;
  try {
    const raw = val?.toDate ? val.toDate() : val;
    return new Date(raw).toISOString();
  } catch {
    return undefined;
  }
};

function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  window.URL.revokeObjectURL(url);
}

export default function LogsPage() {
  const [activeKey, setActiveKey] = useState<'operations' | 'access'>('operations');

  // 操作日志
  const [opFilters, setOpFilters] = useState<Omit<OperationLogQuery, 'page' | 'limit'>>({});
  const [opPagination, setOpPagination] = useState({ page: 1, limit: 20 });
  const [opForm] = Form.useForm<{
    module?: string;
    method?: string;
    path?: string;
    user_id?: string;
    order_id?: string;
    date_range?: [any, any];
  }>();

  const opParams: OperationLogQuery = useMemo(
    () => ({
      ...opFilters,
      page: opPagination.page,
      limit: opPagination.limit,
    }),
    [opFilters, opPagination.limit, opPagination.page]
  );

  const opQuery = useQuery({
    queryKey: ['operationLogs', opParams],
    queryFn: () => listOperationLogs(opParams),
    placeholderData: keepPreviousData,
    enabled: activeKey === 'operations',
  });

  const opExportMutation = useMutation({
    mutationFn: async () => {
      const blob = await exportOperationLogs({ ...opFilters, format: 'csv' });
      downloadBlob(blob, `operation_logs_${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.csv`);
    },
    onSuccess: () => message.success('已开始导出'),
    onError: (err: any) => message.error(err?.message || '导出失败'),
  });

  const opColumns: ColumnsType<OperationLog> = [
    { title: 'ID', dataIndex: 'id', width: 90 },
    { title: 'UserID', dataIndex: 'user_id', width: 100, render: (v: any) => (v ?? '-') },
    { title: '模块', dataIndex: 'module', width: 140, render: (v?: string) => v || '-' },
    { title: '操作', dataIndex: 'operation', ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140, render: (v?: string) => v || '-' },
    { title: '时间', dataIndex: 'created_at', width: 200, render: (v?: string) => (v ? new Date(v).toLocaleString() : '-') },
  ];

  const handleOpFilter = (values: any) => {
    const [startRaw, endRaw] = values.date_range ?? [];
    setOpFilters({
      module: values.module?.trim() || undefined,
      method: values.method || undefined,
      path: values.path?.trim() || undefined,
      user_id: values.user_id?.trim() || undefined,
      order_id: values.order_id?.trim() || undefined,
      start: toISOStringSafe(startRaw),
      end: toISOStringSafe(endRaw),
    });
    setOpPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleOpReset = () => {
    opForm.resetFields();
    setOpFilters({});
    setOpPagination({ page: 1, limit: 20 });
  };

  // 访问日志
  const [accessFilters, setAccessFilters] = useState<Omit<AccessLogQuery, 'page' | 'limit'>>({});
  const [accessPagination, setAccessPagination] = useState({ page: 1, limit: 20 });
  const [accessForm] = Form.useForm<{
    method?: string;
    path?: string;
    user_id?: string;
    status?: string;
    date_range?: [any, any];
  }>();

  const accessParams: AccessLogQuery = useMemo(
    () => ({
      ...accessFilters,
      page: accessPagination.page,
      limit: accessPagination.limit,
    }),
    [accessFilters, accessPagination.limit, accessPagination.page]
  );

  const accessQuery = useQuery({
    queryKey: ['accessLogs', accessParams],
    queryFn: () => listAccessLogs(accessParams),
    placeholderData: keepPreviousData,
    enabled: activeKey === 'access',
  });

  const accessExportMutation = useMutation({
    mutationFn: async () => {
      const blob = await exportAccessLogs({ ...accessFilters, format: 'csv' });
      downloadBlob(blob, `access_logs_${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.csv`);
    },
    onSuccess: () => message.success('已开始导出'),
    onError: (err: any) => message.error(err?.message || '导出失败'),
  });

  const accessColumns: ColumnsType<AccessLog> = [
    { title: 'ID', dataIndex: 'id', width: 90 },
    { title: 'UserID', dataIndex: 'user_id', width: 100, render: (v: any) => (v ?? 0) || '-' },
    {
      title: '方法',
      dataIndex: 'method',
      width: 90,
      render: (v?: string) => (v ? <Tag>{v}</Tag> : '-'),
    },
    { title: '路径', dataIndex: 'path', ellipsis: true },
    { title: '状态', dataIndex: 'status_code', width: 90, render: (v?: number) => (v ? <Tag color={v >= 400 ? 'red' : 'green'}>{v}</Tag> : '-') },
    { title: '耗时(ms)', dataIndex: 'latency', width: 110, render: (v?: number) => (typeof v === 'number' ? v : '-') },
    { title: 'IP', dataIndex: 'ip', width: 140, render: (v?: string) => v || '-' },
    { title: '时间', dataIndex: 'created_at', width: 200, render: (v?: string) => (v ? new Date(v).toLocaleString() : '-') },
  ];

  const handleAccessFilter = (values: any) => {
    const [startRaw, endRaw] = values.date_range ?? [];
    setAccessFilters({
      method: values.method || undefined,
      path: values.path?.trim() || undefined,
      user_id: values.user_id?.trim() || undefined,
      status: values.status?.trim() || undefined,
      start: toISOStringSafe(startRaw),
      end: toISOStringSafe(endRaw),
    });
    setAccessPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleAccessReset = () => {
    accessForm.resetFields();
    setAccessFilters({});
    setAccessPagination({ page: 1, limit: 20 });
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          日志
        </Title>
        <Space>
          <Button
            icon={<DownloadOutlined />}
            disabled={activeKey === 'operations' ? opQuery.isFetching : accessQuery.isFetching}
            loading={activeKey === 'operations' ? opExportMutation.isPending : accessExportMutation.isPending}
            onClick={() => (activeKey === 'operations' ? opExportMutation.mutate() : accessExportMutation.mutate())}
          >
            导出 CSV
          </Button>
        </Space>
      </Space>

      <Tabs
        activeKey={activeKey}
        onChange={(k) => setActiveKey(k as any)}
        items={[
          {
            key: 'operations',
            label: '操作日志',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Form layout="inline" form={opForm} onFinish={handleOpFilter}>
                  <Form.Item name="module" label="模块">
                    <Input allowClear placeholder="例如 rbac/order/store" style={{ width: 180 }} />
                  </Form.Item>
                  <Form.Item name="method" label="方法">
                    <Select allowClear placeholder="全部" options={METHOD_OPTIONS.filter((x) => x.value)} style={{ width: 120 }} />
                  </Form.Item>
                  <Form.Item name="path" label="Path">
                    <Input allowClear placeholder="包含匹配" style={{ width: 200 }} />
                  </Form.Item>
                  <Form.Item name="user_id" label="UserID">
                    <Input allowClear placeholder="精确" style={{ width: 120 }} />
                  </Form.Item>
                  <Form.Item name="order_id" label="OrderID">
                    <Input allowClear placeholder="从 request_data 里匹配" style={{ width: 160 }} />
                  </Form.Item>
                  <Form.Item name="date_range" label="时间">
                    <DatePicker.RangePicker showTime />
                  </Form.Item>
                  <Form.Item>
                    <Space>
                      <Button type="primary" htmlType="submit" loading={opQuery.isFetching}>
                        查询
                      </Button>
                      <Button onClick={handleOpReset} disabled={opQuery.isFetching}>
                        重置
                      </Button>
                    </Space>
                  </Form.Item>
                </Form>

                <Table<OperationLog>
                  rowKey="id"
                  dataSource={opQuery.data?.list ?? []}
                  loading={opQuery.isFetching}
                  columns={opColumns}
                  pagination={{
                    current: opQuery.data?.page ?? opPagination.page,
                    pageSize: opQuery.data?.limit ?? opPagination.limit,
                    total: opQuery.data?.total ?? 0,
                    showSizeChanger: true,
                    onChange: (page, pageSize) => setOpPagination({ page, limit: pageSize }),
                  }}
                />
              </Space>
            ),
          },
          {
            key: 'access',
            label: '访问日志',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Form layout="inline" form={accessForm} onFinish={handleAccessFilter}>
                  <Form.Item name="method" label="方法">
                    <Select allowClear placeholder="全部" options={METHOD_OPTIONS.filter((x) => x.value)} style={{ width: 120 }} />
                  </Form.Item>
                  <Form.Item name="path" label="Path">
                    <Input allowClear placeholder="包含匹配" style={{ width: 240 }} />
                  </Form.Item>
                  <Form.Item name="user_id" label="UserID">
                    <Input allowClear placeholder="精确" style={{ width: 120 }} />
                  </Form.Item>
                  <Form.Item name="status" label="状态码">
                    <Input allowClear placeholder="例如 200/401/500" style={{ width: 140 }} />
                  </Form.Item>
                  <Form.Item name="date_range" label="时间">
                    <DatePicker.RangePicker showTime />
                  </Form.Item>
                  <Form.Item>
                    <Space>
                      <Button type="primary" htmlType="submit" loading={accessQuery.isFetching}>
                        查询
                      </Button>
                      <Button onClick={handleAccessReset} disabled={accessQuery.isFetching}>
                        重置
                      </Button>
                    </Space>
                  </Form.Item>
                </Form>

                <Table<AccessLog>
                  rowKey="id"
                  dataSource={accessQuery.data?.list ?? []}
                  loading={accessQuery.isFetching}
                  columns={accessColumns}
                  pagination={{
                    current: accessQuery.data?.page ?? accessPagination.page,
                    pageSize: accessQuery.data?.limit ?? accessPagination.limit,
                    total: accessQuery.data?.total ?? 0,
                    showSizeChanger: true,
                    onChange: (page, pageSize) => setAccessPagination({ page, limit: pageSize }),
                  }}
                />
              </Space>
            ),
          },
        ]}
      />
    </Space>
  );
}
