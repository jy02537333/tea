import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, Form, Input, InputNumber, Modal, Space, Spin, Table, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState } from 'react';
import {
  MembershipPackage,
  MembershipPackagePayload,
  PartnerLevel,
  PartnerLevelPayload,
  createMembershipPackage,
  createPartnerLevel,
  deleteMembershipPackage,
  deletePartnerLevel,
  listMembershipPackages,
  listPartnerLevels,
  updateMembershipPackage,
  updatePartnerLevel,
} from '../services/membership';

interface MembershipFormValues {
  name: string;
  price: number;
  tea_coin_award?: number;
  discount_rate?: number;
  purchase_discount_rate?: number;
  direct_commission_rate?: number;
  team_commission_rate?: number;
  upgrade_reward_rate?: number;
  type?: string;
}

interface PartnerLevelFormValues {
  name: string;
  purchase_discount_rate?: number;
  direct_commission_rate?: number;
  team_commission_rate?: number;
  upgrade_reward_rate?: number;
  note?: string;
}

export default function MembershipConfigPage() {
  const queryClient = useQueryClient();

  // 会员套餐
  const [pkgModalOpen, setPkgModalOpen] = useState(false);
  const [editingPkg, setEditingPkg] = useState<MembershipPackage | null>(null);
  const [pkgForm] = Form.useForm<MembershipFormValues>();
  const [pkgPagination, setPkgPagination] = useState({ page: 1, limit: 20 });
  const [pkgTypeFilter, setPkgTypeFilter] = useState<string | undefined>(undefined);

  const packagesQuery = useQuery({
    queryKey: ['membership-packages', pkgPagination.page, pkgPagination.limit, pkgTypeFilter ?? 'all'],
    queryFn: () => listMembershipPackages({ page: pkgPagination.page, limit: pkgPagination.limit, type: pkgTypeFilter }),
  });

  const createPkgMutation = useMutation({
    mutationFn: async (values: MembershipFormValues) => {
      const payload: MembershipPackagePayload = values;
      return createMembershipPackage(payload);
    },
    onSuccess: () => {
      message.success('会员套餐已保存');
      setPkgModalOpen(false);
      setEditingPkg(null);
      pkgForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['membership-packages'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '保存会员套餐失败');
    },
  });

  const updatePkgMutation = useMutation({
    mutationFn: async (values: MembershipFormValues) => {
      if (!editingPkg?.id) throw new Error('未选择套餐');
      const payload: Partial<MembershipPackagePayload> = values;
      return updateMembershipPackage(editingPkg.id, payload);
    },
    onSuccess: () => {
      message.success('会员套餐已更新');
      setPkgModalOpen(false);
      setEditingPkg(null);
      pkgForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['membership-packages'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新会员套餐失败');
    },
  });

  const deletePkgMutation = useMutation({
    mutationFn: async (pkg: MembershipPackage) => deleteMembershipPackage(pkg.id),
    onSuccess: () => {
      message.success('已删除会员套餐');
      queryClient.invalidateQueries({ queryKey: ['membership-packages'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '删除会员套餐失败');
    },
  });

  const pkgColumns: ColumnsType<MembershipPackage> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      sorter: (a, b) => a.id - b.id,
      defaultSortOrder: 'descend',
    },
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '价格', dataIndex: 'price', width: 120 },
    { title: '赠送茶币', dataIndex: 'tea_coin_award', width: 120 },
    { title: '会员折扣率', dataIndex: 'discount_rate', width: 120 },
    { title: '拿货折扣率', dataIndex: 'purchase_discount_rate', width: 120 },
    { title: '直推佣金率', dataIndex: 'direct_commission_rate', width: 120 },
    { title: '团队佣金率', dataIndex: 'team_commission_rate', width: 120 },
    { title: '升级奖励率', dataIndex: 'upgrade_reward_rate', width: 120 },
    { title: '类型', dataIndex: 'type', width: 120 },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setEditingPkg(record);
              pkgForm.setFieldsValue({
                name: record.name,
                price: Number(record.price || 0),
                tea_coin_award: record.tea_coin_award ? Number(record.tea_coin_award) : undefined,
                discount_rate: record.discount_rate ? Number(record.discount_rate) : undefined,
                purchase_discount_rate: record.purchase_discount_rate ? Number(record.purchase_discount_rate) : undefined,
                direct_commission_rate: record.direct_commission_rate ? Number(record.direct_commission_rate) : undefined,
                team_commission_rate: record.team_commission_rate ? Number(record.team_commission_rate) : undefined,
                upgrade_reward_rate: record.upgrade_reward_rate ? Number(record.upgrade_reward_rate) : undefined,
                type: record.type,
              });
              setPkgModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            danger
            loading={deletePkgMutation.isPending}
            onClick={() => deletePkgMutation.mutate(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  // 合伙人等级
  const [levelModalOpen, setLevelModalOpen] = useState(false);
  const [editingLevel, setEditingLevel] = useState<PartnerLevel | null>(null);
  const [levelForm] = Form.useForm<PartnerLevelFormValues>();
  const [levelPagination, setLevelPagination] = useState({ page: 1, limit: 20 });

  const levelsQuery = useQuery({
    queryKey: ['partner-levels', levelPagination.page, levelPagination.limit],
    queryFn: () => listPartnerLevels({ page: levelPagination.page, limit: levelPagination.limit }),
  });

  const createLevelMutation = useMutation({
    mutationFn: async (values: PartnerLevelFormValues) => {
      const payload: PartnerLevelPayload = values;
      return createPartnerLevel(payload);
    },
    onSuccess: () => {
      message.success('合伙人等级已保存');
      setLevelModalOpen(false);
      setEditingLevel(null);
      levelForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['partner-levels'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '保存合伙人等级失败');
    },
  });

  const updateLevelMutation = useMutation({
    mutationFn: async (values: PartnerLevelFormValues) => {
      if (!editingLevel?.id) throw new Error('未选择等级');
      const payload: Partial<PartnerLevelPayload> = values;
      return updatePartnerLevel(editingLevel.id, payload);
    },
    onSuccess: () => {
      message.success('合伙人等级已更新');
      setLevelModalOpen(false);
      setEditingLevel(null);
      levelForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['partner-levels'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '更新合伙人等级失败');
    },
  });

  const deleteLevelMutation = useMutation({
    mutationFn: async (level: PartnerLevel) => deletePartnerLevel(level.id),
    onSuccess: () => {
      message.success('已删除合伙人等级');
      queryClient.invalidateQueries({ queryKey: ['partner-levels'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '删除合伙人等级失败');
    },
  });

  const levelColumns: ColumnsType<PartnerLevel> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      sorter: (a, b) => a.id - b.id,
      defaultSortOrder: 'descend',
    },
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '拿货折扣率', dataIndex: 'purchase_discount_rate', width: 140 },
    { title: '直推佣金率', dataIndex: 'direct_commission_rate', width: 140 },
    { title: '团队佣金率', dataIndex: 'team_commission_rate', width: 140 },
    { title: '升级奖励率', dataIndex: 'upgrade_reward_rate', width: 140 },
    { title: '备注', dataIndex: 'note', width: 220 },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setEditingLevel(record);
              levelForm.setFieldsValue({
                name: record.name,
                purchase_discount_rate: record.purchase_discount_rate ? Number(record.purchase_discount_rate) : undefined,
                direct_commission_rate: record.direct_commission_rate ? Number(record.direct_commission_rate) : undefined,
                team_commission_rate: record.team_commission_rate ? Number(record.team_commission_rate) : undefined,
                upgrade_reward_rate: record.upgrade_reward_rate ? Number(record.upgrade_reward_rate) : undefined,
                note: record.note,
              });
              setLevelModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            danger
            loading={deleteLevelMutation.isPending}
            onClick={() => deleteLevelMutation.mutate(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>会员与合伙人配置</Typography.Title>

      {/* 会员套餐配置 */}
      <Card title="会员套餐配置">
        <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
          <Space>
            <Button
              type="primary"
              onClick={() => {
                setEditingPkg(null);
                pkgForm.resetFields();
                pkgForm.setFieldsValue({ type: 'membership' });
                setPkgModalOpen(true);
              }}
            >
              新建会员套餐
            </Button>
          </Space>
          <Space>
            <Typography.Text>按类型筛选：</Typography.Text>
            <Input
              allowClear
              placeholder="例如 membership / partner"
              style={{ width: 220 }}
              value={pkgTypeFilter}
              onChange={(e) => {
                const value = e.target.value?.trim() || undefined;
                setPkgTypeFilter(value);
                setPkgPagination((prev) => ({ ...prev, page: 1 }));
              }}
            />
          </Space>
        </Space>
        {packagesQuery.isLoading && <Spin />}
        {packagesQuery.isError && <Alert type="error" message="无法获取会员套餐列表" showIcon />}
        {packagesQuery.data && (
          <Table<MembershipPackage>
            size="small"
            rowKey="id"
            dataSource={packagesQuery.data.list}
            columns={pkgColumns}
            pagination={{
              current: packagesQuery.data.page ?? pkgPagination.page,
              pageSize: packagesQuery.data.limit ?? pkgPagination.limit,
              total: packagesQuery.data.total,
              showSizeChanger: true,
              onChange: (page, pageSize) =>
                setPkgPagination({ page, limit: pageSize || pkgPagination.limit }),
            }}
          />
        )}
      </Card>

      {/* 合伙人等级配置 */}
      <Card title="合伙人等级配置">
        <Space style={{ marginBottom: 16 }}>
          <Button
            type="primary"
            onClick={() => {
              setEditingLevel(null);
              levelForm.resetFields();
              setLevelModalOpen(true);
            }}
          >
            新建合伙人等级
          </Button>
        </Space>
        {levelsQuery.isLoading && <Spin />}
        {levelsQuery.isError && <Alert type="error" message="无法获取合伙人等级列表" showIcon />}
        {levelsQuery.data && (
          <Table<PartnerLevel>
            size="small"
            rowKey="id"
            dataSource={levelsQuery.data.list}
            columns={levelColumns}
            pagination={{
              current: levelsQuery.data.page ?? levelPagination.page,
              pageSize: levelsQuery.data.limit ?? levelPagination.limit,
              total: levelsQuery.data.total,
              showSizeChanger: true,
              onChange: (page, pageSize) =>
                setLevelPagination({ page, limit: pageSize || levelPagination.limit }),
            }}
          />
        )}
      </Card>

      {/* 会员套餐弹窗 */}
      <Modal
        title={editingPkg ? '编辑会员套餐' : '新建会员套餐'}
        open={pkgModalOpen}
        onCancel={() => {
          if (createPkgMutation.isPending || updatePkgMutation.isPending) return;
          setPkgModalOpen(false);
          setEditingPkg(null);
        }}
        onOk={() => pkgForm.submit()}
        confirmLoading={createPkgMutation.isPending || updatePkgMutation.isPending}
        destroyOnClose
      >
        <Form<MembershipFormValues>
          layout="vertical"
          form={pkgForm}
          initialValues={{ type: 'membership' }}
          onFinish={(values) => {
            if (editingPkg) updatePkgMutation.mutate(values);
            else createPkgMutation.mutate(values);
          }}
        >
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入套餐名称' }]}>
            <Input maxLength={128} placeholder="例如：畅饮VIP" />
          </Form.Item>
          <Form.Item label="价格" name="price" rules={[{ required: true, message: '请输入价格' }]}>
            <InputNumber min={0} style={{ width: '100%' }} placeholder="单位：元" />
          </Form.Item>
          <Form.Item label="赠送茶币" name="tea_coin_award">
            <InputNumber min={0} style={{ width: '100%' }} placeholder="开通时赠送的茶币数量" />
          </Form.Item>
          <Form.Item label="会员折扣率" name="discount_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="例如 0.95 代表 95 折" />
          </Form.Item>
          <Form.Item label="拿货折扣率" name="purchase_discount_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="用于进货价格折扣" />
          </Form.Item>
          <Form.Item label="直推佣金率" name="direct_commission_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="例如 0.1 代表 10%" />
          </Form.Item>
          <Form.Item label="团队佣金率" name="team_commission_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="团队业绩提成比例" />
          </Form.Item>
          <Form.Item label="升级奖励率" name="upgrade_reward_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="用于升级礼包奖励" />
          </Form.Item>
          <Form.Item label="类型标识" name="type">
            <Input maxLength={64} placeholder="默认 membership，可区分不同套餐类型" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 合伙人等级弹窗 */}
      <Modal
        title={editingLevel ? '编辑合伙人等级' : '新建合伙人等级'}
        open={levelModalOpen}
        onCancel={() => {
          if (createLevelMutation.isPending || updateLevelMutation.isPending) return;
          setLevelModalOpen(false);
          setEditingLevel(null);
        }}
        onOk={() => levelForm.submit()}
        confirmLoading={createLevelMutation.isPending || updateLevelMutation.isPending}
        destroyOnClose
      >
        <Form<PartnerLevelFormValues>
          layout="vertical"
          form={levelForm}
          onFinish={(values) => {
            if (editingLevel) updateLevelMutation.mutate(values);
            else createLevelMutation.mutate(values);
          }}
        >
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入等级名称' }]}>
            <Input maxLength={128} placeholder="例如：初级合伙人" />
          </Form.Item>
          <Form.Item label="拿货折扣率" name="purchase_discount_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="用于进货价格折扣" />
          </Form.Item>
          <Form.Item label="直推佣金率" name="direct_commission_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="例如 0.06 代表 6%" />
          </Form.Item>
          <Form.Item label="团队佣金率" name="team_commission_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="团队业绩提成比例" />
          </Form.Item>
          <Form.Item label="升级奖励率" name="upgrade_reward_rate">
            <InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} placeholder="用于升级礼包奖励" />
          </Form.Item>
          <Form.Item label="备注" name="note">
            <Input.TextArea rows={3} maxLength={255} placeholder="可简要描述该等级权益" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
