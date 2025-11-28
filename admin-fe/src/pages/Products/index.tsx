import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Modal, message, Form, Input, InputNumber, Select, Upload } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { getOssSignature } from '../../services/oss';
import { deleteOssFiles } from '../../services/oss-delete';
import { getProducts, createProduct, updateProduct, deleteProduct, updateProduct as updateProductStatus, Product } from '../../services/products';
import type { UploadFile } from 'antd/es/upload/interface';
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [filter, setFilter] = useState<{ category_id?: number; status?: number; keyword?: string }>({});
import { getCategories } from '../../services/categories';
import ProductImagesManager from '../../components/ProductImagesManager';

const PAGE_SIZE = 20;

const Products: React.FC = () => {
  const [data, setData] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit' | null>(null);
  const [current, setCurrent] = useState<Product | null>(null);
  const [imageManagerProduct, setImageManagerProduct] = useState<Product | null>(null);
  const [imageManagerVisible, setImageManagerVisible] = useState(false);
  const [form] = Form.useForm();
  const [categories, setCategories] = useState<{ label: string; value: number }[]>([]);

  useEffect(() => {
    fetch(page, filter);
    fetchCategories();
  }, [page, filter]);

  async function fetch(pageNum: number, filterParams: any) {
    setLoading(true);
    try {
      const res = await getProducts({ page: pageNum, limit: PAGE_SIZE, ...filterParams });
      setData(res.data);
      setTotal(res.total || 0);
    } catch (e: any) {
      message.error(e?.message || '加载失败');
    } finally {
      setLoading(false);
    }
  }
  // 图片上传相关（阿里云 OSS）
  function normFile(e: any) {
    if (Array.isArray(e)) return e;
    return e && e.fileList;
  }

  // 兼容后端 images 可能为 string 或 string[]
  function normalizeImages(images: any): string[] {
    if (!images) return [];
    if (Array.isArray(images)) return images;
    try {
      const arr = JSON.parse(images);
      if (Array.isArray(arr)) return arr;
    } catch {}
    if (typeof images === 'string') return [images];
    return [];
  }

  function getImageUrls(fileList: UploadFile[]) {
    if (!fileList || fileList.length === 0) return [];
    return fileList.map(f => f.url || f.thumbUrl || (f.response && f.response.url)).filter(Boolean);
  }

  async function customOssUpload({ file, onSuccess, onError, onProgress }) {
    try {
      // 获取 OSS 签名
      const sign = await getOssSignature({ dir: 'products/' });
      const formData = new FormData();
      const filename = `${sign.dir}${Date.now()}_${file.name}`;
      formData.append('key', filename);
      formData.append('OSSAccessKeyId', sign.accessid);
      formData.append('policy', sign.policy);
      formData.append('Signature', sign.signature);
      formData.append('success_action_status', '200');
      formData.append('file', file);
      // 上传到 OSS
      const xhr = new XMLHttpRequest();
      xhr.open('POST', sign.host, true);
      xhr.upload.onprogress = (evt) => {
        if (evt.lengthComputable && onProgress) {
          onProgress({ percent: (evt.loaded / evt.total) * 100 });
        }
      };
      xhr.onload = () => {
        if (xhr.status === 200) {
          // 拼接图片完整 URL
          const url = sign.host + '/' + filename;
          onSuccess && onSuccess({ url }, file);
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
  async function handleBatchStatus(status: number) {
    if (!selectedRowKeys.length) return message.warning('请先选择商品');
    Modal.confirm({
      title: `确认将选中商品批量${status === 1 ? '上架' : '下架'}？`,
      onOk: async () => {
        try {
          await Promise.all(selectedRowKeys.map(id => updateProductStatus(Number(id), { status })));
          message.success('批量操作成功');
          setSelectedRowKeys([]);
          fetch(page, filter);
        } catch (e: any) {
          message.error(e?.message || '批量操作失败');
        }
      },
    });
  }
      <Form
        layout="inline"
        style={{ marginBottom: 16 }}
        onFinish={v => setFilter(v)}
        initialValues={filter}
      >
        <Form.Item name="keyword" label="关键词">
          <Input placeholder="商品名/ID" allowClear />
        </Form.Item>
        <Form.Item name="category_id" label="分类">
          <Select allowClear style={{ width: 120 }} options={categories} />
        </Form.Item>
        <Form.Item name="status" label="状态">
          <Select allowClear style={{ width: 100 }} options={[{ label: '上架', value: 1 }, { label: '下架', value: 0 }]} />
        </Form.Item>
        <Form.Item>
          <Button htmlType="submit" type="primary">筛选</Button>
        </Form.Item>
      </Form>

  async function fetchCategories() {
    try {
      const res = await getCategories();
      setCategories(res.map((c) => ({ label: c.name, value: c.id })));
    } catch {}
  }

  function openCreate() {
    setEditMode('create');
    setCurrent(null);
    form.resetFields();
    setModalOpen(true);
  }

  function openEdit(record: Product) {
    setEditMode('edit');
    setCurrent(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setCurrent(null);
    setEditMode(null);
    form.resetFields();
  }

  async function handleSave() {
    try {
      const values = await form.validateFields();
      // 处理图片字段，转为图片url数组
      let images: string[] = [];
      if (Array.isArray(values.images)) {
        images = values.images.map((f: any) => f.url || (f.response && f.response.url)).filter(Boolean);
      }
      const payload = { ...values, images: JSON.stringify(images) };
      if (editMode === 'create') {
        await createProduct(payload);
        message.success('新建成功');
      } else if (editMode === 'edit' && current) {
        await updateProduct(current.id, payload);
        message.success('保存成功');
      }
      closeModal();
      fetch(page);
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    }
  }

  async function handleDelete(record: Product) {
    Modal.confirm({
      title: '确认删除该商品？',
      onOk: async () => {
        try {
          await deleteProduct(record.id);
          message.success('已删除');
          fetch(page);
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  return (
    <div style={{ padding: 16 }}>
      <h1>商品管理</h1>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={openCreate}>新建商品</Button>
        <Button onClick={() => handleBatchStatus(1)} disabled={!selectedRowKeys.length}>批量上架</Button>
        <Button onClick={() => handleBatchStatus(0)} disabled={!selectedRowKeys.length}>批量下架</Button>
      </Space>
      <Table<Product>
        rowKey="id"
        loading={loading}
        dataSource={data}
        rowSelection={{ selectedRowKeys, onChange: setSelectedRowKeys }}
        columns={[
          { title: 'ID', dataIndex: 'id' },
          { title: '名称', dataIndex: 'name' },
          { title: '分类', dataIndex: 'category_id', render: (id) => categories.find(c => c.value === id)?.label || id },
          { title: '价格', dataIndex: 'price' },
          { title: '原价', dataIndex: 'original_price' },
          { title: '库存', dataIndex: 'stock' },
          { title: '图片', dataIndex: 'images', render: (imgs) => {
            const arr = normalizeImages(imgs);
            return arr.length > 0 ? arr.map((url, i) => <img key={i} src={url} alt="" style={{ width: 40, height: 40, objectFit: 'cover', marginRight: 4 }} />) : '-';
          } },
          { title: '状态', dataIndex: 'status', render: (s) => s === 1 ? '上架' : '下架' },
          {
            title: '操作',
            render: (_, record) => (
              <Space>
                <Button size="small" onClick={() => openEdit(record)}>编辑</Button>
                <Button size="small" onClick={() => { setImageManagerProduct(record); setImageManagerVisible(true); }}>图片管理</Button>
                <Button size="small" danger onClick={() => handleDelete(record)}>删除</Button>
              </Space>
            ),
          },
        ]}
        pagination={{
          current: page,
          pageSize: PAGE_SIZE,
          total,
          onChange: setPage,
          showTotal: (t) => `共 ${t} 条`,
        }}
      />
      <ProductImagesManager
        productId={imageManagerProduct?.id || null}
        visible={imageManagerVisible}
        onClose={() => setImageManagerVisible(false)}
        onUpdated={() => { fetch(page, filter); }}
      />
      <Modal
        open={modalOpen}
        title={editMode === 'create' ? '新建商品' : '编辑商品'}
        onCancel={closeModal}
        footer={[
          <Button key="cancel" onClick={closeModal}>取消</Button>,
          <Button key="save" type="primary" onClick={handleSave}>保存</Button>,
        ]}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入商品名称' }]}> <Input /> </Form.Item>
          <Form.Item name="category_id" label="分类" rules={[{ required: true, message: '请选择分类' }]}> <Select options={categories} /> </Form.Item>
          <Form.Item name="price" label="价格" rules={[{ required: true, message: '请输入价格' }]}> <InputNumber min={0} style={{ width: '100%' }} /> </Form.Item>
          <Form.Item name="original_price" label="原价"> <InputNumber min={0} style={{ width: '100%' }} /> </Form.Item>
          <Form.Item name="stock" label="库存"> <InputNumber min={0} style={{ width: '100%' }} /> </Form.Item>
          <Form.Item
            name="images"
            label="图片"
            valuePropName="fileList"
            getValueFromEvent={normFile}
            extra="支持多图上传到阿里云 OSS，删除图片将同步删除 OSS 文件"
          >
            <Upload
              listType="picture"
              maxCount={5}
              multiple
              customRequest={customOssUpload}
              accept="image/*"
              onRemove={async (file) => {
                // 删除图片时调用后端删除接口
                const url = file.url || (file.response && file.response.url);
                if (url) {
                  try {
                    await deleteOssFiles([url]);
                  } catch (e) {
                    message.error('OSS图片删除失败');
                  }
                }
              }}
              onChange={({ fileList }) => {
                // 上传成功后将图片 URL 写入表单
                form.setFieldValue('images', fileList.map(f => ({ ...f, url: f.response?.url || f.url })));
              }}
            >
              <Button icon={<UploadOutlined />}>选择图片</Button>
            </Upload>
          </Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}> <Select options={[{ label: '上架', value: 1 }, { label: '下架', value: 0 }]} /> </Form.Item>
        </Form>
        {/* 多图本地预览 */}
        {form.getFieldValue('images') && form.getFieldValue('images').length > 0 && (
          <div style={{ marginTop: 8 }}>
            {form.getFieldValue('images').map((f: any, i: number) => (
              <img key={i} src={f.url || (f.response && f.response.url)} alt="预览" style={{ width: 80, height: 80, objectFit: 'cover', marginRight: 8 }} />
            ))}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Products;
