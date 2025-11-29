import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, message, Space, Button, DatePicker } from 'antd';
import { accrualSummary, runAccrual, accrualExport } from '../../services/accrual';

export default function Dashboard() {
  const [loading, setLoading] = useState(false);
  const [summary, setSummary] = useState<any>({ record_count: 0, user_count: 0, total_interest: '0' });

  useEffect(() => { fetchSummary(); }, []);

  async function fetchSummary() {
    setLoading(true);
    try {
      const res = await accrualSummary();
      setSummary(res || {});
    } catch (e: any) {
      message.error(e?.message || '加载汇总失败');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ padding: 16 }}>
      <Space style={{ marginBottom: 12 }}>
        <DatePicker.RangePicker onChange={async (v) => {
          try {
            const start = v?.[0]?.format('YYYY-MM-DD');
            const end = v?.[1]?.format('YYYY-MM-DD');
            setLoading(true);
            const res = await accrualSummary({ start, end });
            setSummary(res || {});
          } finally { setLoading(false); }
        }} />
        <Button onClick={async () => {
          try {
            setLoading(true);
            const today = new Date().toISOString().slice(0,10);
            const res = await runAccrual({ date: today });
            message.success(`已触发计提，更新 ${res.updated} 条`);
            fetchSummary();
          } catch (e: any) {
            message.error(e?.message || '触发计提失败');
          } finally { setLoading(false); }
        }}>
          触发计提
        </Button>
        <Button onClick={async () => {
          try {
            const blob = await accrualExport({ format: 'xlsx' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url; a.download = `accrual_${Date.now()}.xlsx`; a.click();
            URL.revokeObjectURL(url);
          } catch (e: any) {
            message.error(e?.message || '导出失败');
          }
        }}>
          导出报表
        </Button>
      </Space>
      <Row gutter={16}>
        <Col span={8}><Card loading={loading}><Statistic title="记录数" value={summary.record_count || 0} /></Card></Col>
        <Col span={8}><Card loading={loading}><Statistic title="用户数" value={summary.user_count || 0} /></Card></Col>
        <Col span={8}><Card loading={loading}><Statistic title="总计息" value={summary.total_interest || '0'} /></Card></Col>
      </Row>
    </div>
  );
}
