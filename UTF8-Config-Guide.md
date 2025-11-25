# UTF-8 编码配置说明

## Go 后端编码配置

### 1. 源代码文件编码
- 所有 `.go` 文件必须保存为 UTF-8 编码（无 BOM）
- Go 默认支持 UTF-8 字符串处理

### 2. HTTP 响应编码
- Content-Type: application/json; charset=utf-8
- 确保API响应正确处理中文字符

### 3. 数据库编码
- MySQL charset: utf8mb4
- Collation: utf8mb4_unicode_ci
- 连接参数: charset=utf8mb4

### 4. 配置文件编码
- config.yaml: UTF-8 编码
- 所有配置文件使用 UTF-8

## 前端编码配置

### 1. HTML 文件
- meta charset="UTF-8"
- 文件保存为 UTF-8 编码

### 2. JavaScript 文件
- 文件保存为 UTF-8 编码
- 使用 UTF-8 字符串处理

### 3. CSS 文件
- @charset "UTF-8";
- 文件保存为 UTF-8 编码

## 验证方法

1. 检查文件编码：file --mime-encoding filename
2. 测试中文字符显示
3. API 响应头检查
4. 数据库字符集验证