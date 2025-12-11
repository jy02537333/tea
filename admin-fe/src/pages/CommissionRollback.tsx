import { useState } from 'react';
import { Alert, Button, Card, Form, InputNumber, Space, Typography, message } from 'antd';
import { useMutation } from '@tanstack/react-query';
import { reverseOrderCommission } from '../services/finance';

interface FormValues {
  order_id?: number;
}

export default function CommissionRollbackPage() {
  const [lastResult, setLastResult] = useState<{ order_id: number; processed: number } | null>(null);
  const [form] = Form.useForm<FormValues>();

  const rollbackMutation = useMutation({
    mutationFn: async (values: FormValues) => {
      if (!values.order_id) throw new Error('请输入订单 ID');
      const res = await reverseOrderCommission(values.order_id);
      return res;
    },
    onSuccess: (res) => {
      setLastResult(res);
      if (res.processed > 0) {
        message.success(`已回滚佣金记录 ${res.processed} 条`);
      } else {
        message.info('未找到可回滚的佣金记录');
      }
    },
    onError: (error: any) => {
      message.error(error?.message || '回滚佣金失败');
    },
  });

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Typography.Title level={4}>佣金回滚工具</Typography.Title>

      <Card>
        <Space direction="vertical" size={8} style={{ width: '100%' }}>
          <Typography.Paragraph type="secondary">
            当订单退款已完成但佣金未正确回滚时，财务/运营可以在此根据订单 ID 一键触发后台回滚该订单下尚未提现的佣金记录。
          </Typography.Paragraph>
          <Alert
            type="info"
            showIcon
            message={
              '推荐使用方式：从操作日志中看到 "commission.rollback_failed" 记录后，复制其中的订单 ID 填入下方进行回滚。'
            }
          />
        </Space>

        <Form<FormValues>
          layout="inline"
          form={form}
          onFinish={(values) => rollbackMutation.mutate(values)}
          style={{ rowGap: 16, columnGap: 16, marginTop: 8 }}
        >
          <Form.Item
            label="订单 ID"
            name="order_id"
            rules={[
              { required: true, message: '请输入订单 ID' },
              {
                validator: (_, value) => {
                  if (value == null) return Promise.resolve();
                  if (!Number.isInteger(value) || value <= 0) {
                    return Promise.reject(new Error('订单 ID 必须为大于 0 的整数'));
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <InputNumber placeholder="请输入订单 ID" min={1} precision={0} style={{ width: 220 }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={rollbackMutation.isPending}>
              一键回滚佣金
            </Button>
          </Form.Item>
        </Form>

        <Typography.Paragraph type="secondary" style={{ marginTop: 8 }}>
          提示：仅会影响该订单下仍处于“冻结/可用”状态的佣金，已提现部分不会自动回滚，需要线下人工处理。
        </Typography.Paragraph>

        {lastResult && (
          <div style={{ marginTop: 16 }}>
            <Alert
              type={lastResult.processed > 0 ? 'success' : 'info'}
              showIcon
              message={`订单 ${lastResult.order_id} 回滚结果：处理记录数 ${lastResult.processed}`}
            />
          </div>
        )}
      </Card>
    </Space>
  );
}
