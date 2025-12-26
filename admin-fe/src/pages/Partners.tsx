import { Button, Modal, Table, Input, Space } from 'antd';
import { ColumnsType } from 'antd/es/table';
import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listPartners, listPartnerCommissions, PartnerRow, CommissionRow } from '../services/partner';

export default function PartnersPage() {
  const [q, setQ] = useState<string>('');
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [selectedPartner, setSelectedPartner] = useState<PartnerRow | null>(null);

  const { data, isLoading } = useQuery<{ list: any[]; page: number; limit: number; total: number }>({
    queryKey: ['partners', page, limit, q],
    queryFn: () => listPartners({ page, limit, q }),
  });

  const columns: ColumnsType<PartnerRow> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '昵称', dataIndex: 'nickname' },
    { title: '手机号', dataIndex: 'phone' },
    { title: '等级ID', dataIndex: 'partner_level_id', width: 120 },
    { title: '加入时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', dataIndex: 'op', width: 180, render: (_, record) => (
        <Space>
          <Button size="small" onClick={() => setSelectedPartner(record)}>查看佣金</Button>
        </Space>
      )
    }
  ];

  return (
    <div>
      <h3>合伙人管理</h3>
      <Space style={{ marginBottom: 12 }}>
        <Input.Search placeholder="搜索昵称或手机号" allowClear onSearch={(v) => { setQ(v); setPage(1); }} style={{ width: 320 }} />
      </Space>
      <Table rowKey="id" columns={columns} dataSource={data?.list} loading={isLoading}
        pagination={{ current: data?.page || page, pageSize: data?.limit || limit, total: data?.total || 0, onChange: (p, ps) => { setPage(p); setLimit(ps || limit); } }} />

      <PartnerCommissionsModal partner={selectedPartner} onClose={() => setSelectedPartner(null)} />
    </div>
  );
}

function PartnerCommissionsModal({ partner, onClose }: { partner: PartnerRow | null; onClose: () => void }) {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const id = partner?.id ?? 0;
  const { data: cmData, isLoading: cmLoading } = useQuery<{ list: any[]; page: number; limit: number; total: number }>({
    queryKey: ['partner-commissions', id, page, limit],
    queryFn: () => listPartnerCommissions(id, { page, limit }),
    enabled: id !== 0,
  });

  const cols: ColumnsType<CommissionRow> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '类型', dataIndex: 'commission_type' },
    { title: '毛额', dataIndex: 'gross_amount' },
    { title: '净额', dataIndex: 'net_amount' },
    { title: '状态', dataIndex: 'status' },
    { title: '时间', dataIndex: 'created_at' },
  ];

  return (
    <Modal title={partner ? `佣金 - ${partner.nickname}` : '佣金'} visible={partner !== null} footer={null} onCancel={onClose} width={800}>
      <Table rowKey="id" columns={cols} dataSource={cmData?.list} loading={cmLoading}
        pagination={{ current: cmData?.page || page, pageSize: cmData?.limit || limit, total: cmData?.total || 0, onChange: (p, ps) => { setPage(p); setLimit(ps || limit); } }} />
    </Modal>
  );
}
