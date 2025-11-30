import React, { useEffect, useState } from 'react';
import { Table, Button, Space, message, Modal } from 'antd';
import Thumbnail from '../../components/Thumbnail';
import { listOssFiles, deleteOssFiles } from '../../services/oss-manager';

const PAGE_SIZE = 20;

export default function OssManager() {
  const [data, setData] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [marker, setMarker] = useState<string>('');
  const [nextMarker, setNextMarker] = useState<string>('');

  useEffect(() => {
    fetchList();
    // eslint-disable-next-line
  }, [marker]);

  async function fetchList() {
    setLoading(true);
    try {
      const res = await listOssFiles({ marker, limit: PAGE_SIZE });
      setData(res.files || []);
      setNextMarker(res.next_marker || '');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete() {
    if (!selected.length) return message.warning('请先选择图片');
    Modal.confirm({
      title: '确认删除选中图片？',
      onOk: async () => {
        try {
          await deleteOssFiles(selected);
          message.success('删除成功');
          setSelected([]);
          fetchList();
        } catch {
          message.error('删除失败');
        }
      },
    });
  }

  return (
    <div style={{ padding: 16 }}>
      <h1>OSS图片管理</h1>
      <Space style={{ marginBottom: 16 }}>
        <Button danger disabled={!selected.length} onClick={handleDelete}>批量删除</Button>
        <Button onClick={() => setMarker(nextMarker)} disabled={!nextMarker}>下一页</Button>
      </Space>
      <Table
        rowKey={(url) => url}
        loading={loading}
        dataSource={data}
        rowSelection={{
          selectedRowKeys: selected,
          onChange: (keys) => setSelected((keys as React.Key[]).map(String)),
        }}
        columns={[{
          title: '图片',
          dataIndex: '',
          render: (url: string) => <Thumbnail src={url} width={80} height={80} />,
        }, {
          title: 'URL',
          dataIndex: '',
          render: (url: string) => <span style={{ wordBreak: 'break-all' }}>{url}</span>,
        }]}
        pagination={false}
      />
    </div>
  );
}
