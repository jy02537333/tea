# Sprint B 会员购买路径自动化测试

## 概述

本目录包含 Sprint B（会员购买路径）的完整自动化测试与验证体系，用于保护会员购买核心功能不被破坏。

## 文件清单

### 测试脚本

- **`scripts/run_membership_integration.sh`** - 端到端会员购买流程集成测试
- **`scripts/assert_membership_flow.sh`** - 流程结果断言与验证

### 输出文件

测试运行后会在 `build-ci-logs/` 目录生成以下文件：

- **`membership_summary_after_purchase.json`** - 购买流程完整摘要
- **`membership_b_flow_checked.json`** - 流程验证结果

## 快速开始

### 前提条件

1. tea-api 服务正在运行（默认端口 9292）
2. 数据库中至少有一个会员套餐
3. 已安装 `curl` 和 `jq`
4. 已导出有效的 JWT 令牌

### 设置环境变量

```bash
export API_BASE="http://127.0.0.1:9292"
export USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # 可选
```

### 运行测试

#### 方式一：使用 Makefile（推荐）

```bash
# 运行普通模式测试
make verify-sprint-b

# 运行严格模式测试
make verify-sprint-b-strict
```

#### 方式二：直接运行脚本

```bash
# 1. 运行集成测试
./scripts/run_membership_integration.sh

# 2. 验证结果（普通模式）
./scripts/assert_membership_flow.sh

# 或验证结果（严格模式）
./scripts/assert_membership_flow.sh --strict
```

## 验证模式说明

### 普通模式

基础验证：
- ✓ 套餐列表获取成功
- ✓ 订单创建成功
- ✓ 支付订单创建成功
- ✓ 订单出现在订单列表中

### 严格模式（--strict）

包含普通模式的所有检查，外加：
- ✓ 订单状态有效性验证
- ✓ 支付状态有效性验证
- ✓ ID 字段格式验证
- ✓ 摘要文件完整性验证

## 测试流程详解

1. **列出会员套餐** - `GET /api/v1/membership-packages`
2. **创建会员订单** - `POST /api/v1/membership-orders`
3. **创建统一支付订单** - `POST /api/v1/payments/unified-order`
4. **模拟支付回调** - `POST <callback_url>`
5. **列出会员订单** - `GET /api/v1/orders?order_type=4`
6. **生成 JSON 摘要文件**

## CI/CD 集成

在 `.github/workflows/api-validation.yml` 中已集成 Sprint B 测试。

所有测试日志和 JSON 文件都会被归档为 artifacts。

## 故障排查

### 常见错误

#### USER_TOKEN is required
```bash
export USER_TOKEN="your_jwt_token_here"
```

#### No membership package found
需要在数据库中创建会员套餐。

#### jq is required but not installed
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq
```

### 查看详细日志

```bash
cat build-ci-logs/membership-flow-*.log
cat build-ci-logs/membership-assertions-*.log
jq . build-ci-logs/membership_b_flow_checked.json
```

## 相关文档

- [doc/prd.md](../../doc/prd.md) - 产品需求文档
- [doc/prd_sprints.md](../../doc/prd_sprints.md) - Sprint 任务拆解

---

维护者：GitHub Copilot | Sprint B（会员购买路径）自动化保护
