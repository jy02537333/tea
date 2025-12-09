import { useMemo, useState } from 'react';
import { Button, Card, Col, DatePicker, Form, InputNumber, Row, Space, Statistic, Typography, message } from 'antd';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { exportAccrual, getAccrualSummary, runAccrual, type RunAccrualResponse } from '../services/accrual';

const { RangePicker } = DatePicker;
const SingleDatePicker = DatePicker;
const DEFAULT_RANGE_DAYS = 14;

interface RunFormValues {
  date?: Dayjs;
  rate?: number;
}

export default function AccrualPage() {
  const queryClient = useQueryClient();
  const [range, setRange] = useState<[Dayjs, Dayjs]>(() => {
    const end = dayjs();
    const start = end.subtract(DEFAULT_RANGE_DAYS - 1, 'day');
    return [start, end];
  });
  const [runForm] = Form.useForm<RunFormValues>();

  const params = useMemo(
    () => ({
      start: range[0].format('YYYY-MM-DD'),
      end: range[1].format('YYYY-MM-DD'),
    }),
    [range]
  );

  const summaryQuery = useQuery({
    queryKey: ['accrualSummary', params.start, params.end],
    queryFn: () => getAccrualSummary(params),
    staleTime: 30_000,
  });

  const runAccrualMutation = useMutation<RunAccrualResponse, Error, RunFormValues>({
    mutationFn: async (values: RunFormValues) => {
      const payload = {
        date: values.date?.format('YYYY-MM-DD'),
        rate: values.rate,
      };
      return runAccrual(payload);
    },
    onSuccess: async (res) => {
      message.success(`计提执行成功，更新记录数：${res.updated}`);
      runForm.resetFields();
      await queryClient.invalidateQueries({ queryKey: ['accrualSummary'] });
    },
    onError: (error: any) => {
      message.error(error?.message || '计提执行失败');
    },
  });

  const exportMutation = useMutation({
    mutationFn: async () => {
      const blob = await exportAccrual({ start: params.start, end: params.end });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `accrual_${params.start}_${params.end}.xlsx`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    },
    onError: (error: any) => {
      message.error(error?.message || '导出失败');
    },
    onSuccess: () => {
      message.success('导出任务已开始，文件已下载');
    },
  });

  const stats = [
    { title: '记录总数', value: summaryQuery.data?.record_count ?? '-', key: 'record_count' },
    { title: '用户数量', value: summaryQuery.data?.user_count ?? '-', key: 'user_count' },
    { title: '累计利息', value: summaryQuery.data?.total_interest ?? '-', key: 'total_interest' },
    { title: '今日订单数', value: summaryQuery.data?.today_orders ?? '-', key: 'today_orders' },
  ];

  return (
    <Space direction="vertical" size={24} style={{ width: '100%' }}>
      <Card>
        <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
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
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => summaryQuery.refetch()} loading={summaryQuery.isFetching}>
              刷新
            </Button>
            <Button
              icon={<DownloadOutlined />}
              onClick={() => exportMutation.mutate()}
              loading={exportMutation.isPending}
            >
              导出报表
            </Button>
          </Space>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        {stats.map((item) => (
          <Col key={item.key} xs={24} sm={12} md={6}>
            <Card>
              <Statistic title={item.title} value={item.value} valueStyle={{ fontSize: 26 }} />
            </Card>
          </Col>
        ))}
      </Row>

      <Card title="触发计提" extra={<Typography.Text type="secondary">可指定日期/利率，留空则使用默认配置</Typography.Text>}>
        <Form<RunFormValues>
          layout="inline"
          form={runForm}
          onFinish={(values) => runAccrualMutation.mutate(values)}
          style={{ rowGap: 16, columnGap: 16 }}
        >
          <Form.Item label="计提日期" name="date">
            <SingleDatePicker allowClear placeholder="默认今天" />
          </Form.Item>
          <Form.Item label="计提利率" name="rate">
            <InputNumber addonAfter="%" min={0} step={0.01} style={{ width: 180 }} placeholder="默认配置" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={runAccrualMutation.isPending}>
              执行计提
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="操作说明">
        <ol style={{ paddingLeft: 20 }}>
          <li>设置统计区间后点击“刷新”或“导出报表”即可获取对应数据。</li>
          <li>执行计提后，系统会更新指定日期内的计息记录，并刷新上方概览。</li>
          <li>导出结果为 Excel（服务器返回的默认格式），可直接分享给财务或运营团队。</li>
        </ol>
      </Card>
    </Space>
  );
}
