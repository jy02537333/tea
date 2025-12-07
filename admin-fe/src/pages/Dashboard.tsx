import { useMemo, useState } from 'react';
import { Button, Card, Col, DatePicker, Row, Space, Statistic, message } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getAccrualSummary, runAccrual } from '../services/accrual';
import type { RunAccrualResponse } from '../services/accrual';

const { RangePicker } = DatePicker;
const DEFAULT_RANGE_DAYS = 7;

export default function DashboardPage() {
  const queryClient = useQueryClient();
  const [range, setRange] = useState<[Dayjs, Dayjs]>(() => {
    const end = dayjs();
    const start = end.subtract(DEFAULT_RANGE_DAYS - 1, 'day');
    return [start, end];
  });

  const params = useMemo(() => ({
    start: range[0].format('YYYY-MM-DD'),
    end: range[1].format('YYYY-MM-DD'),
  }), [range]);

  const summaryQuery = useQuery({
    queryKey: ['accrualSummary', params.start, params.end],
    queryFn: () => getAccrualSummary(params),
    staleTime: 30_000,
  });

  const runAccrualMutation = useMutation<RunAccrualResponse, Error>({
    mutationFn: () => runAccrual({}),
    onSuccess: async (res: RunAccrualResponse) => {
      message.success(`计提执行成功，更新记录数：${res.updated}`);
      await queryClient.invalidateQueries({ queryKey: ['accrualSummary'] });
    },
    onError: (err: any) => {
      message.error(err?.message || '触发计提失败');
    },
  });

  const cards = [
    { key: 'record_count', title: '记录总数', value: summaryQuery.data?.record_count ?? '-' },
    { key: 'user_count', title: '用户数量', value: summaryQuery.data?.user_count ?? '-' },
    { key: 'total_interest', title: '累计利息', value: summaryQuery.data?.total_interest ?? '-' },
    { key: 'today_orders', title: '今日订单数', value: summaryQuery.data?.today_orders ?? '-' },
  ];

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={24}>
      <Card>
        <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <span>日期范围：</span>
            <RangePicker
              allowClear={false}
              value={range}
              onChange={(val) => {
                if (!val || val.length !== 2) return;
                setRange([val[0], val[1]] as [Dayjs, Dayjs]);
              }}
            />
          </Space>
          <Space>
            <Button onClick={() => summaryQuery.refetch()} loading={summaryQuery.isFetching}>
              刷新
            </Button>
            <Button type="primary" loading={runAccrualMutation.isPending} onClick={() => runAccrualMutation.mutate()}>
              触发计提
            </Button>
          </Space>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        {cards.map((card) => (
          <Col xs={24} sm={12} md={6} key={card.key}>
            <Card>
              <Statistic title={card.title} value={card.value} valueStyle={{ fontSize: 28 }} />
            </Card>
          </Col>
        ))}
      </Row>

      <Card title="趋势图（占位）">
        <div style={{ height: 260, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#bbb', border: '1px dashed #e0e0e0' }}>
          图表区域（后续接入 ECharts / AntV）
        </div>
      </Card>

      <Card title="最近任务 / 操作（占位）">
        <p>这里可以展示最近计提任务、导出记录或操作日志。</p>
      </Card>
    </Space>
  );
}
