import React, { useEffect, useState } from 'react';
import { Table, Button, Space } from 'antd';
import { listProducts } from '../../services/products';
import { Product } from '../../services/types';

export default function ProductList() {
  const [data, setData] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetch();
  }, []);

  async function fetch() {
    setLoading(true);
    try {
      const res = await listProducts({ page: 1, limit: 20 });
      // normalize several possible response shapes: Paginated { data: [] }, { items: [] } or raw array
      const maybe = res as any;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setData(items);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ padding: 16 }}>
      <Space style={{ marginBottom: 12 }}>
        <Button type="primary">新建商品</Button>
      </Space>
      <Table<Product>
        rowKey="id"
        loading={loading}
        dataSource={data}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '名称', dataIndex: 'name' },
          { title: '价格', dataIndex: 'price' },
          { title: '库存', dataIndex: 'stock' },
        ]}
      />
    </div>
  );
}
