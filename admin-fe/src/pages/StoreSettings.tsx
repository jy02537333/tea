import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Form, Input, InputNumber, Space, Spin, Table, Typography, message, Popconfirm } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getStore, updateStore, type Store, type StorePayload, listStoreTables, createStoreTable, deleteStoreTable, type StoreTable } from '../services/stores';
import { useAuthContext } from '../hooks/useAuth';

const { Title, Text } = Typography;

function toPayload(values: Partial<Store>): StorePayload {
  return {
    name: String(values.name || '').trim(),
    address: values.address || '',
    phone: values.phone || '',
    latitude: values.latitude,
    longitude: values.longitude,
    business_hours: values.business_hours || '',
    images: values.images || '',
    status: values.status,
  };
}

export default function StoreSettingsPage() {
  const { user } = useAuthContext();
  const lockedStoreId = user?.store_id;
  const queryClient = useQueryClient();
  const [form] = Form.useForm<Partial<Store>>();
  const [tableForm] = Form.useForm<{ table_no: string; capacity?: number; note?: string }>();
  const [tableLoading, setTableLoading] = useState(false);
  const [tables, setTables] = useState<StoreTable[]>([]);

  const storeQuery = useQuery({
    queryKey: ['store-settings-detail', lockedStoreId],
    queryFn: () => getStore(lockedStoreId!),
    enabled: user?.role === 'store' && !!lockedStoreId,
  });

  useEffect(() => {
    if (storeQuery.data) {
      form.setFieldsValue(storeQuery.data);
    }
  }, [form, storeQuery.data]);

  const loadTables = async () => {
    if (!lockedStoreId) return;
    setTableLoading(true);
    try {
      const list = await listStoreTables(lockedStoreId);
      setTables(list);
    } catch (error) {
      message.error('无法获取桌号列表');
    } finally {
      setTableLoading(false);
    }
  };

  useEffect(() => {
    if (lockedStoreId) {
      loadTables();
    } else {
      setTables([]);
    }
  }, [lockedStoreId]);

  const canEdit = useMemo(() => user?.role === 'store' && !!lockedStoreId, [user?.role, lockedStoreId]);

  const saveMutation = useMutation({
    mutationFn: async (values: Partial<Store>) => {
      if (!lockedStoreId) throw new Error('未绑定门店');
      const payload = toPayload(values);
      if (!payload.name) throw new Error('门店名称不能为空');
      await updateStore(lockedStoreId, payload);
    },
    onSuccess: async () => {
      message.success('门店信息已保存');
      await queryClient.invalidateQueries({ queryKey: ['store-settings-detail', lockedStoreId] });
    },
    onError: (err: any) => {
      message.error(err?.message || '保存失败');
    },
  });

  const createTableMutation = useMutation({
    mutationFn: async (vals: { table_no: string; capacity?: number; note?: string }) => {
      if (!lockedStoreId) throw new Error('未绑定门店');
      const payload = { table_no: String(vals.table_no || '').trim(), capacity: vals.capacity, note: vals.note };
      if (!payload.table_no) throw new Error('桌号不能为空');
      await createStoreTable(lockedStoreId, payload);
    },
    onSuccess: async () => {
      message.success('新增桌号成功');
      tableForm.resetFields();
      await loadTables();
    },
    onError: (err: any) => {
      message.error(err?.message || '新增失败');
    },
  });

  const deleteTableMutation = useMutation({
    mutationFn: async (tableId: number) => {
      if (!lockedStoreId) throw new Error('未绑定门店');
      await deleteStoreTable(lockedStoreId, tableId);
    },
    onSuccess: async () => {
      message.success('删除成功');
      await loadTables();
    },
    onError: (err: any) => {
      message.error(err?.message || '删除失败');
    },
  });

  if (user?.role !== 'store') {
    return (
      <Alert
        type="warning"
        showIcon
        message="该页面仅用于门店后台"
        description="当前账号不是门店管理员（role=store）。"
      />
    );
  }

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Title level={4} style={{ margin: 0 }}>
        门店设置
        {lockedStoreId ? (
          <Text type="secondary" style={{ marginLeft: 8 }}>
            (门店ID: {lockedStoreId})
          </Text>
        ) : null}
      </Title>

      {!lockedStoreId && <Alert type="error" showIcon message="门店管理员未绑定门店（store_admins），无法维护门店信息" />}

      {lockedStoreId && (
        <Card title="门店信息" extra={<Text type="secondary">名称/地址/营业时间/联系电话等</Text>}>
          {storeQuery.isLoading && <Spin />}
          {storeQuery.isError && <Alert type="error" showIcon message="无法获取门店信息" />}

          <Form form={form} layout="vertical" onFinish={(vals) => saveMutation.mutate(vals)} disabled={!canEdit}>
            <Form.Item name="name" label="门店名称" rules={[{ required: true, message: '请输入门店名称' }]}>
              <Input placeholder="例如：茶心阁（XX店）" />
            </Form.Item>
            <Form.Item name="phone" label="联系电话">
              <Input placeholder="用于用户联系门店" />
            </Form.Item>
            <Form.Item name="address" label="门店地址">
              <Input placeholder="例如：XX省XX市XX路XX号" />
            </Form.Item>
            <Form.Item name="business_hours" label="营业时间">
              <Input placeholder="例如：09:00-21:00" />
            </Form.Item>
            <Form.Item name="images" label="图片（门头/环境/证照，逗号分隔 URL）">
              <Input.TextArea rows={3} placeholder="https://... , https://..." />
            </Form.Item>

            <Space>
              <Button type="primary" htmlType="submit" loading={saveMutation.isPending} disabled={!lockedStoreId}>
                保存
              </Button>
              <Button onClick={() => form.resetFields()} disabled={saveMutation.isPending}>
                重置
              </Button>
            </Space>
          </Form>
        </Card>
      )}

      {lockedStoreId && (
        <Card title="门店桌号管理" extra={<Text type="secondary">新增/删除堂食桌号</Text>}>
          <Form form={tableForm} layout="inline" onFinish={(vals) => createTableMutation.mutate(vals)}>
            <Form.Item name="table_no" label="桌号" rules={[{ required: true, message: '请输入桌号' }]}> 
              <Input placeholder="例如 A12 或 5 号桌" style={{ width: 200 }}/>
            </Form.Item>
            <Form.Item name="capacity" label="可容纳人数"> 
              <InputNumber placeholder="可选" style={{ width: 160 }} min={0} />
            </Form.Item>
            <Form.Item name="note" label="备注"> 
              <Input placeholder="可选" style={{ width: 200 }}/>
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={createTableMutation.isPending}>新增桌号</Button>
                <Button onClick={() => tableForm.resetFields()} disabled={createTableMutation.isPending}>重置</Button>
              </Space>
            </Form.Item>
          </Form>

          <Table<StoreTable>
            rowKey="id"
            style={{ marginTop: 16 }}
            loading={tableLoading}
            dataSource={tables}
            columns={[
              { title: 'ID', dataIndex: 'id', width: 80 },
              { title: '桌号', dataIndex: 'table_no', width: 160 },
              { title: '可容纳人数', dataIndex: 'capacity', width: 120 },
              { title: '备注', dataIndex: 'note' },
              {
                title: '操作',
                key: 'actions',
                width: 160,
                render: (_, record) => (
                  <Space>
                    <Popconfirm title="确认删除该桌号？" onConfirm={() => deleteTableMutation.mutate(record.id)}>
                      <Button danger type="link">删除</Button>
                    </Popconfirm>
                  </Space>
                ),
              },
            ]}
            pagination={false}
          />
        </Card>
      )}
    </Space>
  );
}
