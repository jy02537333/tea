import { useEffect, useMemo, useState } from 'react';
import { Button, Descriptions, Divider, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AdminUser,
  CreateAdminUserPayload,
  UpdateAdminUserPayload,
  createAdminUser,
  getAdminUsers,
  getUserPermissions,
  resetAdminUserPassword,
  updateAdminUser,
} from '../services/users';

const STATUS_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '启用', value: 1 },
  { label: '停用', value: 2 },
];

const EDIT_STATUS_OPTIONS = [
  { label: '启用', value: 1 },
  { label: '停用', value: 2 },
];

const ROLE_OPTIONS = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
];

const { Title } = Typography;

interface PasswordFormValues {
  new_password: string;
  confirm_password: string;
}

interface CreateFormValues extends CreateAdminUserPayload {
  confirm_password: string;
}

const statusTag = (status?: number) => {
  if (status === 1) return <Tag color="green">启用</Tag>;
  if (status === 2) return <Tag color="red">停用</Tag>;
  return <Tag>未知</Tag>;
};

interface FilterValues {
  user_id?: number;
  keyword?: string;
  status?: number;
}

export default function UsersPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<FilterValues>({});
  const [filterForm] = Form.useForm<FilterValues>();
  const [editForm] = Form.useForm<UpdateAdminUserPayload>();
  const [passwordForm] = Form.useForm<PasswordFormValues>();
  const [createForm] = Form.useForm<CreateFormValues>();
  const [drawerUser, setDrawerUser] = useState<AdminUser | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const usersQuery = useQuery({
    queryKey: ['adminUsers', filters.user_id ?? 'all'],
    queryFn: () => getAdminUsers(filters.user_id ? { user_id: filters.user_id } : {}),
  });

  useEffect(() => {
    if (drawerUser) {
      editForm.setFieldsValue({
        nickname: drawerUser.nickname,
        phone: drawerUser.phone,
        role: drawerUser.role,
        status: drawerUser.status,
      });
      passwordForm.resetFields();
    } else {
      editForm.resetFields();
      passwordForm.resetFields();
    }
  }, [drawerUser, editForm, passwordForm]);

  const tableData = useMemo(() => {
    const list: AdminUser[] = usersQuery.data ?? [];
    const keyword = filters.keyword?.trim().toLowerCase();
    return list.filter((user) => {
      if (filters.status && user.status !== filters.status) return false;
      if (keyword && !`${user.nickname ?? ''} ${user.phone ?? ''}`.toLowerCase().includes(keyword)) {
        return false;
      }
      return true;
    });
  }, [usersQuery.data, filters.keyword, filters.status]);

  const openDrawer = async (user: AdminUser) => {
    setPermissions([]);
    setDrawerUser(user);
    try {
      const perms = await getUserPermissions(user.id);
      setPermissions(perms);
    } catch (error) {
      setPermissions([]);
    }
  };

  const updateMutation = useMutation<AdminUser, any, UpdateAdminUserPayload>({
    mutationFn: (payload: UpdateAdminUserPayload) => {
      if (!drawerUser) {
        return Promise.reject(new Error('请选择用户'));
      }
      return updateAdminUser(drawerUser.id, payload);
    },
    onSuccess: (data: AdminUser) => {
      message.success('用户信息已更新');
      setDrawerUser(data);
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '更新失败');
    },
  });

  const passwordMutation = useMutation<{ message: string }, any, string>({
    mutationFn: (newPassword: string) => {
      if (!drawerUser) {
        return Promise.reject(new Error('请选择用户'));
      }
      return resetAdminUserPassword(drawerUser.id, newPassword);
    },
    onSuccess: () => {
      message.success('密码已重置');
      passwordForm.resetFields();
    },
    onError: (error: any) => {
      message.error(error?.message || '重置失败');
    },
  });

  const createMutation = useMutation<AdminUser, any, CreateAdminUserPayload>({
    mutationFn: (payload: CreateAdminUserPayload) => createAdminUser(payload),
    onSuccess: () => {
      message.success('用户已创建');
      setCreateModalOpen(false);
      createForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '创建失败');
    },
  });

  type ToggleStatusVariables = { userId: number; status: number };
  const toggleStatusMutation = useMutation<AdminUser, any, ToggleStatusVariables>({
    mutationFn: ({ userId, status }: ToggleStatusVariables) => updateAdminUser(userId, { status }),
    onSuccess: (data: AdminUser) => {
      message.success('状态已更新');
      setDrawerUser((prev) => (prev && prev.id === data.id ? data : prev));
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '操作失败');
    },
  });

  const togglingUserId = toggleStatusMutation.variables?.userId;

  const handleSubmitEdit = async (values: UpdateAdminUserPayload) => {
    if (!drawerUser) return;
    await updateMutation.mutateAsync(values);
  };

  const handleSubmitPassword = async (values: PasswordFormValues) => {
    if (!drawerUser) return;
    if (values.new_password !== values.confirm_password) {
      passwordForm.setFields([
        {
          name: 'confirm_password',
          errors: ['两次输入的密码不一致'],
        },
      ]);
      return;
    }
    await passwordMutation.mutateAsync(values.new_password);
  };

  const handleSubmitCreate = async (values: CreateFormValues) => {
    if (values.password !== values.confirm_password) {
      createForm.setFields([
        {
          name: 'confirm_password',
          errors: ['两次输入的密码不一致'],
        },
      ]);
      return;
    }

    const payload: CreateAdminUserPayload = {
      username: values.username.trim(),
      password: values.password,
      phone: values.phone.trim(),
      nickname: values.nickname?.trim(),
      role: values.role?.trim(),
      status: values.status,
    };

    await createMutation.mutateAsync(payload);
  };

  const handleToggleStatus = (user: AdminUser) => {
    if (toggleStatusMutation.isPending && togglingUserId === user.id) {
      return;
    }
    const nextStatus = user.status === 1 ? 2 : 1;
    toggleStatusMutation.mutate({ userId: user.id, status: nextStatus });
  };

  const columns: ColumnsType<AdminUser> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '昵称', dataIndex: 'nickname' },
    { title: '手机号', dataIndex: 'phone', width: 140 },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      render: (role?: string) => (role ? <Tag>{role}</Tag> : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (status?: number) => statusTag(status),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 200,
      render: (val?: string) => (val ? new Date(val).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space size={8}>
          <Button type="link" onClick={() => openDrawer(record)}>
            查看详情
          </Button>
          <Button
            type="link"
            danger={record.status === 1}
            onClick={() => handleToggleStatus(record)}
            loading={toggleStatusMutation.isPending && togglingUserId === record.id}
          >
            {record.status === 1 ? '禁用' : '启用'}
          </Button>
        </Space>
      ),
    },
  ];

  const handleFilter = (values: FilterValues) => {
    const next: FilterValues = { ...values };
    if (typeof next.user_id === 'string') {
      next.user_id = next.user_id ? Number(next.user_id) : undefined;
    }
    setFilters(next);
  };

  const handleReset = () => {
    filterForm.resetFields();
    setFilters({});
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Form layout="inline" form={filterForm} onFinish={handleFilter}>
          <Form.Item name="user_id" label="用户ID">
            <InputNumber min={1} style={{ width: 180 }} placeholder="精确 ID" />
          </Form.Item>
          <Form.Item name="keyword" label="关键字">
            <Input allowClear placeholder="昵称/手机号" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear style={{ width: 140 }} options={STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={usersQuery.isFetching}>
                筛选
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
        <Space>
          <Button onClick={() => usersQuery.refetch()} loading={usersQuery.isFetching}>
            刷新
          </Button>
          <Button
            type="primary"
            onClick={() => {
              createForm.resetFields();
              setCreateModalOpen(true);
            }}
          >
            新增用户
          </Button>
        </Space>
      </Space>

      <Table
        bordered
        loading={usersQuery.isLoading}
        rowKey="id"
        dataSource={tableData}
        columns={columns}
        pagination={false}
      />

      <Drawer
        title={drawerUser ? `用户详情 · #${drawerUser.id}` : '用户详情'}
        width={520}
        open={!!drawerUser}
        onClose={() => {
          setDrawerUser(null);
          setPermissions([]);
          editForm.resetFields();
          passwordForm.resetFields();
        }}
        destroyOnClose
      >
        {drawerUser && (
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="昵称">{drawerUser.nickname || '-'}</Descriptions.Item>
              <Descriptions.Item label="手机号">{drawerUser.phone || '-'}</Descriptions.Item>
              <Descriptions.Item label="角色">{drawerUser.role || '-'}</Descriptions.Item>
              <Descriptions.Item label="状态">{statusTag(drawerUser.status)}</Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {drawerUser.created_at ? new Date(drawerUser.created_at).toLocaleString() : '-'}
              </Descriptions.Item>
            </Descriptions>
            <div>
              <strong>权限列表</strong>
              <div style={{ marginTop: 8 }}>
                {permissions.length === 0 && <span style={{ color: '#888' }}>暂无权限数据</span>}
                {permissions.map((perm) => (
                  <Tag key={perm} color="blue" style={{ marginBottom: 8 }}>
                    {perm}
                  </Tag>
                ))}
              </div>
            </div>
            <Divider />
            <Title level={5}>编辑资料</Title>
            <Form layout="vertical" form={editForm} onFinish={handleSubmitEdit}>
              <Form.Item name="nickname" label="昵称">
                <Input placeholder="请输入昵称" allowClear />
              </Form.Item>
              <Form.Item name="phone" label="手机号">
                <Input placeholder="请输入手机号" allowClear />
              </Form.Item>
              <Form.Item name="role" label="角色">
                <Input placeholder="例如 admin / user" allowClear />
              </Form.Item>
              <Form.Item name="status" label="状态">
                <Select allowClear placeholder="请选择" options={EDIT_STATUS_OPTIONS} />
              </Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>
                  保存修改
                </Button>
                <Button onClick={() => editForm.resetFields()} disabled={updateMutation.isPending}>
                  重置表单
                </Button>
              </Space>
            </Form>
            <Divider />
            <Title level={5}>重置密码</Title>
            <Form layout="vertical" form={passwordForm} onFinish={handleSubmitPassword}>
              <Form.Item
                name="new_password"
                label="新密码"
                rules={[{ required: true, message: '请输入新密码' }, { min: 6, message: '至少 6 位字符' }]}
              >
                <Input.Password placeholder="输入新密码" allowClear />
              </Form.Item>
              <Form.Item
                name="confirm_password"
                label="确认新密码"
                dependencies={['new_password']}
                rules={[{ required: true, message: '请再次输入新密码' }]}
              >
                <Input.Password placeholder="再次输入以确认" allowClear />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={passwordMutation.isPending}>
                确认重置
              </Button>
            </Form>
          </Space>
        )}
      </Drawer>

      <Modal
        title="新增用户"
        open={createModalOpen}
        onCancel={() => {
          if (!createMutation.isPending) {
            setCreateModalOpen(false);
            createForm.resetFields();
          }
        }}
        onOk={() => createForm.submit()}
        confirmLoading={createMutation.isPending}
        destroyOnClose
      >
        <Form layout="vertical" form={createForm} onFinish={handleSubmitCreate} initialValues={{ status: 1 }}>
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="登录用户名" autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="phone"
            label="手机号"
            rules={[{ required: true, message: '请输入手机号' }]}
          >
            <Input placeholder="用于找回 / 唯一标识" autoComplete="off" />
          </Form.Item>
          <Form.Item name="nickname" label="昵称">
            <Input placeholder="可选" autoComplete="off" />
          </Form.Item>
          <Form.Item name="role" label="角色">
            <Select allowClear placeholder="请选择角色" options={ROLE_OPTIONS} />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select allowClear placeholder="默认为启用" options={EDIT_STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="password"
            label="初始密码"
            rules={[{ required: true, message: '请输入初始密码' }, { min: 6, message: '至少 6 位字符' }]}
          >
            <Input.Password placeholder="设置登录密码" autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认密码"
            dependencies={['password']}
            rules={[{ required: true, message: '请再次输入密码' }]}
          >
            <Input.Password placeholder="再次输入以确认" autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
