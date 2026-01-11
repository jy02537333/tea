import { useMemo, useState } from 'react';
import { Alert, Button, Card, Segmented, Space, Statistic, Tag, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getStoreOrderStatsScoped } from '../services/stores';
import { useAuthContext } from '../hooks/useAuth';

const { Title, Text } = Typography;

const STATUS_META: Record<number, { label: string; color: string }> = {
  1: { label: '待付款', color: 'orange' },
  2: { label: '已付款', color: 'blue' },
  3: { label: '配送中', color: 'purple' },
  4: { label: '已完成', color: 'green' },
  5: { label: '已取消', color: 'red' },
};

export default function StoreHomePage() {
  const navigate = useNavigate();
  const { user } = useAuthContext();
  const lockedStoreId = user?.store_id;

  const [days, setDays] = useState<number>(7);

  const statsQuery = useQuery({
    queryKey: ['store-home-stats', lockedStoreId, days],
    queryFn: () => getStoreOrderStatsScoped(lockedStoreId!, { days }),
    enabled: !!lockedStoreId,
  });

  const statusTags = useMemo(() => {
    const items = statsQuery.data?.status_counts ?? [];
    const sorted = [...items].sort((a, b) => a.status - b.status);
    return sorted.map((it) => {
      const meta = STATUS_META[it.status];
      return (
        <Tag key={it.status} color={meta?.color}>
          {meta?.label ?? `状态${it.status}`}: {it.count}
        </Tag>
      );
    });
  }, [statsQuery.data?.status_counts]);

  if (user?.role !== 'store') {
    return (
      <Alert
        type="warning"
        showIcon
        message="该页面仅用于门店后台"
        description="当前账号不是门店管理员（role=store）。"
      />
    );
  }

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          门店首页
          {lockedStoreId ? (
            <Text type="secondary" style={{ marginLeft: 8 }}>
              (门店ID: {lockedStoreId})
            </Text>
          ) : null}
        </Title>

        <Space wrap>
          <Text type="secondary">时间维度：</Text>
          <Segmented
            value={days}
            options={[
              { label: '近7天', value: 7 },
              { label: '近30天', value: 30 },
            ]}
            onChange={(val) => setDays(Number(val))}
          />
        </Space>
      </Space>

      {!lockedStoreId && <Alert type="error" showIcon message="门店管理员未绑定门店（store_admins），无法查看经营数据" />}

      {lockedStoreId && (
        <Card title="经营参谋" loading={statsQuery.isLoading}>
          {statsQuery.isError && <Alert type="error" showIcon message="无法获取门店订单统计" />}
          {statsQuery.data && (
            <Space direction="vertical" size={12} style={{ width: '100%' }}>
              <Space size={24} wrap>
                <Statistic title="订单数" value={statsQuery.data.total_orders} />
                <Statistic title="成交额（已完成）" value={`￥${Number(statsQuery.data.completed_amount ?? 0).toFixed(2)}`} />
              </Space>
              <div>{statusTags}</div>
            </Space>
          )}
        </Card>
      )}

      {lockedStoreId && (
        <Card title="工作台">
          <Space wrap>
            <Button type="primary" onClick={() => navigate('/store-orders')}>
              待处理订单/订单列表
            </Button>
            <Button onClick={() => navigate('/store-products')}>商品管理</Button>
            <Button onClick={() => navigate('/store-finance')}>财务管理</Button>
            <Button onClick={() => navigate('/store-settings')}>门店设置</Button>
          </Space>
        </Card>
      )}
    </Space>
  );
}
