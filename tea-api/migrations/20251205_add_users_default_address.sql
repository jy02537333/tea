-- Migration: 2025-12-05 add users.default_address and users.default_address_updated_at
-- Up: add the columns if they do not exist
-- Down: drop the columns

-- 注意：在执行前请先使用 mysqldump 备份相关表或数据库。

ALTER TABLE `users`
ADD COLUMN IF NOT EXISTS `default_address` JSON DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `default_address_updated_at` DATETIME NULL;

-- 回滚（如果需要）:
-- ALTER TABLE `users` DROP COLUMN IF EXISTS `default_address`;
-- ALTER TABLE `users` DROP COLUMN IF EXISTS `default_address_updated_at`;
