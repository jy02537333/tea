import React, { useState } from 'react';
import { Table, Button, Space, Modal, Input, message } from 'antd';
import { ColumnsType } from 'antd/es/table';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listAdminWithdrawals, approveWithdrawal, rejectWithdrawal, AdminWithdrawalRow } from '../services/withdrawal';

export default function PartnerWithdrawalsPage() {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [filterUser, setFilterUser] = useState('');
  const [selected, setSelected] = useState<AdminWithdrawalRow | null>(null);
  const [remark, setRemark] = useState('');
  const qc = useQueryClient();

  const { data, isLoading } = useQuery<{ list: AdminWithdrawalRow[]; page: number; limit: number; total: number }>({
    queryKey: ['admin-withdrawals', page, limit, filterUser],
    queryFn: () => listAdminWithdrawals({ page, limit, user_id: filterUser || undefined }),
  });

  const approveMut = useMutation({
    mutationFn: ({ id, remark }: { id: number; remark?: string }) => approveWithdrawal(id, { remark }),
    onSuccess: () => { message.success('已受理'); qc.invalidateQueries({ queryKey: ['admin-withdrawals'] }); setRemark(''); setSelected(null); },
    onError: (e: any) => { message.error(e?.message || '操作失败'); },
  });
  const rejectMut = useMutation({
    mutationFn: ({ id, remark }: { id: number; remark?: string }) => rejectWithdrawal(id, { remark }),
    onSuccess: () => { message.success('已拒绝并解冻'); qc.invalidateQueries({ queryKey: ['admin-withdrawals'] }); setRemark(''); setSelected(null); },
    onError: (e: any) => { message.error(e?.message || '操作失败'); },
  });

  const columns: ColumnsType<AdminWithdrawalRow> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '提现单号', dataIndex: 'withdraw_no' },
    { title: '用户ID', dataIndex: 'user_id', width: 100 },
    { title: '金额', dataIndex: 'amount', width: 120 },
    { title: '手续费', dataIndex: 'fee', width: 120 },
    { title: '实付', dataIndex: 'actual_amount', width: 120 },
    { title: '状态', dataIndex: 'status', width: 120 },
    { title: '申请时间', dataIndex: 'requested_at', width: 180 },
    {
      title: '操作', width: 220, render: (_, record) => (
        <Space>
          <Button size="small" onClick={() => { setSelected(record); setRemark(''); }}>审核</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <h3>合伙人提现审核</h3>
      <Space style={{ marginBottom: 12 }}>
        <Input placeholder="按用户ID过滤" value={filterUser} onChange={(e) => setFilterUser(e.target.value)} style={{ width: 200 }} />
        <Button onClick={() => { setPage(1); qc.invalidateQueries({ queryKey: ['admin-withdrawals'] }); }}>查询</Button>
      </Space>
      <Table rowKey="id" columns={columns} dataSource={data?.list} loading={isLoading}
        pagination={{ current: data?.page || page, pageSize: data?.limit || limit, total: data?.total || 0, onChange: (p, ps) => { setPage(p); setLimit(ps || limit); } }} />

      <Modal title={selected ? `审核提现 ${selected.withdraw_no}` : '审核'} visible={selected !== null} onCancel={() => setSelected(null)} footer={null}>
        {selected && (
          <div>
            <p>用户ID：{selected.user_id}</p>
            <p>金额：{selected.amount}（手续费 {selected.fee}，实付 {selected.actual_amount}）</p>
            <Input.TextArea rows={4} placeholder="备注（选填）" value={remark} onChange={(e) => setRemark(e.target.value)} style={{ marginBottom: 12 }} />
            <Space>
              <Button type="primary" onClick={() => approveMut.mutate({ id: selected.id, remark })} loading={approveMut.isPending}>受理</Button>
              <Button danger onClick={() => rejectMut.mutate({ id: selected.id, remark })} loading={rejectMut.isPending}>拒绝并解冻</Button>
              <Button onClick={() => setSelected(null)}>取消</Button>
            </Space>
          </div>
        )}
      </Modal>
    </div>
  );
}
