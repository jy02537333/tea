import React, { useEffect, useRef, useState } from 'react';
import Thumbnail from './Thumbnail';
import { Modal, List, Button, message, Upload, InputNumber, Space, Tooltip, Checkbox, Progress } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { getOssSignature } from '../services/oss';
import { getProductImages, addProductImage, updateProductImage, deleteProductImage, batchDeleteProductImages } from '../services/products';

interface Props {
  productId: number | null;
  visible: boolean;
  onClose: () => void;
  onUpdated?: () => void;
}

const ProductImagesManager: React.FC<Props> = ({ productId, visible, onClose, onUpdated }) => {
  const [images, setImages] = useState<any[]>([]);
  const [selectedIds, setSelectedIds] = useState<Record<number, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const dragIndexRef = useRef<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  useEffect(() => {
    if (visible && productId) fetchImages();
    // eslint-disable-next-line
  }, [visible, productId]);

  async function fetchImages() {
    if (!productId) return;
    setLoading(true);
    try {
      const res = await getProductImages(productId);
      setImages(res || []);
    } catch (e: any) {
      message.error('获取图片失败');
    } finally {
      setLoading(false);
    }
  }

  // custom upload to OSS, then create ProductImage record
  async function customOssUpload({ file, onSuccess, onError, onProgress }: any) {
    try {
      const sign = await getOssSignature({ dir: 'products/' });
      const formData = new FormData();
      const filename = `${sign.dir}${Date.now()}_${file.name}`;
      formData.append('key', filename);
      formData.append('OSSAccessKeyId', sign.accessid);
      formData.append('policy', sign.policy);
      formData.append('Signature', sign.signature);
      formData.append('success_action_status', '200');
      formData.append('file', file);
      // create a temporary optimistic item to show upload progress
      const tempId = Date.now() * -1;
      setImages(prev => [{ id: tempId, image_url: '', progress: 0, isTemp: true }, ...prev]);

      const xhr = new XMLHttpRequest();
      xhr.open('POST', sign.host, true);
      xhr.upload.onprogress = (evt) => {
        if (evt.lengthComputable && onProgress) {
          const percent = (evt.loaded / evt.total) * 100;
          onProgress({ percent });
          // update temp progress
          setImages(prev => prev.map(it => it.id === tempId ? { ...it, progress: percent } : it));
        }
      };
      xhr.onload = async () => {
        if (xhr.status === 200) {
          const url = sign.host + '/' + filename;
          // create ProductImage record
          try {
            await addProductImage(productId!, { image_url: url, sort: 0, is_main: false });
            onSuccess && onSuccess({ url }, file);
            // refresh list to replace temp item
            await fetchImages();
            onUpdated && onUpdated();
          } catch (e) {
            onError && onError(e);
          }
        } else {
          onError && onError(new Error('上传失败'));
        }
      };
      xhr.onerror = () => onError && onError(new Error('上传失败'));
      xhr.send(formData);
    } catch (e) {
      onError && onError(e);
    }
  }

  async function handleDelete(id: number) {
    Modal.confirm({
      title: '确认删除该图片？',
      onOk: async () => {
        try {
          // optimistic remove
          const old = images.slice();
          setImages(images.filter(i => i.id !== id));
          await deleteProductImage(productId!, id);
          message.success('删除成功');
          onUpdated && onUpdated();
        } catch (e: any) {
          message.error(e?.response?.data?.message || '删除失败');
          fetchImages();
        }
      }
    });
  }

  async function handleBatchDelete() {
    const ids = Object.keys(selectedIds).filter(k => selectedIds[Number(k)]).map(k => Number(k));
    if (ids.length === 0) return message.info('请先选择要删除的图片');
    Modal.confirm({
      title: `确认删除 ${ids.length} 张图片？`,
      onOk: async () => {
        const old = images.slice();
        try {
          // optimistic remove
          setImages(images.filter(i => !ids.includes(i.id)));
          // call backend batch delete
          await batchDeleteProductImages(productId!, ids);
          message.success('批量删除完成');
          setSelectedIds({});
          onUpdated && onUpdated();
        } catch (e) {
          message.error('批量删除出错，已刷新');
          fetchImages();
        }
      }
    });
  }

  // Drag & drop handlers for reordering
  function onDragStart(e: React.DragEvent, index: number) {
    dragIndexRef.current = index;
    e.dataTransfer.effectAllowed = 'move';
    // nicer cursor during drag
    try { document.body.style.cursor = 'grabbing'; } catch {}
  }

  function onDragOver(e: React.DragEvent) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    // keep showing placeholder as user drags
  }

  async function onDrop(e: React.DragEvent, dropIndex: number) {
    e.preventDefault();
    const from = dragIndexRef.current;
    if (from === null || from === undefined) return;
    if (from === dropIndex) return;
    const newList = images.slice();
    const [moved] = newList.splice(from, 1);
    newList.splice(dropIndex, 0, moved);
    // assign new sort values based on index
    const updates = newList.map((it, idx) => ({ id: it.id, sort: idx }));
    // optimistic reorder
    setImages(newList.map((it, idx) => ({ ...it, sort: idx })));
    try {
      await Promise.all(updates.map(u => updateProductImage(productId!, u.id, { sort: u.sort }).catch(err => ({ err, id: u.id }))));
      message.success('排序已保存');
      onUpdated && onUpdated();
    } catch (e) {
      message.error('保存排序失败，已刷新');
      fetchImages();
    } finally {
      dragIndexRef.current = null;
      setDragOverIndex(null);
      try { document.body.style.cursor = ''; } catch {}
    }
  }

  function onDragEnter(e: React.DragEvent, index: number) {
    e.preventDefault();
    setDragOverIndex(index);
  }

  function onDragLeave(_e: React.DragEvent) {
    setDragOverIndex(null);
  }

  async function handleSetMain(id: number) {
    try {
      // optimistic update
      const old = images.slice();
      setImages(images.map(i => ({ ...i, is_main: i.id === id })));
      await updateProductImage(productId!, id, { is_main: true });
      message.success('设为主图成功');
      onUpdated && onUpdated();
    } catch (e: any) {
      message.error(e?.response?.data?.message || '操作失败');
      fetchImages();
    }
  }

  return (
    <Modal open={visible} title="商品图片管理" footer={null} onCancel={onClose} width={800}>
      <Space style={{ marginBottom: 12 }}>
        <Upload
          customRequest={customOssUpload}
          showUploadList={false}
          accept="image/*"
        >
          <Button icon={<UploadOutlined />}>上传图片</Button>
        </Upload>
        <Button danger onClick={handleBatchDelete}>批量删除</Button>
        <Button onClick={fetchImages}>刷新</Button>
      </Space>
      <List
        loading={loading}
        dataSource={images}
        renderItem={(item) => {
          const idx = images.findIndex(i => i.id === item.id);
          const isDragOver = dragOverIndex === idx;
          return (
            <List.Item
              key={item.id}
              draggable={!item.isTemp}
              onDragStart={(e) => onDragStart(e, idx)}
              onDragEnter={(e) => onDragEnter(e, idx)}
              onDragOver={onDragOver}
              onDragLeave={onDragLeave}
              onDrop={(e) => onDrop(e, idx)}
              actions={[
                <Checkbox key="select" checked={!!selectedIds[item.id]} onChange={(e) => setSelectedIds(prev => ({ ...prev, [item.id]: e.target.checked }))} />,
                <Button key="main" type={item.is_main ? 'primary' : 'link'} onClick={() => handleSetMain(item.id)}>{item.is_main ? '主图' : '设为主图'}</Button>,
                <Button key="del" danger onClick={() => handleDelete(item.id)}>删除</Button>
              ]}
            >
              <div style={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                <List.Item.Meta
                  avatar={
                    <span style={{ display: 'inline-block', width: 160, height: 100, marginRight: 8, cursor: item.isTemp ? 'default' : 'grab', transition: 'transform .12s ease, box-shadow .12s ease', boxShadow: isDragOver ? '0 8px 20px rgba(0,0,0,0.12)' : undefined, transform: isDragOver ? 'scale(1.02)' : undefined, border: isDragOver ? '2px dashed #1890ff' : undefined }}>
                      <Thumbnail src={item.image_url || undefined} width={160} height={100} />
                    </span>
                  }
                  title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <span style={{ wordBreak: 'break-all', maxWidth: 360, overflow: 'hidden', textOverflow: 'ellipsis' }}>{item.image_url}</span>
                      <Tooltip title="排序越小越靠前"><span>排序: </span></Tooltip>
                      <InputNumber size="small" defaultValue={item.sort} onBlur={async (e) => {
                        const v = Number((e.target as HTMLInputElement).value || 0);
                        if (v !== item.sort) {
                          try {
                            await updateProductImage(productId!, item.id, { sort: v });
                            fetchImages();
                            onUpdated && onUpdated();
                          } catch {
                            message.error('更新排序失败');
                          }
                        }
                      }} />
                    </div>
                  }
                  description={item.is_main ? '当前主图' : ''}
                />
                {item.isTemp && <div style={{ width: 160, marginLeft: 12 }}><Progress percent={Math.round(item.progress || 0)} size="small" /></div>}
              </div>
            </List.Item>
          );
        }}
      />
    </Modal>
  );
};

export default ProductImagesManager;
