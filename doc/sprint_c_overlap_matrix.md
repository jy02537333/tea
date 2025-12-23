# Sprint C 重叠矩阵与拆分合并计划（截至 2025-12-23）

本文档汇总当前 master 与开放 PR 在 Sprint C 范围内的功能覆盖与重叠情况，并给出建议的拆分与合并计划，以加速安全交付。

## 一、基线现状（master）

- 已有：
  - 管理端会员套餐与合伙人等级管理（tea-api/internal/handler/membership_admin.go，路由已注册）。
  - 管理端佣金运维：手动释放与订单反冲（tea-api/internal/handler/commission_admin.go，路由位于 /api/v1/admin/finance/commission/...）。
  - 上传服务与 Aliyun OSS SDK 使用基础（上传文件接口存在）。
  - 钱包、提现、订单、门店等基础域已实现。
- 缺失：
  - 推荐关系（Referral）用户侧接口与闭包表。
  - 佣金用户侧接口（预览/创建/列表/汇总）。
  - 会员/合伙人礼包购买与升级（用户侧）。
  - 门店后台接单/拒单/打印任务与订单日志模型。
  - 管理端 OSS 直传策略接口（POST /api/v1/admin/storage/oss/policy）。
  - User 模型扩展字段（membership_package_id、partner_level_id）及对应迁移。
  - Sprint C 的专属实现说明文档。

## 二、开放 PR 概览（与 Sprint C 关联）

- PR #44（Draft）：Sprint C 核心实现（分销、佣金、合伙人、门店后台、OSS 直传策略、数据模型与迁移、路由与文档）。
- PR #50（Ready）：最小联调与 CI 加强；财务提现 remark JSON 解析与前端导出增强；不直接引入 Sprint C 用户侧接口。
- PR #51（Ready）：评审清单与 CI 文档工作流；纯文档/流程增强。
- PR #45/#47/#46：围绕 Sprint A/B 的测试与工具增强；不引入 Sprint C 用户侧接口。

## 三、Sprint C 功能重叠矩阵

| 模块 | master | PR #44 | PR #50 | PR #51 | 备注 |
| --- | --- | --- | --- | --- | --- |
| 推荐关系（Referral）用户侧接口与闭包表 | 缺失 | 新增 handler/referral.go、model.Referral/ReferralClosure、路由 | 无 | 无 | 依赖闭包表；对佣金计算有前置关系 |
| 佣金（用户侧：预览/创建/列表/汇总） | 缺失 | 新增 handler/commission.go、路由 | CI 中有“commission minimal flow”步骤（基于管理端释放） | 无 | 与 master 的管理端释放/反冲共存；需路由/命名对齐 |
| 管理端佣金运维（释放/反冲） | 已有（commission_admin.go 路由在 finance/commission） | 新增 /api/v1/admin/commissions/unfreeze | 复用 master | 无 | 存在路由与命名重叠，建议统一为 master 既有风格 |
| 会员/合伙人礼包购买与升级（用户侧） | 缺失 | 新增 handler/partner.go（列表/购买/升级）、路由 | 无 | 无 | 依赖用户扩展字段与佣金派发 |
| 合伙人等级管理（管理端） | 已有（membership_admin.go） | 亦实现（partner.go 管理端段落） | 无 | 无 | 存在重复实现，建议沿用 master 版本，PR #44 删除重复 |
| 门店后台（接单/拒单/打印） | 缺失 | 新增 handler/partner.go + PrintHandler、路由 | 无 | 无 | 引入 OrderLog 模型；打印任务通过队列占位 |
| 管理端 OSS 直传策略 | 缺失 | 新增 upload.go:GetOSSPolicy + service/upload.go:GenerateOSSPolicy | 文档提及最小集成，但未提供路由 | 无 | 低风险、解耦良好，推荐先合并 |
| User 模型扩展字段与迁移 | 缺失 | 新增字段于 model/base.go + 迁移 20251215_add_partner_fields_to_users.sql | 无 | 无 | 需谨慎迁移与幂等；为伙伴/佣金用户侧前置 |
| OrderLog 模型 | 缺失 | 新增于 model/commission.go | 无 | 无 | 门店操作与退款路径记录所需 |
| Sprint C 说明文档 | 缺失 | 新增 tea-api/docs/SPRINT_C_README.md | 无 | 有评审流程文档 | 建议在各模块 PR 中同步维护 |

## 四、拆分与合并计划（建议序列）

为降低风险并通过 A-first 的主干门禁，建议将 PR #44 拆分为 7 个小 PR，按序合并：

1) PR-SC1：管理端 OSS 直传策略
- 范围：upload.go:GetOSSPolicy、service/upload.go:GenerateOSSPolicy、pkg/utils/common.go（Base64/HMAC-SHA1）。
- 依赖/风险：与现有上传服务兼容；不改动数据结构；低风险。
- 验收：本地最小联调生成策略 JSON；文档 doc/api/oss.md/tea-api/docs 同步说明。

2) PR-SC2：推荐关系（Referral）与闭包表（用户侧）
- 范围：handler/referral.go、model.Referral/ReferralClosure、路由与迁移（若需）。
- 依赖/风险：新增表结构与只增不破的接口；对现有功能无破坏。
- 验收：列表/统计接口可用；提供最小集成脚本与证据文件。

3) PR-SC3：用户模型扩展字段与迁移（独立）
- 范围：membership_package_id、partner_level_id 字段及迁移脚本。
- 依赖/风险：需幂等与回滚方案；上线前确认 DBA 对齐索引策略。
- 验收：迁移成功且不影响既有查询；字段为空默认兼容。

4) PR-SC4：佣金（用户侧）接口
- 范围：handler/commission.go、用户侧路由；复用 master 的管理端释放/反冲实现，不新增重复路由。
- 依赖/风险：依赖 PR-SC2（闭包表）与 PR-SC3（用户字段）；与订单域读取只读。
- 验收：预览/创建/列表/汇总可用；最小联调脚本与证据归档（非阻断）。

5) PR-SC5：会员/合伙人购买与升级（用户侧）
- 范围：partner.go 用户侧接口与路由；删除/忽略与 master 重叠的管理端等级接口。
- 依赖/风险：依赖 PR-SC3（用户字段）；与佣金派发有耦合，但可先返回成功路径并生成占位佣金。
- 验收：购买/升级成功路径打通；最小联调与证据归档。

6) PR-SC6：门店后台接单/拒单/打印任务 + 订单日志
- 范围：accept/reject 接口、PrintHandler 创建打印任务、OrderLog 持久化与路由。
- 依赖/风险：与订单域事务一致性需注意；退款路径可先走占位/队列。
- 验收：成功/拒单路径可用；打印任务生成；日志落库；最小联调与证据归档。

7) PR-SC7：Sprint C 文档与路由对齐
- 范围：tea-api/docs/SPRINT_C_README.md 与 router.go 路由清理/对齐 master 风格（避免与 finance/commission 既有路由冲突）。
- 依赖/风险：与上述各模块变更同步落地；保持文档与代码一致。
- 验收：文档完整、路由统一、CI 绿。

## 五、合并顺序与门禁建议

- 先合并 PR #51（评审与 CI 文档）与 PR #50（最小联调/CI 与前端解析增强），确保评审与验证通路畅通。
- 按上述 PR-SC1 → PR-SC2 → PR-SC3 → PR-SC4 → PR-SC5 → PR-SC6 → PR-SC7 顺序推进，保证前置依赖就绪。
- 路由对齐策略：
  - 佣金管理端：沿用 master 路由（/api/v1/admin/finance/commission/...），不新增 /admin/commissions/unfreeze 重复路由。
  - 合伙人等级管理端：沿用 membership_admin.go 既有实现，删除重复代码，仅新增用户侧能力。

## 六、风险与验证

- 迁移风险：用户字段与闭包表需幂等与索引；建议先在测试库演练。
- 路由冲突：PR #44 的部分管理端路由与 master 重叠，拆分时统一到 master 风格。
- 证据与最小联调：为每个子 PR 提供最小联调脚本（bash）与产物（build-ci-logs/**），Sprint A/B 检查仍为阻断门禁；Sprint C 证据暂定非阻断。

---

维护者：项目协作组（Sprint C 交付加速）