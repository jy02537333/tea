import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Divider, Form, InputNumber, Space, Spin, Table, Tag, Transfer, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Permission,
  Role,
  assignPermissions,
  assignRoleToUser,
  getPermissions,
  getRolePermissions,
  getRoles,
  revokePermission,
  revokeRoleFromUser,
} from '../services/rbac';

interface AssignRoleFormValues {
  user_id?: number;
  role_id?: number;
}

export default function RbacPage() {
  const queryClient = useQueryClient();
  const [selectedRoleId, setSelectedRoleId] = useState<number | null>(null);
  const [targetKeys, setTargetKeys] = useState<string[]>([]);
  const [assignForm] = Form.useForm<AssignRoleFormValues>();

  const rolesQuery = useQuery({ queryKey: ['rbacRoles'], queryFn: () => getRoles() });
  const permissionsQuery = useQuery({ queryKey: ['rbacPermissions'], queryFn: () => getPermissions() });

  const rolePermissionsQuery = useQuery({
    queryKey: ['rbacRolePermissions', selectedRoleId],
    queryFn: () => getRolePermissions(selectedRoleId!),
    enabled: !!selectedRoleId,
  });

  useEffect(() => {
    if (!selectedRoleId) {
      setTargetKeys([]);
      return;
    }
    const ids = rolePermissionsQuery.data?.map((rp) => String(rp.permission_id)) ?? [];
    setTargetKeys(ids);
    assignForm.setFieldsValue({ role_id: selectedRoleId || undefined });
  }, [rolePermissionsQuery.data, selectedRoleId, assignForm]);

  const transferData = useMemo(() => {
    if (!permissionsQuery.data) return [];
    return permissionsQuery.data.map((perm: Permission) => ({
      key: String(perm.id),
      title: perm.display_name || perm.name,
      description: perm.module ? `${perm.module} · ${perm.action ?? ''}` : perm.name,
    }));
  }, [permissionsQuery.data]);

  const savePermissionsMutation = useMutation({
    mutationFn: async () => {
      if (!selectedRoleId) return;
      const currentKeys = new Set((rolePermissionsQuery.data ?? []).map((rp) => String(rp.permission_id)));
      const targetKeySet = new Set(targetKeys);
      const added = Array.from(targetKeySet).filter((key) => !currentKeys.has(key)).map((key) => Number(key));
      const removed = Array.from(currentKeys).filter((key) => !targetKeySet.has(key)).map((key) => Number(key));
      if (!added.length && !removed.length) {
        message.info('权限未变更');
        return;
      }
      if (added.length) {
        await assignPermissions(selectedRoleId, added);
      }
      for (const id of removed) {
        await revokePermission(selectedRoleId, id);
      }
      await queryClient.invalidateQueries({ queryKey: ['rbacRolePermissions', selectedRoleId] });
      message.success('权限已更新');
    },
  });

  const assignRoleMutation = useMutation({
    mutationFn: async (values: AssignRoleFormValues & { action: 'assign' | 'revoke' }) => {
      if (!values.user_id || !values.role_id) {
        throw new Error('请填写用户 ID 与角色 ID');
      }
      if (values.action === 'assign') {
        await assignRoleToUser(values.user_id, values.role_id);
        return '角色已分配';
      }
      await revokeRoleFromUser(values.user_id, values.role_id);
      return '角色已移除';
    },
    onSuccess: (msg) => {
      message.success(msg ?? '操作成功');
    },
    onError: (error: any) => {
      message.error(error?.message || '操作失败');
    },
  });

  const columns: ColumnsType<Role> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '角色标识', dataIndex: 'name' },
    { title: '显示名称', dataIndex: 'display_name' },
    {
      title: '状态',
      key: 'status',
      render: (_, record) => (record.id === selectedRoleId ? <Tag color="green">当前</Tag> : '-'),
      width: 120,
    },
  ];

  const handleAssignRole = async (action: 'assign' | 'revoke') => {
    try {
      const values = await assignForm.validateFields();
      await assignRoleMutation.mutateAsync({ ...values, action });
    } catch (error) {
      // antd 已提示
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space align="start" size={16} wrap style={{ width: '100%' }}>
        <Card title="角色列表" style={{ flex: 1, minWidth: 360 }} extra={<Button onClick={() => rolesQuery.refetch()}>刷新</Button>}>
          <Table
            rowKey="id"
            loading={rolesQuery.isLoading}
            dataSource={rolesQuery.data}
            columns={columns}
            pagination={false}
            size="small"
            onRow={(record) => ({
              onClick: () => setSelectedRoleId(record.id),
              style: { cursor: 'pointer', background: record.id === selectedRoleId ? '#f6ffed' : undefined },
            })}
          />
        </Card>

        <Card
          title={selectedRoleId ? `角色权限 · #${selectedRoleId}` : '角色权限'}
          style={{ flex: 2, minWidth: 420 }}
          extra={
            <Button type="primary" disabled={!selectedRoleId} loading={savePermissionsMutation.isPending} onClick={() => savePermissionsMutation.mutate()}>
              保存权限
            </Button>
          }
        >
          {!selectedRoleId && <Alert message="请从左侧选择一个角色" type="info" showIcon />}
          {selectedRoleId && (
            <>
              <Spin spinning={permissionsQuery.isLoading || rolePermissionsQuery.isLoading}>
                <Transfer
                  dataSource={transferData}
                  titles={['全部权限', '已拥有']}
                  targetKeys={targetKeys}
                  onChange={(nextKeys) => setTargetKeys(nextKeys as string[])}
                  render={(item) => item.title}
                  listStyle={{ width: 220, height: 360 }}
                />
              </Spin>
              <Divider />
              <Form form={assignForm} layout="inline" style={{ rowGap: 12 }}>
                <Form.Item label="用户 ID" name="user_id" rules={[{ required: true, message: '请输入用户 ID' }]}>
                  <InputNumber min={1} style={{ width: 200 }} placeholder="用户 ID" />
                </Form.Item>
                <Form.Item label="角色 ID" name="role_id" initialValue={selectedRoleId} rules={[{ required: true, message: '请输入角色 ID' }]}>
                  <InputNumber min={1} style={{ width: 200 }} placeholder="角色 ID" />
                </Form.Item>
                <Form.Item>
                  <Space>
                    <Button type="primary" onClick={() => handleAssignRole('assign')} loading={assignRoleMutation.isPending}>
                      分配角色
                    </Button>
                    <Button danger onClick={() => handleAssignRole('revoke')} loading={assignRoleMutation.isPending}>
                      移除角色
                    </Button>
                  </Space>
                </Form.Item>
              </Form>
            </>
          )}
        </Card>
      </Space>
    </Space>
  );
}
