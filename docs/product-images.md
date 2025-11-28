++"""markdown
# 商品图片管理 API

支持基于 `ProductImage` 的图片管理，包含排序与主图标记。

1. 获取商品图片列表
   - 路径：`GET /api/v1/products/:id/images`
   - 权限：公开（可根据需要改为仅管理员）
   - 返回：
```json
{
  "code": 0,
  "data": [
    {"id": 1, "product_id": 10, "image_url": "https://...", "sort": 0, "is_main": true},
    ...
  ]
}
```

2. 添加商品图片
   - 路径：`POST /api/v1/products/:id/images`
   - 权限：需要登录且为 `admin` 角色
   - 请求体：
```json
{
  "image_url": "https://...",
  "sort": 0,
  "is_main": false
}
```
   - 返回：新建的图片记录

3. 更新图片（排序 / 主图）
   - 路径：`PUT /api/v1/products/:id/images/:image_id`
   - 权限：需要登录且为 `admin` 角色
   - 请求体（至少包含一个字段）：
```json
{
  "sort": 10,
  "is_main": true
}
```
   - 说明：当 `is_main` 设为 `true` 时，后端会自动取消同一商品下其他图片的 `is_main` 标记，保证唯一主图。

4. 删除图片
   - 路径：`DELETE /api/v1/products/:id/images/:image_id`
   - 权限：需要登录且为 `admin` 角色

注意：图片文件本身的存储/删除应调用 OSS 接口（如 `/api/v1/oss/delete`）。后端已实现自动同步：在删除单张图片或删除商品时，后端会尝试调用 OSS 删除对应文件（best-effort），建议前端仍可在必要时主动调用 OSS 删除以做双重保证。

示例：
- 推荐在商品详情页展示 `images_list` 字段，并在 Admin 商品编辑页提供排序与设为主图操作。

"""