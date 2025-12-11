import { useMemo, useState } from 'react';
import { Card, Col, DatePicker, Radio, Row, Space, Spin, Statistic, Table, Typography, message } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useQuery } from '@tanstack/react-query';
import { getFinanceSummary, type FinanceDailyRow, type FinanceStoreRow, type FinanceSummary } from '../services/finance';

const { RangePicker } = DatePicker;

interface QueryParams {
  start: string;
  end: string;
}

function formatAmount(val: string | number | undefined): string {
  if (val == null) return '-';
  const n = Number(val);
  if (Number.isNaN(n)) return String(val);
  return n.toFixed(2);
}

export default function FinanceSummaryPage() {
  const [range, setRange] = useState<[Dayjs, Dayjs]>(() => {
    const end = dayjs();
    const start = end.subtract(6, 'day');
    return [start, end];
  });
  const [groupBy, setGroupBy] = useState<'day' | 'store'>('day');

  const params: QueryParams = useMemo(
    () => ({
      start: range[0].format('YYYY-MM-DD'),
      end: range[1].format('YYYY-MM-DD'),
    }),
    [range]
  );

  const summaryQuery = useQuery<{ summary: FinanceSummary; rows?: (FinanceDailyRow | FinanceStoreRow)[] }, Error>({
    queryKey: ['financeSummary', params.start, params.end, groupBy],
    queryFn: async () => {
      try {
        return await getFinanceSummary({ start: params.start, end: params.end, group: groupBy });
      } catch (err: any) {
        message.error(err?.message || '获取财务概览失败');
        throw err;
      }
    },
    staleTime: 30_000,
  });

  const summary = summaryQuery.data?.summary;
  const rows = (summaryQuery.data?.rows ?? []) as (FinanceDailyRow | FinanceStoreRow)[];

  return (
    <Space direction="vertical" size={24} style={{ width: '100%' }}>
      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Typography.Title level={4} style={{ margin: 0 }}>
            财务收支概览
          </Typography.Title>
          <Space>
            <Typography.Text>统计区间：</Typography.Text>
            <RangePicker
              allowClear={false}
              value={range}
              onChange={(val) => {
                if (!val || val.length !== 2) return;
                setRange([val[0], val[1]] as [Dayjs, Dayjs]);
              }}
            />
            <Radio.Group
              value={groupBy}
              onChange={(e) => setGroupBy(e.target.value)}
              style={{ marginLeft: 16 }}
            >
              <Radio.Button value="day">按日</Radio.Button>
              <Radio.Button value="store">按门店</Radio.Button>
            </Radio.Group>
          </Space>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}>
          <Card>
            {summaryQuery.isLoading && !summary ? (
              <Spin />
            ) : (
              <Statistic
                title="总收款金额"
                value={summary ? formatAmount(summary.total_payments_amount) : '-'}
                suffix="元"
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            {summaryQuery.isLoading && !summary ? (
              <Spin />
            ) : (
              <Statistic
                title="总退款金额"
                value={summary ? formatAmount(summary.total_refunds_amount) : '-'}
                suffix="元"
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            {summaryQuery.isLoading && !summary ? (
              <Spin />
            ) : (
              <Statistic
                title="净入账金额"
                value={summary ? formatAmount(summary.net_amount) : '-'}
                suffix="元"
              />
            )}
          </Card>
        </Col>
      </Row>

      <Card title={groupBy === 'day' ? '按日收支明细' : '按门店收支明细'}>
        {groupBy === 'day' ? (
          <Table<FinanceDailyRow>
            rowKey={(row) => row.date}
            loading={summaryQuery.isLoading}
            dataSource={rows as FinanceDailyRow[]}
            pagination={false}
            columns={[
              { title: '日期', dataIndex: 'date', width: 140 },
              { title: '收款笔数', dataIndex: 'pay_count', width: 120 },
              {
                title: '收款金额',
                dataIndex: 'pay_amount',
                width: 140,
                render: (val: string) => formatAmount(val),
              },
              { title: '退款笔数', dataIndex: 'refund_count', width: 120 },
              {
                title: '退款金额',
                dataIndex: 'refund_amount',
                width: 140,
                render: (val: string) => formatAmount(val),
              },
              {
                title: '净入账金额',
                dataIndex: 'net_amount',
                width: 160,
                render: (val: string) => formatAmount(val),
              },
            ]}
          />
        ) : (
          <Table<FinanceStoreRow>
            rowKey={(row) => row.store_id}
            loading={summaryQuery.isLoading}
            dataSource={rows as FinanceStoreRow[]}
            pagination={false}
            columns={[
              { title: '门店ID', dataIndex: 'store_id', width: 100 },
              { title: '门店名称', dataIndex: 'store_name', width: 200 },
              { title: '收款笔数', dataIndex: 'pay_count', width: 120 },
              {
                title: '收款金额',
                dataIndex: 'pay_amount',
                width: 140,
                render: (val: string) => formatAmount(val),
              },
              { title: '退款笔数', dataIndex: 'refund_count', width: 120 },
              {
                title: '退款金额',
                dataIndex: 'refund_amount',
                width: 140,
                render: (val: string) => formatAmount(val),
              },
              {
                title: '净入账金额',
                dataIndex: 'net_amount',
                width: 160,
                render: (val: string) => formatAmount(val),
              },
            ]}
          />
        )}
      </Card>
    </Space>
  );
}
