-- Migration: 2026-01-14 add cart_items.store_id + cart_items.reserved_quantity
-- Purpose:
-- - store_id: keep store context for store orders
-- - reserved_quantity: inventory reservation for cart items

-- +migrate Up
+ALTER TABLE `cart_items`
+  ADD COLUMN IF NOT EXISTS `store_id` BIGINT UNSIGNED NULL AFTER `sku_id`,
+  ADD COLUMN IF NOT EXISTS `reserved_quantity` INT NOT NULL DEFAULT 0 AFTER `quantity`;

-- +migrate Down
+ALTER TABLE `cart_items`
+  DROP COLUMN IF EXISTS `reserved_quantity`,
+  DROP COLUMN IF EXISTS `store_id`;
