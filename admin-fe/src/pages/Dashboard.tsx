import { useMemo, useState } from 'react';
import { Button, Card, Col, DatePicker, Row, Space, Statistic, message, Select } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getAccrualSummary, runAccrual } from '../services/accrual';
import { getDashboardTodos, getOrderTrends, getDashboardSummary, type OrderTrendPoint } from '../services/dashboard';
import { useNavigate } from 'react-router-dom';
import type { RunAccrualResponse } from '../services/accrual';
import TrendChart from '../components/TrendChart';

const { RangePicker } = DatePicker;
const DEFAULT_RANGE_DAYS = 7;

export default function DashboardPage() {
  const navigate = useNavigate();
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

  const todosQuery = useQuery({
    queryKey: ['dashboardTodos'],
    queryFn: () => getDashboardTodos(),
    staleTime: 30_000,
  });

  const [trendsDays, setTrendsDays] = useState<number>(DEFAULT_RANGE_DAYS);
  const trendsQuery = useQuery<OrderTrendPoint[]>({
    queryKey: ['orderTrends', trendsDays],
    queryFn: () => getOrderTrends(trendsDays),
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

  // 今日概览（后台 summary）
  const dashSummaryQuery = useQuery({
    queryKey: ['dashboardSummary'],
    queryFn: () => getDashboardSummary(),
    staleTime: 30_000,
  });

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

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={12}>
          <Card loading={dashSummaryQuery.isLoading} title="今日概览">
            <Space size={24} wrap>
              <Statistic title="今日订单数" value={dashSummaryQuery.data?.today_order_count ?? 0} />
              <Statistic title="今日已支付" value={dashSummaryQuery.data?.today_paid_order_count ?? 0} />
              <Statistic title="今日销售额 (￥)" value={(dashSummaryQuery.data?.today_sales_amount ?? 0).toFixed(2)} />
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={12}>
          <Card loading={dashSummaryQuery.isLoading} title="昨日 / 近7日 / 近30日">
            <Space size={24} wrap>
              <Statistic title="昨日已支付" value={dashSummaryQuery.data?.yesterday_paid_order_count ?? 0} />
              <Statistic title="昨日销售额 (￥)" value={(dashSummaryQuery.data?.yesterday_sales_amount ?? 0).toFixed(2)} />
              <Statistic title="近7日已支付" value={dashSummaryQuery.data?.last7d_paid_order_count ?? 0} />
              <Statistic title="近7日销售额 (￥)" value={(dashSummaryQuery.data?.last7d_sales_amount ?? 0).toFixed(2)} />
              <Statistic title="近30日已支付" value={dashSummaryQuery.data?.last30d_paid_order_count ?? 0} />
              <Statistic title="近30日销售额 (￥)" value={(dashSummaryQuery.data?.last30d_sales_amount ?? 0).toFixed(2)} />
            </Space>
          </Card>
        </Col>
      </Row>

      <Card
        title={`订单趋势（近${trendsDays}天）`}
        extra={
          <Space size={8}>
            <Select
              size="small"
              style={{ width: 96 }}
              value={trendsDays}
              options={[
                { label: '近7天', value: 7 },
                { label: '近14天', value: 14 },
                { label: '近30天', value: 30 },
              ]}
              onChange={(v) => setTrendsDays(v)}
            />
            <Button size="small" onClick={() => trendsQuery.refetch()} loading={trendsQuery.isFetching}>刷新</Button>
          </Space>
        }
      >
        {trendsQuery.isLoading ? (
          <div style={{ height: 260, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#bbb' }}>加载中...</div>
        ) : trendsQuery.isError ? (
          <div style={{ height: 260, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#ff4d4f' }}>趋势数据加载失败</div>
        ) : (
          <TrendChart data={trendsQuery.data ?? []} />
        )}
      </Card>

      <Card title="待办事项">
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Space style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span>待处理客服工单</span>
            <Space>
              <strong>{todosQuery.data?.ticket_pending_count ?? 0}</strong>
              <Button type="link" size="small" onClick={() => navigate('/tickets')}>
                前往处理
              </Button>
            </Space>
          </Space>
          <Space style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span>待发货订单</span>
            <Space>
              <strong>{todosQuery.data?.order_to_ship_count ?? 0}</strong>
              <Button type="link" size="small" onClick={() => navigate('/orders?status=paid')}>
                查看订单
              </Button>
            </Space>
          </Space>
          <Space style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span>待处理提现</span>
            <Space>
              <strong>{todosQuery.data?.withdraw_pending_count ?? 0}</strong>
              <Button type="link" size="small" onClick={() => navigate('/finance/withdraws?status=pending')}>
                查看提现
              </Button>
            </Space>
          </Space>
        </Space>
      </Card>
    </Space>
  );
}
