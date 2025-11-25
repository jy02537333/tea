import React from 'react';
import { Table } from 'antd';

export default function DataTable<T extends { id?: number | string }>(props: any) {
  return <Table rowKey={(r: any) => r.id} {...props} />;
}
