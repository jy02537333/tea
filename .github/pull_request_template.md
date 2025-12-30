# Pull Request Checklist

请在提交前勾选与本次改动相关的检查项；CI 将依据这些标准化表述自动匹配并在实现到位时勾选为完成。

## 通用
- [ ] 更新相关文档索引（README、PRD、设计说明）
- [ ] 证据产物已生成并可访问（build-ci-logs/* 或外链）

## 后端（示例）
- [ ] 后端实现 POST /api/v1/admin/storage/oss/policy（后续切换直传策略）

## 小程序（wx-fe）
- [ ] 扫码进店（最小版）
  - 入口参数解析：在 app 级解析 `store_id`（query/scene/H5 URL），持久化到 `current_store_id`
  - 页面接管：首页/分类接管路由 `store_id` 并同步到存储；商品详情在存在 `store_id` 时也持久化
  - 一致性：订单列表/详情、购物车、结算按当前门店维度展示与过滤

- [ ] 门店详情：证照、导航/拨号、门店商品映射
  - 证照/基础信息：展示门店证照与基本信息
  - 导航/拨号：支持一键导航与拨号
  - 商品映射：门店商品与外卖列表的映射与展示

> 注：本模板中的关键行将被 CI 工作流自动识别并在实现到位后替换为已勾选状态（见 .github/workflows/update-pr-checklist.yml）。
