# A1 登录（Admin Auth）

**目标**：实现管理员登录页并完成 token 存储、拦截器注入与自动跳转。

**路由**：`#/login`

**关键接口**：`POST /api/v1/user/dev-login`、`POST /api/v1/user/login`

**输入**：用户名/密码 或 openid（dev）

**输出**：保存 `token`（localStorage），跳转到 `#/dashboard`

**验收标准**：
- 成功登录后 `localStorage.token` 存在且 axios 拦截器在后续请求中使用
- 失败时显示错误 Toast；401 自动跳回登录

**估时**：0.5d
