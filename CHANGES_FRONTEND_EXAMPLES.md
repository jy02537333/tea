# 前端示例变更记录

## 用户信息与概要字段对齐（统一来源：users/me/summary）
- 目的：统一前端消费来源，减少多接口拼装与契约分裂。
- 推荐做法：优先使用 `GET /api/v1/users/me/summary` 获取用户等级与权益等聚合字段；`GET /api/v1/user/info` 仍可用，但其字段已与概要对齐。

### 字段清单（两接口对齐）
- 基本：`id`、`uid`、`open_id`、`nickname`、`avatar`、`phone`、`gender`、`balance`、`points`
- 会员/等级：`membership_package_id`、`partner_level_id`、`membership_level_name`
- 权益：`discount_rate`、`purchase_discount_rate`、`direct_commission_rate`、`team_commission_rate`、`upgrade_reward_rate`

### 前端契约建议
- 在类型定义中加入上述字段，并在调用层统一从 `users/me/summary` 读取，用于“我的”页或用户卡片展示。
- 如需兼容 `user/info` 的既有逻辑，可复用同一类型（字段一致）。

示例（TypeScript 类型片段）：

```ts
export interface UserProfile {
	id: number;
	uid?: string;
	open_id?: string;
	nickname?: string;
	avatar?: string;
	phone?: string;
	gender?: number;
	balance?: number;
	points?: number;
	membership_package_id?: number | null;
	partner_level_id?: number | null;
	membership_level_name?: string;
	discount_rate?: number;
	purchase_discount_rate?: number;
	direct_commission_rate?: number;
	team_commission_rate?: number;
	upgrade_reward_rate?: number;
}
```

> 注：当不存在会员或等级时，`membership_level_name` 返回 `visitor`，折扣默认 `1.0`，返利默认 `0`。
CHANGES: Frontend examples scaffold and typing improvements

Summary:
- Added Taro example pages and configuration for `wx-fe`.
- Converted example services to use concrete TypeScript types and PaginatedResponse/ApiResponse wrappers.
- Added/expanded domain types: ProductSku, OrderItem, Order, Coupon, Store, User, AuthResponse, etc.
- Updated admin-fe example pages to use domain types and improved response extraction logic.

Files changed (high-level):
- wx-fe/src/app.config.ts (existing) and new wx-fe/src/pages.json (added)
- wx-fe/src/pages/product-list/index.tsx (updated to Taro + typed)
- wx-fe/src/pages/order-detail/index.tsx (updated to Taro + typed)
- wx-fe/src/pages/README_EXAMPLES.md (updated run steps)
- wx-fe/src/services/*.ts (auth, products, orders, coupons, cart, stores) - updated typing and return types
- admin-fe/src/pages/ProductList/index.tsx (minor fixes)
- admin-fe/src/pages/OrderDetail/index.tsx (typed)
- admin-fe/src/pages/README_EXAMPLES.md (run steps added)
- admin-fe/src/services/*.ts (orders, rbac, stores, coupons, payments, accrual) - typed
- admin-fe/src/services/types.ts (expanded domain types)

Suggested commit message:

feat(frontend): scaffold Taro example pages and tighten service typings

- Add Taro page config (`wx-fe/src/app.config.ts` / `wx-fe/src/pages.json`) and example pages
- Replace `any` responses with concrete types in `wx-fe` and `admin-fe` services
- Add domain types (Order, OrderItem, ProductSku, Coupon, Store, AuthResponse, etc.)
- Update README_EXAMPLES with run steps for admin & wx examples

Reviewer notes:
- This change focuses on scaffolding and typing; runtime verification requires installing dependencies and running each project (see README_EXAMPLES in respective folders).
- Export/download endpoints still return blobs and are left as `responseType: 'blob'` and return raw `res.data`.



---

CHANGES: Parse withdrawal remark JSON in admin & mini-program

Summary:
- Admin frontend lists now parse and display standardized withdrawal `remark` JSON fields.
- CSV exports include parsed fields for finance review and reconciliation.
- Mini-program (wx-fe) store finance list safely parses and shows key fields.
- Sprint doc updated (B-07) to reflect frontend/admin parsing.

Files changed (high-level):
- admin-fe/src/pages/StoreFinance.tsx (add parsed columns and CSV fields)
- admin-fe/src/pages/Stores.tsx (add parsed columns and CSV fields)
- wx-fe/src/pages/store-finance/index.tsx (safe JSON parsing and display)
- doc/prd_sprints.md (update B-07 note)

Suggested commit message:

feat(finance): parse withdrawal remark JSON in admin/wx; extend CSV export

- Parse and display `phase`, `withdraw_no`, `currency`, `amount_cents`, `fee_cents`, `net_cents`
- Extend admin CSV export to include parsed remark JSON fields
- Add safe parsing in wx-fe store finance page (graceful fallback on invalid JSON)
- Update Sprint B (B-07) documentation note

Reviewer notes:
- Parsing uses try/catch; invalid/non-JSON remarks fall back gracefully without breaking UI.
- Verify type checks and builds for both `admin-fe` and `wx-fe` after installing deps.
- Consider extracting a shared remark JSON parse utility to reduce duplication across views.


