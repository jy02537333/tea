import React, { useEffect, useState } from 'react';
import { Card, Descriptions, Spin } from 'antd';
import { getOrder } from '../../services/orders';
import { OrderSummary } from '../../services/types';

export default function OrderDetail({ id }: { id: number }) {
  const [order, setOrder] = useState<OrderSummary | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  async function fetch() {
    setLoading(true);
    try {
      const res = await getOrder(id);
      setOrder(res as OrderSummary);
    } finally {
      setLoading(false);
    }
  }

  if (loading) return <Spin />;
  if (!order) return null;

  return (
    <Card style={{ padding: 16 }}>
      <Descriptions title={`订单 ${order.order_no || order.id}`} bordered>
        <Descriptions.Item label="订单号">{order.order_no}</Descriptions.Item>
        <Descriptions.Item label="金额">{order.pay_amount}</Descriptions.Item>
        <Descriptions.Item label="状态">{order.status}</Descriptions.Item>
        <Descriptions.Item label="收货信息">{order.address_info}</Descriptions.Item>
      </Descriptions>
    </Card>
  );
}
