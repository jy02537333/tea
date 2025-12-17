# 迁移说明（给 DBA）

目的
- 将 `tea-api/sql迁移.sql` 应用到目标 MySQL 数据库 `tea_shop`，并保留可恢复的备份。

前提
 在能访问 `127.0.0.1:3308` 的受控主机上运行。建议在运维/DBA 主机执行。
- 确保有足够磁盘空间存放备份文件。
- 建议在执行前将备份文件复制到安全存储（非本机或上云存储）。

步骤（推荐，已封装为脚本 `scripts/migration_for_dba.sh`）

1) 在仓库根目录获取最新代码（可选）

```bash
cd /home/frederic/project/tea
git pull origin master
```

2) 运行脚本（示例：带环境变量更安全）

```bash
export DB_PW='gs963852'
bash scripts/migration_for_dba.sh --host 127.0.0.1 --port 3308 --user root --db tea_shop
```

或直接传密码（不推荐，但示例提供）

```bash
bash scripts/migration_for_dba.sh --host 127.0.0.1 --port 3308 --user root --password 'gs963852' --db tea_shop
```

3) 脚本会执行：
- 使用 `mysqldump` 生成备份到 `build-ci-logs/tea_shop_backup_YYYY-MM-DD_hhmmss.sql`
- 将 `tea-api/sql迁移.sql` 应用到 `tea_shop` 数据库（stdout/stderr 会记录到 `build-ci-logs`）
- 迁移后列出表并保存到 `build-ci-logs/migration_verify_after.txt`

### 直接执行 `db/schema.sql` 的标准步骤（不走脚本时）

当需要直接以仓库内的权威 DDL 文件 `db/schema.sql` 建库/更新结构（例如在本地开发或 CI 初始化阶段），请按以下步骤执行：

1) 备份现有数据库（强烈推荐）：

```bash
# 导出当前库，便于回滚
mysqldump -h 127.0.0.1 -P 3308 -u root -p'gs963852' --routines --events --triggers --databases tea_shop \
	> build-ci-logs/tea_shop_backup_$(date +%F_%H%M%S).sql
```

2) 执行 `db/schema.sql`：

```bash
# 以 schema.sql 为唯一来源执行建表/变更（幂等设计的 DDL 可重复执行）
mysql -h 127.0.0.1 -P 3308 -u root -p'gs963852' tea_shop < db/schema.sql
```

3) 验证关键表是否存在：

```bash
mysql -h 127.0.0.1 -P 3308 -u root -p'gs963852' -e "\
	USE tea_shop; \
	SHOW TABLES LIKE 'users'; \
	SHOW TABLES LIKE 'wallets'; \
	SHOW TABLES LIKE 'points_transactions'; \
	SHOW TABLES LIKE 'coupons'; \
	SHOW TABLES LIKE 'membership_packages'; \
	SHOW TABLES LIKE 'user_memberships'; \
" | tee build-ci-logs/migration_verify_after.txt
```

4) 出现错误时的回滚：

```bash
mysql -h 127.0.0.1 -P 3308 -u root -p'gs963852' < build-ci-logs/tea_shop_backup_YYYY-MM-DD_hhmmss.sql
```

注意：
- `db/schema.sql` 与 `doc/db_schema.md` 需同步维护，任何表结构变更应同时更新两者。
- 若生产环境使用 Flyway/Go migrate，请将 `db/schema.sql` 拆分为版本化迁移脚本并走审核流程。

4) 回滚（如果需要）

```bash
mysql -h 127.0.0.1 -P 3308 -u root -p'gs963852' < build-ci-logs/tea_shop_backup_YYYY-MM-DD_hhmmss.sql
```

注意事项
- 请勿在生产高峰期执行。建议在维护窗口或在 staging 先行演练。
- 密码不要直接放在命令行上（会被 ps 等命令看到），建议使用 `DB_PW` 环境变量或 `~/.my.cnf`。
- 若有疑问，请联系项目负责人并在变更单中记录此次迁移时间、执行者、备份路径与验证输出位置。

日志与验证文件
- 备份文件: `build-ci-logs/tea_shop_backup_*.sql`
- 迁移应用日志: `build-ci-logs/migration_apply_*.log`（stdout/stderr）
- 迁移后验证: `build-ci-logs/migration_verify_after.txt`

如需我也可以代为执行迁移（在你授权下），或将这些文件通过邮件/Slack 发给 DBA。