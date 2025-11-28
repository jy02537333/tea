import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Modal, message, Form, Input } from 'antd';
import { listUsers, updateUser, disableUser, createUser, UserDetail } from '../../services/users';

const PAGE_SIZE = 20;

const Users: React.FC = () => {
  const [data, setData] = useState<UserDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [selected, setSelected] = useState<UserDetail | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [form] = Form.useForm();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [createForm] = Form.useForm();
    function openCreate() {
      setCreateModalOpen(true);
      createForm.resetFields();
    }

    function closeCreate() {
      setCreateModalOpen(false);
      createForm.resetFields();
    }

    async function handleCreate() {
      try {
        const values = await createForm.validateFields();
        await createUser(values);
        message.success('新建成功');
        closeCreate();
        fetch(page);
      } catch (e: any) {
        message.error(e?.message || '新建失败');
      }
    }
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetch(page);
  }, [page]);

  async function fetch(pageNum: number) {
    setLoading(true);
    try {
      const res = await listUsers({ page: pageNum, limit: PAGE_SIZE });
      setData(res.data);
      setTotal(res.total || 0);
    } catch (e: any) {
      message.error(e?.message || '加载失败');
    } finally {
      setLoading(false);
    }
  }

  function openDetail(record: UserDetail) {
    setSelected(record);
    setEditMode(false);
    setModalOpen(true);
    form.setFieldsValue(record);
  }

  function closeModal() {
    setModalOpen(false);
    setSelected(null);
    setEditMode(false);
    form.resetFields();
  }

  function startEdit() {
    setEditMode(true);
    if (selected) form.setFieldsValue(selected);
  }

  async function handleSave() {
    if (!selected) return;
    try {
      const values = await form.validateFields();
      await updateUser(selected.id, values);
      message.success('保存成功');
      closeModal();
      fetch(page);
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    }
  }

  async function handleDisable() {
    if (!selected) return;
    Modal.confirm({
      title: '确认禁用该用户？',
      onOk: async () => {
        try {
          await disableUser(selected.id);
          message.success('已禁用');
          closeModal();
          fetch(page);
        } catch (e: any) {
          message.error(e?.message || '操作失败');
        }
      },
    });
  }

  return (
    <div style={{ padding: 16 }}>
      <h1>用户管理</h1>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={openCreate}>新建用户</Button>
      </Space>
            <Modal
              open={createModalOpen}
              title="新建用户"
              onCancel={closeCreate}
              footer={[
                <Button key="cancel" onClick={closeCreate}>取消</Button>,
                <Button key="save" type="primary" onClick={handleCreate}>保存</Button>,
              ]}
            >
              <Form form={createForm} layout="vertical">
                <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}> <Input /> </Form.Item>
                <Form.Item name="phone" label="手机号"> <Input /> </Form.Item>
                {/* 可扩展更多字段 */}
              </Form>
            </Modal>
      <Table<UserDetail>
        rowKey="id"
        loading={loading}
        dataSource={data}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '昵称', dataIndex: 'nickname' },
          { title: '手机号', dataIndex: 'phone' },
          { title: '角色', dataIndex: 'roles', render: (roles?: string[]) => roles?.join(', ') },
          { title: '注册时间', dataIndex: 'created_at' },
          {
            title: '操作',
            render: (_, record) => (
              <Space>
                <Button size="small" onClick={() => openDetail(record)}>
                  详情
                </Button>
                <Button size="small" danger onClick={() => { setSelected(record); handleDisable(); }}>
                  禁用
                </Button>
              </Space>
            ),
          },
        ]}
        pagination={{
          current: page,
          pageSize: PAGE_SIZE,
          total,
          onChange: setPage,
          showTotal: (t) => `共 ${t} 条`,
        }}
      />
      <Modal
        open={modalOpen}
        title={editMode ? '编辑用户' : '用户详情'}
        onCancel={closeModal}
        footer={editMode ? [
          <Button key="cancel" onClick={closeModal}>取消</Button>,
          <Button key="save" type="primary" onClick={handleSave}>保存</Button>,
        ] : [
          <Button key="edit" type="primary" onClick={startEdit}>编辑</Button>,
          <Button key="close" onClick={closeModal}>关闭</Button>,
        ]}
      >
        {selected ? (
          editMode ? (
            <Form form={form} layout="vertical" initialValues={selected}>
              <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}> <Input /> </Form.Item>
              <Form.Item name="phone" label="手机号"> <Input /> </Form.Item>
              {/* 可扩展更多字段 */}
            </Form>
          ) : (
            <div>
              <p>ID: {selected.id}</p>
              <p>昵称: {selected.nickname}</p>
              <p>手机号: {selected.phone}</p>
              <p>角色: {selected.roles?.join(', ')}</p>
              <p>注册时间: {selected.created_at}</p>
            </div>
          )
        ) : null}
      </Modal>
    </div>
  );
};

export default Users;
