import { useEffect, useMemo, useState, lazy, Suspense } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, Drawer, Form, Input, Space, Switch, Tabs, Typography, message, Select, Modal, Table } from 'antd';
import { useLocation, useNavigate } from 'react-router-dom';
import { listSystemConfigs, upsertSystemConfigs } from '../services/systemConfigs';
import { useAuthContext } from '../hooks/useAuth';
import type { AdminUser, CreateAdminUserPayload, UpdateAdminUserPayload } from '../services/users';
import { createAdminUser, getAdminUsers, resetAdminUserPassword, updateAdminUser } from '../services/users';
import { assignRoleToUser, getRoles, Role } from '../services/rbac';

const RbacPage = lazy(() => import('./Rbac'));

const { Title } = Typography;

const BASIC_KEYS = ['site_logo_url', 'site_phone', 'site_copyright'] as const;
const CONTENT_KEYS = ['content_about', 'content_help', 'content_privacy', 'content_terms'] as const;

type FormValues = Record<(typeof BASIC_KEYS)[number] | (typeof CONTENT_KEYS)[number], string>;

function toMap(list: { config_key: string; config_value: string }[]) {
  const map: Record<string, string> = {};
  for (const item of list) map[item.config_key] = item.config_value ?? '';
  return map;
}

export default function SystemSettingsPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { hasPermission } = useAuthContext();
  const canManage = hasPermission('system:config:manage');

  const searchParams = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const initialTab = searchParams.get('tab') || 'basic';
  const [activeTab, setActiveTab] = useState<string>(initialTab);

  useEffect(() => {
    setActiveTab(initialTab);
  }, [initialTab]);

  const allKeys = useMemo(() => [...BASIC_KEYS, ...CONTENT_KEYS] as string[], []);
  const [form] = Form.useForm<FormValues>();

  const query = useQuery({
    queryKey: ['system-configs', allKeys.join(',')],
    queryFn: () => listSystemConfigs({ keys: allKeys }),
  });

  const initialValues = useMemo(() => {
    const list = query.data?.list ?? [];
    const map = toMap(list);
    const v: any = {};
    for (const k of allKeys) v[k] = map[k] ?? '';
    return v as FormValues;
  }, [allKeys, query.data?.list]);

  const saveMutation = useMutation({
    mutationFn: async (values: FormValues) => {
      const items = allKeys.map((k) => ({
        config_key: k,
        config_value: String((values as any)[k] ?? ''),
        config_type: 'string',
        status: 1,
      }));
      return upsertSystemConfigs(items);
    },
    onSuccess: async () => {
      message.success('已保存');
      await queryClient.invalidateQueries({ queryKey: ['system-configs'] });
    },
    onError: (err: any) => message.error(err?.message || '保存失败'),
  });

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          系统设置
        </Title>
        <Space>
          <Button onClick={() => query.refetch()} loading={query.isFetching}>
            刷新
          </Button>
          <Button type="primary" onClick={() => form.submit()} loading={saveMutation.isPending} disabled={!canManage}>
            保存
          </Button>
        </Space>
      </Space>

      {!canManage && (
        <Alert
          type="warning"
          showIcon
          message="当前账号缺少 system:config:manage 权限，已禁用保存按钮。"
        />
      )}

      {query.isError && (
        <Alert
          type="error"
          showIcon
          message="加载系统配置失败"
          description={(query.error as any)?.message || '请确认已登录且具备 system:config:view 权限'}
        />
      )}

      <Form
        form={form}
        layout="vertical"
        initialValues={initialValues}
        onFinish={(values) => saveMutation.mutate(values)}
      >
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key);
            const params = new URLSearchParams(location.search);
            if (key === 'basic') params.delete('tab');
            else params.set('tab', key);
            navigate({ search: params.toString() ? `?${params.toString()}` : '' }, { replace: true });
          }}
          items={[
            {
              key: 'basic',
              label: '基础配置',
              children: (
                <Card>
                  <Form.Item label="Logo URL" name="site_logo_url">
                    <Input placeholder="https://..." allowClear />
                  </Form.Item>
                  <Form.Item label="客服电话" name="site_phone">
                    <Input placeholder="例如 400-xxx-xxxx" allowClear />
                  </Form.Item>
                  <Form.Item label="版权信息" name="site_copyright">
                    <Input placeholder="例如 © 茶心阁" allowClear />
                  </Form.Item>
                </Card>
              ),
            },
            {
              key: 'content',
              label: '内容管理',
              children: (
                <Card>
                  <Form.Item label="关于我们" name="content_about">
                    <Input.TextArea rows={6} placeholder="支持纯文本/Markdown（如后端渲染可再扩展）" />
                  </Form.Item>
                  <Form.Item label="帮助文档" name="content_help">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                  <Form.Item label="隐私政策" name="content_privacy">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                  <Form.Item label="用户协议" name="content_terms">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                </Card>
              ),
            },
            {
              key: 'admins',
              label: '管理员',
              children: <AdminManagementPanel />,
            },
          ]}
        />
      </Form>
    </Space>
  );
}

function AdminManagementPanel() {
  const queryClient = useQueryClient();
  const { hasPermission } = useAuthContext();
  const canManageAdmins = hasPermission('admin:user:manage') || hasPermission('system:config:manage');

  const usersQuery = useQuery({
    queryKey: ['adminUsers'],
    queryFn: () => getAdminUsers(),
  });

  const rolesQuery = useQuery({
    queryKey: ['rbacRoles'],
    queryFn: () => getRoles(),
  });

  const roleOptions = useMemo(
    () => (rolesQuery.data ?? []).map((r: Role) => ({ label: r.display_name || r.name, value: r.id })),
    [rolesQuery.data]
  );

  const [tab, setTab] = useState<'roles' | 'admins'>('admins');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<AdminUser | null>(null);
  const [form] = Form.useForm<any>();

  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditing(null);
    form.resetFields();
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ status: 1 });
    setDrawerOpen(true);
  };

  const openEdit = (user: AdminUser) => {
    setEditing(user);
    form.resetFields();
    form.setFieldsValue({
      nickname: user.nickname,
      phone: user.phone,
      status: user.status ?? 1,
      rbac_role_id: undefined,
    });
    setDrawerOpen(true);
  };

  const saveMutation = useMutation({
    mutationFn: async (values: any) => {
      const status = values.status ? 1 : 2;
      if (editing) {
        const payload: UpdateAdminUserPayload = {
          nickname: values.nickname?.trim() || undefined,
          phone: values.phone?.trim() || undefined,
          status,
        };
        await updateAdminUser(editing.id, payload);
        if (values.rbac_role_id) {
          await assignRoleToUser(editing.id, Number(values.rbac_role_id));
        }
        return;
      }
      const payload: CreateAdminUserPayload = {
        username: String(values.username ?? '').trim(),
        password: String(values.password ?? ''),
        phone: String(values.phone ?? '').trim(),
        nickname: values.nickname?.trim() || undefined,
        status,
      };
      const created = await createAdminUser(payload);
      if (values.rbac_role_id) {
        await assignRoleToUser(created.id, Number(values.rbac_role_id));
      }
    },
    onSuccess: async () => {
      message.success(editing ? '管理员已更新' : '管理员已创建');
      await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
      closeDrawer();
    },
    onError: (err: any) => message.error(err?.message || '保存失败'),
  });

  const [resetModalOpen, setResetModalOpen] = useState(false);
  const [resetTarget, setResetTarget] = useState<AdminUser | null>(null);
  const [resetForm] = Form.useForm<{ new_password: string }>();

  const resetMutation = useMutation({
    mutationFn: async (payload: { userId: number; newPassword: string }) => {
      await resetAdminUserPassword(payload.userId, payload.newPassword);
    },
    onSuccess: () => {
      message.success('密码已重置');
      setResetModalOpen(false);
      setResetTarget(null);
      resetForm.resetFields();
    },
    onError: (err: any) => message.error(err?.message || '重置失败'),
  });

  const dataSource = usersQuery.data ?? [];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 90 },
    { title: '昵称', dataIndex: 'nickname', width: 160, render: (v: string) => v || '-' },
    { title: '手机号', dataIndex: 'phone', width: 160, render: (v: string) => v || '-' },
    { title: '角色', dataIndex: 'role', width: 120, render: (v: string) => v || '-' },
    { title: '状态', dataIndex: 'status', width: 100, render: (v: number) => (v === 2 ? '禁用' : '启用') },
    { title: '创建时间', dataIndex: 'created_at', width: 180, render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_: any, record: AdminUser) => (
        <Space>
          <Button type="link" onClick={() => openEdit(record)} disabled={!canManageAdmins}>
            编辑
          </Button>
          <Button
            type="link"
            onClick={() => {
              setResetTarget(record);
              resetForm.resetFields();
              setResetModalOpen(true);
            }}
            disabled={!canManageAdmins}
          >
            重置密码
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Tabs
        activeKey={tab}
        onChange={(key) => setTab(key as any)}
        items={[
          {
            key: 'admins',
            label: '管理员列表',
            children: (
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
                  <Typography.Title level={5} style={{ margin: 0 }}>
                    管理员
                  </Typography.Title>
                  <Space>
                    <Button onClick={() => usersQuery.refetch()} loading={usersQuery.isFetching}>
                      刷新
                    </Button>
                    <Button type="primary" onClick={openCreate} disabled={!canManageAdmins}>
                      新增管理员
                    </Button>
                  </Space>
                </Space>

                {usersQuery.isError && (
                  <Alert type="error" showIcon message="加载管理员列表失败" description={(usersQuery.error as any)?.message || ''} />
                )}

                <Card>
                  <Table
                    rowKey="id"
                    loading={usersQuery.isLoading}
                    dataSource={dataSource}
                    columns={columns as any}
                    pagination={false}
                    size="small"
                  />
                </Card>
              </Space>
            ),
          },
          {
            key: 'roles',
            label: '权限组',
            children: (
              <Suspense fallback={<div style={{ padding: 8 }}>加载中...</div>}>
                <RbacPage />
              </Suspense>
            ),
          },
        ]}
      />

      <Drawer
        title={editing ? `编辑管理员 · #${editing.id}` : '新增管理员'}
        width={520}
        open={drawerOpen}
        onClose={closeDrawer}
        destroyOnClose
        extra={
          <Space>
            <Button onClick={closeDrawer}>取消</Button>
            <Button type="primary" onClick={() => form.submit()} loading={saveMutation.isPending} disabled={!canManageAdmins}>
              保存
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => saveMutation.mutate(values)}
          initialValues={{ status: 1 }}
        >
          {!editing && (
            <>
              <Form.Item label="用户名" name="username" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input allowClear placeholder="用于登录" />
              </Form.Item>
              <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
                <Input.Password placeholder="请输入初始密码" />
              </Form.Item>
            </>
          )}

          <Form.Item label="手机号" name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input allowClear placeholder="手机号" />
          </Form.Item>
          <Form.Item label="昵称" name="nickname">
            <Input allowClear placeholder="可选" />
          </Form.Item>

          <Form.Item label="状态" name="status" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <Form.Item label="分配权限组（RBAC 角色）" name="rbac_role_id">
            <Select
              allowClear
              placeholder={rolesQuery.isLoading ? '加载角色...' : '选择角色（可选）'}
              loading={rolesQuery.isLoading}
              options={roleOptions}
            />
          </Form.Item>
        </Form>
      </Drawer>

      <Modal
        title={resetTarget ? `重置密码 · #${resetTarget.id}` : '重置密码'}
        open={resetModalOpen}
        okText="确认重置"
        cancelText="取消"
        confirmLoading={resetMutation.isPending}
        onCancel={() => {
          setResetModalOpen(false);
          setResetTarget(null);
        }}
        onOk={async () => {
          const values = await resetForm.validateFields();
          if (!resetTarget) return;
          await resetMutation.mutateAsync({ userId: resetTarget.id, newPassword: values.new_password });
        }}
        destroyOnClose
      >
        <Form form={resetForm} layout="vertical">
          <Form.Item label="新密码" name="new_password" rules={[{ required: true, message: '请输入新密码' }]}>
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
