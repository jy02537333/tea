-- Migration: 2025-12-10 add store_products.biz_type for store product business type
-- Up: add the column if it does not exist
-- Down: drop the column

-- 注意：在执行前请先使用 mysqldump 备份相关表或数据库。

ALTER TABLE `store_products`
ADD COLUMN IF NOT EXISTS `biz_type` TINYINT NOT NULL DEFAULT 1 COMMENT '1:服务 2:外卖 3:其他';

-- 回滚（如果需要）:
-- ALTER TABLE `store_products` DROP COLUMN IF EXISTS `biz_type`;
