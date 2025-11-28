# Admin-FE 图片删除与主图操作示例

下面给出一个简单的 React + Ant Design 示例（可用于 admin-fe），演示如何在商品图片列表中添加“删除”与“设为主图”交互。

- 后端接口：
  - 删除图片：`DELETE /api/v1/products/:id/images/:image_id`（需要管理员权限）
  - 设为主图/更新排序：`PUT /api/v1/products/:id/images/:image_id`（需要管理员权限）
  - 说明：后端会自动在删除时尝试同步删除 OSS 上的文件（best-effort），前端无需额外调用 OSS 删除接口；若你希望前端主动删除 OSS，请先 `POST /api/v1/oss/delete`。

示例代码片段：

```tsx
import React from 'react'
import { Button, List, Modal, message } from 'antd'
import axios from 'axios'

const ProductImages = ({ productId, images, onUpdated }) => {
  const handleDelete = async (imageId) => {
    Modal.confirm({
      title: '确认删除该图片？',
      onOk: async () => {
        try {
          await axios.delete(`/api/v1/products/${productId}/images/${imageId}`)
          message.success('删除成功')
          onUpdated && onUpdated()
        } catch (err) {
          message.error(err?.response?.data?.message || '删除失败')
        }
      }
    })
  }

  const handleSetMain = async (imageId) => {
    try {
      await axios.put(`/api/v1/products/${productId}/images/${imageId}`, { is_main: true })
      message.success('设为主图成功')
      onUpdated && onUpdated()
    } catch (err) {
      message.error(err?.response?.data?.message || '操作失败')
    }
  }

  return (
    <List
      dataSource={images}
      renderItem={item => (
        <List.Item
          actions={[
            <Button danger onClick={() => handleDelete(item.id)}>删除</Button>,
            <Button type="link" onClick={() => handleSetMain(item.id)}>设为主图</Button>
          ]}
        >
          <img src={item.image_url} alt="" style={{ width: 120 }} />
        </List.Item>
      )}
    />
  )
}

export default ProductImages
```

说明与建议：
- UI 操作由管理员发起，后端会做权限校验；前端应确保当前用户为 admin（或通过后端返回的 403 处理）。
- 删除操作后端会先尝试删除 OSS 文件并最后删除 DB 记录（best-effort）。如果你想要更快的 UX，可以先即时在前端移除图片，再在后台异步完成删除。
- 推荐同时在 Admin 后台记录图片变更日志，便于追踪误删。