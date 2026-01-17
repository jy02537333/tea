-- Migration: 2026-01-06 add orders.table_id and orders.table_no for dine-in table tracking
--
-- Notes:
-- - table_id: numeric ID from QR code
-- - table_no: human-friendly table label/number

-- +migrate Up
+ALTER TABLE `orders`
+  ADD COLUMN IF NOT EXISTS `table_id` BIGINT UNSIGNED DEFAULT 0 AFTER `store_id`,
+  ADD COLUMN IF NOT EXISTS `table_no` VARCHAR(50) DEFAULT '' AFTER `table_id`;
+
+-- +migrate Down
+ALTER TABLE `orders`
+  DROP COLUMN IF EXISTS `table_no`,
+  DROP COLUMN IF EXISTS `table_id`;
