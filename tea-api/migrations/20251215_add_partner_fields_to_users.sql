-- Sprint C: Add partner and membership fields to users table
-- This migration adds support for partnership levels and membership packages

-- Add membership and partner level fields to users table
ALTER TABLE `users` ADD COLUMN IF NOT EXISTS `membership_package_id` BIGINT UNSIGNED NULL COMMENT '会员礼包ID' AFTER `role`;
ALTER TABLE `users` ADD COLUMN IF NOT EXISTS `partner_level_id` BIGINT UNSIGNED NULL COMMENT '合伙人等级ID' AFTER `membership_package_id`;

-- Add indexes for new fields
ALTER TABLE `users` ADD INDEX `idx_users_membership_package` (`membership_package_id`);
ALTER TABLE `users` ADD INDEX `idx_users_partner_level` (`partner_level_id`);
