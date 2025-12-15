# Sprint A/B 回归测试指南

本文档描述如何运行 Sprint A/B 关键接口的回归测试，以确保购物车、下单、可用券列表等核心功能的稳定性。

## 测试范围

### Sprint A - 核心下单能力
- **购物车接口**
  - `GET /api/v1/cart` - 获取用户购物车
  - `POST /api/v1/cart/items` - 加入购物车
  - `PUT /api/v1/cart/items/:id` - 更新购物车商品数量
  - `DELETE /api/v1/cart/items/:id` - 删除购物车商品

- **下单接口**
  - `GET /api/v1/orders` - 获取订单列表
  - `GET /api/v1/orders/:id` - 获取订单详情
  - `POST /api/v1/orders` - 创建订单
  - `GET /api/v1/products` - 获取商品列表（下单前置）
  - `GET /api/v1/stores` - 获取门店列表（门店下单前置）

### Sprint B - 用户与会员体系
- **优惠券接口**
  - `GET /api/v1/coupons` - 获取可用优惠券模板列表
  - `GET /api/v1/user/coupons` - 获取用户优惠券
  - `POST /api/v1/coupons/claim` - 领取优惠券

- **用户接口**
  - `GET /api/v1/user/info` - 获取用户信息
  - `GET /api/v1/user/me` - 获取当前用户
  - `GET /api/v1/users/me/summary` - 获取用户中心聚合数据

## 运行方式

### 方式 1: Go 测试（推荐）

在项目根目录运行：

```bash
cd tea-api
go test -v ./test -run Test_SprintAB_Regression
```

或运行所有测试：

```bash
make test-api
```

### 方式 2: Shell 脚本回归测试

#### 2.1 基础回归测试（仅接口可用性）

```bash
# 确保 API 服务在 9292 端口运行
bash scripts/run_sprint_ab_regression.sh
```

输出示例：
```
==================================================================
Sprint A/B 回归测试 - 关键接口稳定性验证
测试范围：购物车、下单、可用券列表
==================================================================

=== Sprint A: 购物车 (Cart) ===
[购物车] 获取用户购物车
  GET http://127.0.0.1:9292/api/v1/cart
  ✅ 状态码: 200 (成功)
  响应码: 0
  详情: build-ci-logs/sprint_ab_regression/...

...

==================================================================
测试摘要
==================================================================

✅ 成功: 15
⚠️  警告: 2
❌ 失败: 0
```

#### 2.2 集成测试（完整业务流程）

```bash
# 测试完整流程：登录 -> 浏览商品 -> 加入购物车 -> 查看优惠券 -> 查看订单
bash scripts/run_sprint_ab_integration.sh
```

输出示例：
```
==================================================================
Sprint A/B 集成回归测试
测试完整业务流程：购物车 -> 下单 -> 优惠券
==================================================================

步骤 1: 健康检查
----------------------------------------
✅ 健康检查 - 成功 (状态码: 200)

步骤 2: 用户登录 (开发环境)
----------------------------------------
✅ 用户登录 - 成功 (状态码: 200)
  Token: eyJhbGciOiJIUzI1NiIs...

...

==================================================================
测试摘要
==================================================================

通过: 10 个测试
失败: 0 个测试

✅ 所有测试通过！Sprint A/B 关键接口稳定
```

### 方式 3: API 验证脚本

使用现有的 API 验证脚本：

```bash
bash scripts/run_api_validation.sh
```

## 环境变量配置

### 基础配置

```bash
# API 服务地址（默认: http://127.0.0.1:9292）
export BASE=http://127.0.0.1:9292

# Token 文件位置（用于需要认证的接口）
export TOKEN_FILE=build-ci-logs/admin_login_response.json
```

### 开发环境快速启动

```bash
# 1. 启动 API 服务
bash run-tea-api.sh

# 2. 等待服务就绪
sleep 5

# 3. 运行回归测试
bash scripts/run_sprint_ab_regression.sh
```

## 测试结果

### 成功标准

- ✅ 所有关键接口返回 HTTP 200
- ✅ 响应 JSON 格式正确，code=0 表示业务成功
- ✅ 购物车、订单、优惠券等核心功能可正常访问

### 失败处理

如果测试失败：

1. 检查 API 服务是否正常运行：
   ```bash
   curl http://127.0.0.1:9292/api/v1/health
   ```

2. 查看详细日志：
   ```bash
   cat build-ci-logs/sprint_ab_regression/summary.txt
   ```

3. 检查具体接口响应：
   ```bash
   cat build-ci-logs/sprint_ab_regression/*.json
   ```

4. 查看 API 服务日志：
   ```bash
   cat logs/tea-api.log
   ```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Sprint A/B Regression Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  regression:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Start API Server
        run: |
          bash run-tea-api.sh &
          sleep 10
      
      - name: Run Regression Tests
        run: |
          bash scripts/run_sprint_ab_regression.sh
          bash scripts/run_sprint_ab_integration.sh
      
      - name: Run Go Tests
        run: |
          cd tea-api
          go test -v ./test -run Test_SprintAB_Regression
      
      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: regression-test-results
          path: build-ci-logs/
```

## 常见问题

### Q1: 认证失败（401 错误）

**问题**: 某些接口返回 401 Unauthorized

**解决方案**:
1. 确保已生成有效的 token：
   ```bash
   # 开发环境登录
   curl -X POST http://127.0.0.1:9292/api/v1/user/dev-login \
     -H "Content-Type: application/json" \
     -d '{"openid":"test_user"}' \
     > build-ci-logs/admin_login_response.json
   ```

2. 检查 token 文件是否存在且格式正确

### Q2: 资源不存在（404 错误）

**问题**: 某些接口返回 404 Not Found

**解决方案**:
1. 检查测试数据是否已初始化
2. 某些接口需要特定的资源 ID（如订单 ID、商品 ID）
3. 使用实际存在的资源 ID 进行测试

### Q3: 数据库连接失败

**问题**: 测试启动时数据库连接失败

**解决方案**:
1. 确保 MySQL 服务运行正常
2. 检查配置文件 `tea-api/configs/config.yaml`
3. 验证数据库连接参数

## 性能基准

基于 Sprint A/B 的性能要求：

| 接口 | 响应时间 (P95) | TPS |
|------|---------------|-----|
| 获取购物车 | < 200ms | > 100 |
| 获取商品列表 | < 300ms | > 200 |
| 创建订单 | < 500ms | > 50 |
| 获取优惠券列表 | < 200ms | > 100 |

## 扩展测试

如需添加新的回归测试：

1. **Shell 脚本方式**：编辑 `scripts/run_sprint_ab_regression.sh`，添加新的 `run_test` 调用

2. **Go 测试方式**：在 `tea-api/test/sprint_ab_regression_test.go` 添加新的 `t.Run()` 测试用例

3. **API 验证方式**：编辑 `scripts/run_api_validation.sh`，在 `REQUESTS` 数组中添加新接口

## 参考文档

- [PRD Sprint 任务拆解](../doc/prd_sprints.md)
- [API 文档](../tea-api/docs/)
- [进度报告](../tea-api/docs/progress-report.md)
