import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, message } from 'antd';
import { accrualSummary } from '../../services/accrual';

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
      <Row gutter={16}>
        <Col span={8}><Card loading={loading}><Statistic title="记录数" value={summary.record_count || 0} /></Card></Col>
        <Col span={8}><Card loading={loading}><Statistic title="用户数" value={summary.user_count || 0} /></Card></Col>
        <Col span={8}><Card loading={loading}><Statistic title="总计息" value={summary.total_interest || '0'} /></Card></Col>
      </Row>
    </div>
  );
}
