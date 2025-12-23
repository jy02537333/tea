# Sprint C 拆分规划（SC1–SC7）

目标：将原 PR #44 的大体量改动拆分为若干可审阅、可回滚的小 PR，逐步合入 master，并让 CI 的最小联调与非阻塞校验起作用。

## 通用约定
- 分支命名：`feat/sc{n}-<slug>`（如：`feat/sc1-oss-policy`）
- PR 标题：`SC{n}: <简要说明>`；标签：`Sprint C`
- PR 内容：范围说明、涉及文件、验证方式、回滚影响、与其他 PR 的依赖关系
- CI：保持非阻塞最小联调；上传 `build-ci-logs/**`；若触发校验，附带摘要评论

---

## SC1: OSS 直传 Policy 与工具函数
- 范围：
  - 新增/确认后端 `POST /api/v1/admin/storage/oss/policy`（仅管理员）
  - `HMAC-SHA1` 与 Base64 工具函数（若 master 已有则复用）
  - 文档：说明直传参数与演示 cURL
- 涉及文件（示例）：
  - `tea-api/internal/router/router.go`（route 挂载）
  - `tea-api/internal/handler/storage/oss.go`（若无则新增）
  - `tea-api/pkg/utils/common.go`（工具复用或补充）
  - `docs/ci/minimal-integration.md` 追加直传小节
- 验证：
  - `make run-min-integration` 含 OSS policy 软性校验（not_ready/not_found 允许）

## SC2: Referral（邀请/分销关系）基础接口
- 范围：记录推荐关系、查询统计；不含复杂深度聚合
- 文件：`internal/handler/referral/*`、`router.go`、必要模型（若采用 SQL/migration，先最小化）
- 验证：最小联调脚本新增 1–2 条接口校验

## SC3: Commission 引擎（冻结→可用→已结算）基础版
- 范围：
  - 仅直推（Direct）佣金计算与创建；订单号幂等
  - 状态迁移：`frozen`→`available`（简化版定时/手工触发）
- 文件：`internal/handler/commission/*`、`router.go`
- 验证：最小联调新增创建与汇总查询

## SC4: Membership/Partner 升级与基础档位
- 范围：购买套餐触发升级、档位查询；先不引入复杂权益
- 文件：`internal/handler/partner/*`、`router.go`、必要模型
- 验证：最小联调新增 1 条接口校验

## SC5: Commission 间推与升级奖励（可选强化）
- 范围：在 SC3 基础上补充间推/升级奖励计算路径
- 验证：与 SC3 用例扩展，不新增破坏性接口

## SC6: 门店后台与打印（轻量）
- 范围：订单接受/拒绝基础接口；打印任务创建（占位实现）
- 验证：仅通路校验，返回占位数据

## SC7: 文档与路由清理、重构收尾
- 范围：
  - 路由表与注释统一
  - 文档交叉链接修正（`doc/prd_sprints.md` 与 `docs/ci/*`）
  - 前端 .env.example 与 README 再梳理

---

## 依赖与顺序建议
1. SC1 → 直传工具与最小联调先到位，降低后续阻力
2. SC2 → 关系建模为佣金奠定数据前提
3. SC3 → 基础佣金路径先跑通
4. SC4 → 角色/档位补齐
5. SC5/SC6 → 逐步增强，优先度次之
6. SC7 → 收尾统一

## 开发提示
- 避免一次性引入所有模型字段，按需最小集
- 路由与 Handler 拆分子包，减少冲突
- CI 失败不阻塞合并，但摘要会提示回归点
