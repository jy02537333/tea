# OSS 图片管理接口说明

## 1. 获取 OSS 文件列表
- 路径：`GET /api/v1/oss/list`
- 参数：
  - `prefix`（可选）：文件名前缀过滤
  - `marker`（可选）：翻页游标
  - `limit`（可选）：每页数量，默认100
- 返回：
```json
{
  "files": ["url1", "url2", ...],
  "next_marker": "下一页游标"
}
```

## 2. 批量删除 OSS 文件
- 路径：`POST /api/v1/oss/delete`
- 参数：
```json
{
  "urls": ["url1", "url2", ...]
}
```
- 返回：
```json
{
  "code": 0,
  "message": "success"
}
```

## 说明
- 需登录且仅管理员（admin 角色）可调用，建议后台管理端使用。
- 支持多图批量管理、删除。
- 建议前端分页加载、批量操作。
