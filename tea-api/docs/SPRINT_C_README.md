# Sprint C Implementation - 分销与合伙人系统

## 概述

Sprint C 实现了茶心阁小程序的核心分销体系、佣金计算、合伙人升级和门店后台管理功能。

## 已实现功能

### 1. 推荐关系管理 (Referral System)

#### API 端点

- `POST /api/v1/referral/record` - 记录推荐关系
- `GET /api/v1/users/:user_id/referral-info` - 获取用户推荐信息
- `GET /api/v1/users/:user_id/referred-users` - 获取直推用户列表
- `GET /api/v1/users/:user_id/referral-stats` - 获取推荐统计数据

#### 功能特点

- 支持推荐关系记录（直推）
- 使用闭包表(referrals_closure)维护推荐链路
- 支持多层级推荐关系查询
- 防止自己推荐自己
- 推荐关系一旦建立不可修改

### 2. 佣金计算与管理 (Commission System)

#### API 端点

- `POST /api/v1/commissions/calculate` - 计算佣金（预览）
- `POST /api/v1/commissions` - 创建佣金记录
- `GET /api/v1/users/:user_id/commissions` - 查询用户佣金列表
- `GET /api/v1/users/:user_id/commissions/summary` - 获取佣金汇总
- `POST /api/v1/admin/commissions/unfreeze` - 解冻佣金（管理端）

#### 佣金计算规则

**计算基数**:
```
结算基数 = 订单商品总额 - 运费 - 优惠券 - 折扣
```

**佣金类型**:

1. **直接佣金 (direct)**: 
   - 计算: `floor(结算基数 * 直推比例)`
   - 直推比例根据用户等级确定（10%-45%）

2. **间接佣金 (indirect)**: 
   - 计算: `floor(直推佣金 * 团队管理奖比例)`
   - 团队管理奖比例默认为 10%
   - 发放给直推人的上级

3. **升级奖励 (upgrade)**:
   - 用户升级合伙人时给推荐人的奖励
   - 计算: `floor(礼包价格 * 升级奖励比例)`

**佣金状态流转**:
```
frozen (冻结) → available (可提现) → paid (已提现)
```

- 新生成的佣金默认状态为 `frozen`
- 7天后可解冻为 `available`
- 提现后状态变为 `paid`

### 3. 合伙人/会员系统 (Partner/Membership System)

#### API 端点

- `GET /api/v1/partner/packages` - 获取礼包列表
- `POST /api/v1/membership/purchase` - 购买会员礼包
- `POST /api/v1/partner/upgrade` - 升级合伙人
- `GET /api/v1/admin/partner-levels` - 获取合伙人等级列表（管理端）
- `POST /api/v1/admin/partner-levels` - 创建合伙人等级（管理端）
- `PUT /api/v1/admin/partner-levels/:id` - 更新合伙人等级（管理端）

#### 功能特点

- 支持多种会员/合伙人礼包
- 购买礼包后自动升级用户等级
- 发放茶币奖励
- 计算并发放升级奖励佣金给推荐人
- 升级后享受对应等级的佣金比例和拿货折扣

### 4. 门店后台管理 (Store Backend)

#### API 端点

- `GET /api/v1/stores/:store_id/orders` - 获取门店订单列表
- `POST /api/v1/stores/:store_id/orders/:order_id/accept` - 接单
- `POST /api/v1/stores/:store_id/orders/:order_id/reject` - 拒单
- `POST /api/v1/stores/:store_id/print/jobs` - 创建打印任务

#### 功能特点

- 门店订单管理（列表、筛选、分页）
- 接单/拒单功能
- 订单日志记录
- 打印任务创建（支持后厨单、收据、标签）

### 5. OSS 直传支持 (OSS Direct Upload)

#### API 端点

- `POST /api/v1/admin/storage/oss/policy` - 获取 OSS 上传策略

#### 功能特点

- 生成带签名的 OSS 上传策略
- 支持多种业务类型（商品图片、品牌Logo、门店图片）
- 自动生成对象key前缀（按日期组织）
- 30分钟过期时间

#### 使用示例

```javascript
// 请求上传策略
POST /api/v1/admin/storage/oss/policy
{
  "business": "product_image"
}

// 响应
{
  "code": 0,
  "data": {
    "access_key_id": "xxx",
    "policy": "base64_encoded_policy",
    "signature": "hmac_sha1_signature",
    "expire": 1702123456,
    "host": "https://bucket.oss-cn-hangzhou.aliyuncs.com",
    "object_key_prefix": "admin/products/2024/12/15/"
  }
}
```

## 数据模型

### 新增模型

1. **Referral** - 推荐关系表
2. **ReferralClosure** - 推荐闭包表
3. **Commission** - 佣金记录
4. **CommissionTransaction** - 佣金流水
5. **MembershipPackage** - 会员/合伙人礼包
6. **PartnerLevel** - 合伙人等级
7. **UserBankAccount** - 用户银行卡
8. **Wallet** - 钱包
9. **WalletTransaction** - 钱包流水
10. **OrderLog** - 订单日志

### User 模型扩展

```go
type User struct {
    // ... 原有字段
    MembershipPackageID *uint `gorm:"index" json:"membership_package_id"`
    PartnerLevelID      *uint `gorm:"index" json:"partner_level_id"`
}
```

## 数据库迁移

执行以下迁移文件：

```sql
tea-api/migrations/20251215_add_partner_fields_to_users.sql
```

或手动执行：

```sql
ALTER TABLE `users` ADD COLUMN `membership_package_id` BIGINT UNSIGNED NULL;
ALTER TABLE `users` ADD COLUMN `partner_level_id` BIGINT UNSIGNED NULL;
ALTER TABLE `users` ADD INDEX `idx_users_membership_package` (`membership_package_id`);
ALTER TABLE `users` ADD INDEX `idx_users_partner_level` (`partner_level_id`);
```

## 测试

运行佣金服务测试：

```bash
cd tea-api
go test ./internal/service/commission/... -v
```

运行 handler 测试：

```bash
go test ./internal/handler/... -v
```

## 集成说明

### 在订单完成时触发佣金计算

在订单状态变更为"已完成"时，应该调用佣金创建接口：

```go
// 订单完成后的处理逻辑
if order.Status == "completed" {
    // 查询是否有推荐人
    var referral Referral
    if db.Where("referred_user_id = ?", order.UserID).First(&referral).Error == nil {
        // 调用佣金计算接口
        // POST /api/v1/commissions
    }
}
```

### 定时任务解冻佣金

建议使用 cron 定时任务每天执行一次佣金解冻：

```bash
# 每天凌晨2点执行
0 2 * * * curl -X POST http://localhost:9292/api/v1/admin/commissions/unfreeze
```

## 待完成功能

- [ ] 提现申请和审核流程
- [ ] 银行卡管理接口
- [ ] 订单完成自动触发佣金计算
- [ ] 推荐码/分享链接生成
- [ ] 打印队列与实际打印机对接
- [ ] 订单通知推送（WebSocket/轮询）

## API 文档

详细的 API 文档请参考：

- `doc/prd_sprints.md` - Sprint C 任务拆解与接口定义
- `doc/openapi/commission.yaml` - 佣金相关 OpenAPI 规范

## 注意事项

1. **金额单位**: 所有金额字段使用"分"为单位存储（整数），避免浮点误差
2. **幂等性**: 佣金创建接口支持幂等，同一订单多次调用不会重复生成
3. **权限控制**: 敏感操作需要添加适当的权限中间件
4. **事务处理**: 涉及多表操作的接口使用数据库事务保证一致性
5. **错误处理**: 所有接口返回统一的错误格式 `{code, message, data}`

## 联系方式

如有问题，请联系开发团队或在 GitHub 提 Issue。
